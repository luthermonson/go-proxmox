// Audit tool for issues #199 + #178: cross-checks the go-proxmox config/options
// structs against the PVE API schema. Flags two classes of bug:
//
//  1. #199 — Go zero value differs from the PVE-documented default, so a
//     marshalled struct silently overwrites the server-side default.
//  2. #178 — Go field is plain int/uint/bool but the PVE schema declares
//     "type": "boolean", so the field type is wrong on the wire.
//
// Run from the audit/ directory:
//
//	# fetch the PVE schema (3.7 MB, gitignored — regenerate when PVE updates)
//	curl -sS https://pve.proxmox.com/pve-docs/api-viewer/apidoc.js \
//	  | awk 'NR==1{sub(/^const apiSchema = /,"")} NR<=65015{print}' \
//	  > apidoc.json
//	# regenerate the report
//	go run . > report.md
//
// The line-bound `awk` is brittle — if PVE restructures apidoc.js the line
// number boundary will change. Adjust by finding the line with bare `]` that
// closes the top-level array.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

// PVE schema (apidoc.json) is a tree of endpoints. Each leaf node has an
// `info` map keyed by HTTP method, whose `parameters` field describes the
// request params. Defaults live on individual properties.
type schemaNode struct {
	Path     string                 `json:"path"`
	Text     string                 `json:"text"`
	Info     map[string]methodInfo  `json:"info"`
	Children []schemaNode           `json:"children"`
}

type methodInfo struct {
	Method     string          `json:"method"`
	Name       string          `json:"name"`
	Parameters paramSchema     `json:"parameters"`
	Returns    json.RawMessage `json:"returns"`
}

type paramSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]propSchema  `json:"properties"`
}

// propSchema captures only the bits of the PVE param schema we need.
// We use json.RawMessage for `default` because PVE uses mixed scalar types
// (int / string / bool) and we just want to know whether one is present
// and what its raw value is.
type propSchema struct {
	Type        json.RawMessage `json:"type"`
	Description string          `json:"description"`
	Default     json.RawMessage `json:"default"`
	Optional    json.RawMessage `json:"optional"`
	Format      json.RawMessage `json:"format"`
	Enum        json.RawMessage `json:"enum"`
}

// endpointSelector picks which PVE endpoint(s) define defaults for a given
// Go struct. We prefer the create endpoints (POST) because they document
// every param's default; PUT/config endpoints often omit defaults assuming
// you read them on POST.
type endpointSelector struct {
	StructName string
	Endpoints  []endpointKey // ordered: first match wins per param
}

type endpointKey struct {
	Path   string
	Method string
}

