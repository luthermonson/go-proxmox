// Package apidrift parses the upstream Proxmox VE API schema (apidoc.js) and
// diffs the declared *return* types of each endpoint against the Go structs
// this library decodes those responses into. It powers the `apidrift` command
// and the weekly type-drift GitHub Actions workflow.
//
// Scope and honesty: the library does not annotate which Go type decodes which
// endpoint, so a fully-automatic path→type mapping is not possible. apidrift
// instead diffs a curated registry (see registry.go) of endpoint→type
// mappings and reports every schema endpoint that declares a structured return
// but is absent from that registry as "unmapped" (candidate for wiring up or a
// genuinely unimplemented endpoint). The registry is a seed, not the complete
// surface — that is by design, and the weekly workflow grows it over time.
package apidrift

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// Property is one declared field of a schema return object.
type Property struct {
	// Name is the JSON key as PVE declares it.
	Name string
	// Type is the JSON schema type ("string", "integer", "boolean", "number",
	// "object", "array", "null") or "" when PVE declares none.
	Type string
	// Properties holds the declared sub-fields of an object (or of an array's
	// object items). It is nil when the property declares no structure — PVE
	// publishing no sub-structure is a meaningful signal, not an omission.
	Properties map[string]Property
}

// HasProperties reports whether the property declares a structured object (or
// array-of-objects) with named sub-fields. A Go field that decodes such a
// property into a map or interface{} is flagged as UNTYPED.
func (p Property) HasProperties() bool { return len(p.Properties) > 0 }

// ReturnKind classifies an endpoint's declared return shape.
type ReturnKind string

const (
	// KindObject is a JSON object with (possibly zero) declared properties.
	KindObject ReturnKind = "object"
	// KindArray is a JSON array; for arrays of objects, Properties holds the
	// item object's fields.
	KindArray ReturnKind = "array"
	// KindScalar is a bare scalar (string/integer/boolean/number/null).
	KindScalar ReturnKind = "scalar"
	// KindEmpty is an object/array that declares no properties at all — PVE
	// says "there is a response here but its shape is unspecified".
	KindEmpty ReturnKind = "empty"
	// KindNone means the method declares no `returns` block.
	KindNone ReturnKind = "none"
)

// Return is the parsed `returns` schema of a single endpoint method.
type Return struct {
	Kind       ReturnKind
	ScalarType string // set when Kind == KindScalar
	Properties map[string]Property
}

// Structured reports whether the return declares named properties worth
// diffing against a Go struct.
func (r Return) Structured() bool {
	return (r.Kind == KindObject || r.Kind == KindArray) && len(r.Properties) > 0
}

// Endpoint is one (method, path) pair with its parsed return schema.
type Endpoint struct {
	Method  string
	Path    string
	Returns Return
}

// Key is the registry/diff key: lowercase method + space + path.
func (e Endpoint) Key() string { return key(e.Method, e.Path) }

func key(method, path string) string {
	return toLower(method) + " " + path
}

func toLower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

// rawNode mirrors the apidoc.js tree. Each node carries a `path`, an `info`
// map keyed by HTTP method, and `children`.
type rawNode struct {
	Path     string                     `json:"path"`
	Text     string                     `json:"text"`
	Info     map[string]json.RawMessage `json:"info"`
	Children []rawNode                  `json:"children"`
}

type rawMethod struct {
	Returns json.RawMessage `json:"returns"`
}

// ParseAPIDoc unwraps the `const <name> = [ ... ];` JS array out of apidoc.js
// and parses it into a flat, deterministically-sorted list of endpoints.
func ParseAPIDoc(raw []byte) ([]Endpoint, error) {
	blob, err := extractSchemaJSON(raw)
	if err != nil {
		return nil, err
	}
	var tree []rawNode
	if err := json.Unmarshal(blob, &tree); err != nil {
		return nil, fmt.Errorf("apidrift: parse apidoc.js schema: %w", err)
	}

	var out []Endpoint
	var walk func(n rawNode)
	walk = func(n rawNode) {
		if n.Path != "" {
			for method, rawInfo := range n.Info {
				var m rawMethod
				if err := json.Unmarshal(rawInfo, &m); err != nil {
					continue
				}
				out = append(out, Endpoint{
					Method:  method,
					Path:    n.Path,
					Returns: parseReturn(m.Returns),
				})
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, t := range tree {
		walk(t)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Method < out[j].Method
	})
	return out, nil
}

// rawSchema is the minimal JSON-schema shape apidoc.js uses for `returns`.
type rawSchema struct {
	Type       string               `json:"type"`
	Properties map[string]rawSchema `json:"properties"`
	Items      *rawSchema           `json:"items"`
}

func parseReturn(raw json.RawMessage) Return {
	if len(bytes.TrimSpace(raw)) == 0 || string(bytes.TrimSpace(raw)) == "null" {
		return Return{Kind: KindNone}
	}
	var s rawSchema
	if err := json.Unmarshal(raw, &s); err != nil {
		return Return{Kind: KindNone}
	}

	switch {
	case s.Type == "null":
		// A declared null return carries no data — treat it like no return.
		return Return{Kind: KindNone}
	case s.Type == "array" || s.Items != nil:
		var props map[string]Property
		if s.Items != nil {
			props = parseProps(s.Items.Properties)
		}
		if len(props) == 0 {
			return Return{Kind: KindEmpty}
		}
		return Return{Kind: KindArray, Properties: props}
	case s.Type == "object" || len(s.Properties) > 0:
		props := parseProps(s.Properties)
		if len(props) == 0 {
			return Return{Kind: KindEmpty}
		}
		return Return{Kind: KindObject, Properties: props}
	case s.Type != "":
		return Return{Kind: KindScalar, ScalarType: s.Type}
	default:
		return Return{Kind: KindEmpty}
	}
}

func parseProps(in map[string]rawSchema) map[string]Property {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]Property, len(in))
	for name, s := range in {
		p := Property{Name: name, Type: s.Type}
		switch {
		case s.Type == "array" && s.Items != nil:
			p.Properties = parseProps(s.Items.Properties)
		case len(s.Properties) > 0:
			p.Properties = parseProps(s.Properties)
		}
		out[name] = p
	}
	return out
}

// extractSchemaJSON peels the `const <name> = [ ... ];` array out of the
// upstream JS file. We anchor the end on `\n]\n;` rather than the last `]`
// because the handler code below the schema contains template literals whose
// `]` would otherwise be matched first. This mirrors mage/endpoints.
func extractSchemaJSON(raw []byte) ([]byte, error) {
	start := bytes.IndexByte(raw, '[')
	end := bytes.Index(raw, []byte("\n]\n;"))
	if start < 0 || end <= start {
		// Fall back to a plain JSON array (used by tests and cached fixtures).
		trimmed := bytes.TrimSpace(raw)
		if len(trimmed) > 0 && trimmed[0] == '[' {
			return trimmed, nil
		}
		return nil, fmt.Errorf("apidrift: could not locate JSON array boundaries in apidoc.js")
	}
	return raw[start : end+2], nil
}