// targets lists the structs we want to audit and the PVE endpoints that own
// their defaults. Add more as we discover them.
var targets = []endpointSelector{
	{
		StructName: "ContainerConfig",
		Endpoints: []endpointKey{
			{"/nodes/{node}/lxc", "POST"},
			{"/nodes/{node}/lxc/{vmid}/config", "PUT"},
		},
	},
	{
		StructName: "VirtualMachineConfig",
		Endpoints: []endpointKey{
			{"/nodes/{node}/qemu", "POST"},
			{"/nodes/{node}/qemu/{vmid}/config", "PUT"},
		},
	},
	{
		StructName: "VzdumpConfig",
		Endpoints: []endpointKey{
			{"/nodes/{node}/vzdump", "POST"},
			{"/cluster/backup", "POST"},
		},
	},
	{
		StructName: "VirtualMachineCloneOptions",
		Endpoints:  []endpointKey{{"/nodes/{node}/qemu/{vmid}/clone", "POST"}},
	},
	{
		StructName: "ContainerCloneOptions",
		Endpoints:  []endpointKey{{"/nodes/{node}/lxc/{vmid}/clone", "POST"}},
	},
	{
		StructName: "VirtualMachineMigrateOptions",
		Endpoints:  []endpointKey{{"/nodes/{node}/qemu/{vmid}/migrate", "POST"}},
	},
	{
		StructName: "ContainerMigrateOptions",
		Endpoints:  []endpointKey{{"/nodes/{node}/lxc/{vmid}/migrate", "POST"}},
	},
	{
		StructName: "VirtualMachineMoveDiskOptions",
		Endpoints:  []endpointKey{{"/nodes/{node}/qemu/{vmid}/move_disk", "POST"}},
	},
	{
		StructName: "VirtualMachineBackupOptions",
		Endpoints:  []endpointKey{{"/nodes/{node}/vzdump", "POST"}},
	},
	{
		StructName: "ClusterStorageOptions",
		Endpoints:  []endpointKey{{"/storage", "POST"}, {"/storage/{storage}", "PUT"}},
	},
	{
		StructName: "VNCProxyOptions",
		Endpoints: []endpointKey{
			{"/nodes/{node}/qemu/{vmid}/vncproxy", "POST"},
			{"/nodes/{node}/lxc/{vmid}/vncproxy", "POST"},
			{"/nodes/{node}/vncshell", "POST"},
		},
	},
	{
		StructName: "StorageDownloadURLOptions",
		Endpoints:  []endpointKey{{"/nodes/{node}/storage/{storage}/download-url", "POST"}},
	},
	{
		StructName: "FirewallNodeOption",
		Endpoints:  []endpointKey{{"/nodes/{node}/firewall/options", "PUT"}},
	},
	{
		StructName: "FirewallVirtualMachineOption",
		Endpoints: []endpointKey{
			{"/nodes/{node}/qemu/{vmid}/firewall/options", "PUT"},
			{"/nodes/{node}/lxc/{vmid}/firewall/options", "PUT"},
		},
	},
	{
		StructName: "VNetOptions",
		Endpoints:  []endpointKey{{"/cluster/sdn/vnets", "POST"}, {"/cluster/sdn/vnets/{vnet}", "PUT"}},
	},
	{
		StructName: "SDNZoneOptions",
		Endpoints:  []endpointKey{{"/cluster/sdn/zones", "POST"}, {"/cluster/sdn/zones/{zone}", "PUT"}},
	},
	{
		StructName: "PoolUpdateOption",
		Endpoints:  []endpointKey{{"/pools/{poolid}", "PUT"}},
	},
	{
		StructName: "DomainSyncOptions",
		Endpoints:  []endpointKey{{"/access/domains/{realm}/sync", "POST"}},
	},
	{
		StructName: "UserOptions",
		Endpoints:  []endpointKey{{"/access/users", "POST"}, {"/access/users/{userid}", "PUT"}},
	},
	{
		StructName: "PermissionsOptions",
		Endpoints:  []endpointKey{{"/access/permissions", "GET"}},
	},
}

// flattenSchema walks the schema tree and indexes endpoints by (path, method).
func flattenSchema(nodes []schemaNode, out map[endpointKey]methodInfo) {
	for _, n := range nodes {
		for method, info := range n.Info {
			out[endpointKey{Path: n.Path, Method: method}] = info
		}
		flattenSchema(n.Children, out)
	}
}

// goField describes one field on a struct found in types.go.
type goField struct {
	StructName string
	Name       string
	GoType     string
	JSONTag    string
	OmitEmpty  bool
	Line       int
}

// parseStructs walks types.go and returns all fields on the named structs.
func parseStructs(path string, want map[string]bool) ([]goField, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var fields []goField
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if !want[ts.Name.Name] {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, f := range st.Fields.List {
				goType := exprString(f.Type)
				tag := ""
				omit := false
				if f.Tag != nil {
					raw := strings.Trim(f.Tag.Value, "`")
					tag, omit = jsonTag(raw)
				}
				if tag == "-" {
					continue
				}
				for _, n := range f.Names {
					fields = append(fields, goField{
						StructName: ts.Name.Name,
						Name:       n.Name,
						GoType:     goType,
						JSONTag:    tag,
						OmitEmpty:  omit,
						Line:       fset.Position(n.Pos()).Line,
					})
				}
			}
		}
	}
	return fields, nil
}

func exprString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprString(t.X)
	case *ast.ArrayType:
		return "[]" + exprString(t.Elt)
	case *ast.MapType:
		return "map[" + exprString(t.Key) + "]" + exprString(t.Value)
	case *ast.SelectorExpr:
		return exprString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return fmt.Sprintf("%T", e)
	}
}

func jsonTag(rawTag string) (string, bool) {
	for _, kv := range strings.Split(rawTag, " ") {
		if !strings.HasPrefix(kv, "json:") {
			continue
		}
		val := strings.Trim(strings.TrimPrefix(kv, "json:"), `"`)
		parts := strings.Split(val, ",")
		omit := false
		for _, p := range parts[1:] {
			if p == "omitempty" {
				omit = true
			}
		}
		return parts[0], omit
	}
	return "", false
}

// schemaTypes returns the PVE schema types declared for a property. PVE uses
// either a bare string ("boolean") or a list (["boolean", "string"]); we
// normalize to a slice of lowercase strings.
func schemaTypes(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{strings.ToLower(single)}
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		out := make([]string, 0, len(list))
		for _, t := range list {
			out = append(out, strings.ToLower(t))
		}
		return out
	}
	return nil
}

func schemaIsBool(p propSchema) bool {
	for _, t := range schemaTypes(p.Type) {
		if t == "boolean" {
			return true
		}
	}
	return false
}

// goIsIntOrBool reports whether the field's Go type is already IntOrBool or
// a pointer to it (treated equivalent for the type-mismatch rule from #178).
func goIsIntOrBool(goType string) bool {
	return goType == "IntOrBool" || goType == "*IntOrBool"
}

// goIsIntFamily catches int/uint variants that need to be retyped to IntOrBool
// when the schema declares "boolean".
func goIsIntFamily(goType string) bool {
	switch strings.TrimPrefix(goType, "*") {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"bool":
		return true
	}
	return false
}

// goZeroForType returns the Go zero value as a JSON-comparable raw value, or
// the empty string if the type isn't a simple scalar we know how to map.
func goZeroForType(goType string) string {
	switch goType {
	case "string":
		return `""`
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return "0"
	case "bool":
		return "false"
	case "IntOrBool":
		return "0" // marshals as 0/1; zero-value bool → 0
	case "StringOrInt", "StringOrFloat64", "StringOrUint64":
		return `""`
	}
	if strings.HasPrefix(goType, "*") {
		return "nil"
	}
	return ""
}

// compareDefault returns true if the PVE default differs from the Go zero value.
// Both are compared as their JSON string representation (after light normalization).
func compareDefault(pveDefault json.RawMessage, goZero string) (differs bool, pveStr string) {
	if len(pveDefault) == 0 {
		return false, ""
	}
	pveStr = strings.TrimSpace(string(pveDefault))
	// Strings come quoted; unify "true"/"false" / "1"/"0" representations
	// for IntOrBool so we don't false-flag matches.
	switch goZero {
	case "false":
		if pveStr == "0" || pveStr == `"0"` {
			return false, pveStr
		}
	case "0":
		if pveStr == "false" || pveStr == `"0"` {
			return false, pveStr
		}
	case `""`:
		if pveStr == `""` {
			return false, pveStr
		}
	}
	if pveStr == goZero {
		return false, pveStr
	}
	return true, pveStr
}

func main() {
	raw, err := os.ReadFile("apidoc.json")
	if err != nil {
		fail(err)
	}
	var schema []schemaNode
	if err := json.Unmarshal(raw, &schema); err != nil {
		fail(err)
	}
	endpoints := make(map[endpointKey]methodInfo)
	flattenSchema(schema, endpoints)

	want := make(map[string]bool, len(targets))
	for _, t := range targets {
		want[t.StructName] = true
	}
	fields, err := parseStructs("../types.go", want)
	if err != nil {
		fail(err)
	}
	byStruct := make(map[string][]goField)
	for _, f := range fields {
		byStruct[f.StructName] = append(byStruct[f.StructName], f)
	}

	// latchWriter wraps os.Stdout so we only have to check one error at the
	// end of report generation. CLAUDE.md forbids `_ =`-style discards on
	// fmt.Fprint*, but threading an error return through every print call
	// would obscure the output structure.
	out := &latchWriter{w: os.Stdout}
	out.writeln("# PVE-defaults audit (issues #199 + #178)")
	out.writeln("")
	out.writeln("Generated by `audit/main.go`. Two rules combined:")
	out.writeln("1. **Defaults rule (#199)**: Go zero value differs from the documented PVE default → send the wrong value when unset.")
	out.writeln("2. **Type rule (#178)**: PVE schema declares `\"type\": \"boolean\"` but the Go field is plain `int`/`uint`/`bool` instead of `IntOrBool` → wrong type on the wire.")
	out.writeln("")
	out.writeln("Recommendation matrix: schema=bool + defaults match → `IntOrBool`; schema=bool + defaults differ → `*IntOrBool`; otherwise → `*T` only when defaults differ.")
	out.writeln("")
	out.writeln("> **Note on `FirewallNodeOption.Enable` vs `FirewallVirtualMachineOption.Enable`**: per-node firewall defaults to `1` (ships enabled), per-VM defaults to `0` (opt-in). This asymmetry is intentional in PVE's three-gate firewall design (cluster + node + VM) and is *not* a schema bug.")
	out.writeln("")

	// Diagnostics: which endpoints did we resolve?
	missingEndpoints := []string{}
	for _, t := range targets {
		hit := false
		for _, ep := range t.Endpoints {
			if _, ok := endpoints[ep]; ok {
				hit = true
				break
			}
		}
		if !hit {
			missingEndpoints = append(missingEndpoints, t.StructName)
		}
	}
	if len(missingEndpoints) > 0 {
		out.writeln("## ⚠ Endpoints not found in schema (audit incomplete for these)")
		for _, n := range missingEndpoints {
			out.writef("- `%s`\n", n)
		}
		out.writeln("")
	}

	totalFlagged := 0
	for _, t := range targets {
		flds := byStruct[t.StructName]
		if len(flds) == 0 {
			continue
		}
		// Merge param dictionaries from all configured endpoints.
		params := make(map[string]propSchema)
		usedEndpoint := ""
		for _, ep := range t.Endpoints {
			info, ok := endpoints[ep]
			if !ok {
				continue
			}
			if usedEndpoint == "" {
				usedEndpoint = ep.Method + " " + ep.Path
			}
			for k, v := range info.Parameters.Properties {
				if _, exists := params[k]; !exists {
					params[k] = v
				}
			}
		}
		flagged := []row{}
		untracked := []goField{}
		for _, f := range flds {
			if f.JSONTag == "" {
				continue
			}
			p, ok := params[f.JSONTag]
			if !ok {
				// Indexed fields like net0..9: PVE schema uses pattern keys
				// (e.g. "net[n]"). Try mapping numeric suffix → "[n]".
				if mapped, found := tryIndexedLookup(f.JSONTag, params); found {
					p = mapped
					ok = true
				}
			}
			if !ok {
				untracked = append(untracked, f)
				continue
			}
			goZero := goZeroForType(f.GoType)
			isBool := schemaIsBool(p)
			typeMismatch := isBool && !goIsIntOrBool(f.GoType) && goIsIntFamily(f.GoType)

			if goZero == "" {
				// non-scalar field — can't auto-judge; show it if PVE has a default
				// or if there's a type mismatch we still want to flag.
				if len(p.Default) > 0 || typeMismatch {
					flagged = append(flagged, row{f, p, "?", string(p.Default), "non-scalar; manual review"})
				}
				continue
			}
			differs, pveStr := compareDefault(p.Default, goZero)
			if differs || typeMismatch {
				flagged = append(flagged, row{
					f:         f,
					p:         p,
					goZero:    goZero,
					pveStr:    pveStr,
					recommend: combinedRecommendation(f.GoType, isBool, differs),
				})
			}
		}

		out.writef("## `%s`\n", t.StructName)
		if usedEndpoint != "" {
			out.writef("Schema source: `%s`\n\n", usedEndpoint)
		}
		if len(flagged) == 0 {
			out.writeln("_No mismatches found._")
		} else {
			out.writeln("| Field | json tag | Go type | PVE type | Go zero | PVE default | Recommended fix |")
			out.writeln("|-------|----------|---------|----------|---------|-------------|-----------------|")
			sort.Slice(flagged, func(i, j int) bool { return flagged[i].f.Line < flagged[j].f.Line })
			for _, r := range flagged {
				ts := schemaTypes(r.p.Type)
				pveType := strings.Join(ts, "|")
				if pveType == "" {
					pveType = "?"
				}
				out.writef("| `%s` (L%d) | `%s` | `%s` | `%s` | `%s` | `%s` | %s |\n",
					r.f.Name, r.f.Line, r.f.JSONTag, r.f.GoType, pveType, r.goZero, r.pveStr, r.recommend)
			}
			totalFlagged += len(flagged)
		}
		if len(untracked) > 0 {
			out.writef("\n<details><summary>%d field(s) not found in schema (likely read-only or our naming differs)</summary>\n\n", len(untracked))
			for _, u := range untracked {
				out.writef("- `%s` (json: `%s`, type: `%s`)\n", u.Name, u.JSONTag, u.GoType)
			}
			out.writeln("\n</details>")
		}
		out.writeln("")
	}
	out.writef("---\n**Total fields needing a fix: %d**\n", totalFlagged)
	if out.err != nil {
		fail(out.err)
	}
}

// latchWriter is a tiny io.Writer-ish helper that captures the first error
// from a sequence of fmt.Fprint* calls so the caller checks once at the end.
type latchWriter struct {
	w   io.Writer
	err error
}

func (lw *latchWriter) writeln(s string) {
	if lw.err == nil {
		_, lw.err = fmt.Fprintln(lw.w, s)
	}
}

func (lw *latchWriter) writef(format string, args ...any) {
	if lw.err == nil {
		_, lw.err = fmt.Fprintf(lw.w, format, args...)
	}
}

type row struct {
	f         goField
	p         propSchema
	goZero    string
	pveStr    string
	recommend string
}

// combinedRecommendation produces the fix recommendation per the matrix from
// issue-178 + issue-199:
//
//	schema=bool, defaults match  → IntOrBool   (type fix only)
//	schema=bool, defaults differ → *IntOrBool  (type + pointer)
//	schema=other, defaults match → no change
//	schema=other, defaults differ → *T         (pointer only)
func combinedRecommendation(goType string, schemaIsBool, defaultsDiffer bool) string {
	if schemaIsBool {
		if defaultsDiffer {
			if goType == "*IntOrBool" {
				return "already `*IntOrBool` — verify usage"
			}
			return "use `*IntOrBool` (schema=boolean + default differs)"
		}
		if goType == "IntOrBool" || goType == "*IntOrBool" {
			return "already IntOrBool-typed"
		}
		return "use `IntOrBool` (schema=boolean; defaults match so pointer not required)"
	}
	// non-boolean schema — defaults-only fix
	if !defaultsDiffer {
		return "no change needed"
	}
	if strings.HasPrefix(goType, "*") {
		return "already a pointer — verify usage"
	}
	switch goType {
	case "string":
		return "use `*string` (or sentinel) — PVE default is non-empty"
	default:
		return "use `*" + goType + "`"
	}
}

// tryIndexedLookup maps fields like "net0".."net9" → schema key "net[n]"
// (PVE uses bracketed pattern keys for repeated params).
func tryIndexedLookup(tag string, params map[string]propSchema) (propSchema, bool) {
	for i := len(tag) - 1; i >= 0; i-- {
		if tag[i] < '0' || tag[i] > '9' {
			if i == len(tag)-1 {
				return propSchema{}, false
			}
			prefix := tag[:i+1]
			if p, ok := params[prefix+"[n]"]; ok {
				return p, true
			}
			return propSchema{}, false
		}
	}
	return propSchema{}, false
}

func fail(err error) {
	// log.Fatalf writes to stderr and exits; using it instead of
	// fmt.Fprintln + os.Exit keeps the unchecked-error linter quiet.
	log.SetFlags(0)
	log.Fatalf("audit: %v", err)
}
