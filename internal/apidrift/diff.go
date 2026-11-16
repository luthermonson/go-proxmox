package apidrift

import (
	"reflect"
	"sort"
	"strings"
)

// Mapping ties one schema endpoint (method + path, exactly as apidoc.js spells
// them, e.g. "GET /nodes/{node}/disks/list") to the Go type the library
// decodes that response into. For array responses, Type is the element type.
type Mapping struct {
	Method string
	Path   string
	Type   reflect.Type
}

// EndpointDrift is the diff result for one mapped endpoint.
type EndpointDrift struct {
	Method string
	Path   string
	GoType string
	// Missing lists schema return properties with no matching json tag on the
	// Go type (a response field the library silently drops).
	Missing []string
	// Untyped lists schema properties that declare a nested object but whose
	// Go field decodes into a map/interface{} (structure the library throws
	// away).
	Untyped []string
}

// HasDrift reports whether this endpoint has any finding.
func (d EndpointDrift) HasDrift() bool { return len(d.Missing) > 0 || len(d.Untyped) > 0 }

// Report is the full drift result, deterministically ordered.
type Report struct {
	// Drifted holds mapped endpoints that have at least one finding.
	Drifted []EndpointDrift
	// Clean lists mapped endpoints with no findings (for the summary).
	Clean []string
	// Unmapped lists schema endpoints that declare a structured return but are
	// absent from the registry — candidates to wire up, or unimplemented.
	Unmapped []string
	// StaleMappings lists registry entries whose path+method no longer exist
	// in the schema (PVE removed/renamed the endpoint).
	StaleMappings []string
	// SchemaEndpoints / StructuredEndpoints / MappedEndpoints are counts for
	// the summary line.
	SchemaEndpoints     int
	StructuredEndpoints int
	MappedEndpoints     int
}

// HasDrift reports whether the report contains anything a CI gate should fail
// on: a mapped endpoint with missing/untyped fields, or a stale registry
// mapping. Unmapped endpoints are informational (the long tail) and do not
// fail the gate on their own.
func (r Report) HasDrift() bool {
	return len(r.Drifted) > 0 || len(r.StaleMappings) > 0
}

// Diff compares the schema's declared return properties against the Go structs
// in the registry.
func Diff(endpoints []Endpoint, registry []Mapping) Report {
	schemaByKey := make(map[string]Endpoint, len(endpoints))
	structured := 0
	for _, e := range endpoints {
		schemaByKey[e.Key()] = e
		if e.Returns.Structured() {
			structured++
		}
	}

	mapped := make(map[string]struct{}, len(registry))
	var rep Report
	rep.SchemaEndpoints = len(endpoints)
	rep.StructuredEndpoints = structured

	for _, m := range registry {
		k := key(m.Method, m.Path)
		mapped[k] = struct{}{}
		ep, ok := schemaByKey[k]
		if !ok {
			rep.StaleMappings = append(rep.StaleMappings, k)
			continue
		}
		rep.MappedEndpoints++
		fields := collectJSONFields(m.Type)
		drift := EndpointDrift{Method: strings.ToUpper(m.Method), Path: m.Path, GoType: typeName(m.Type)}
		for name, prop := range ep.Returns.Properties {
			// encoding/json resolves keys to struct fields case-insensitively
			// (after an exact match), so mirror that to avoid false "missing"
			// reports for fields like CPU decoding the schema's "cpu".
			f, present := fields[strings.ToLower(name)]
			if !present {
				drift.Missing = append(drift.Missing, name)
				continue
			}
			if prop.HasProperties() && isUntyped(f) {
				drift.Untyped = append(drift.Untyped, name)
			}
		}
		sort.Strings(drift.Missing)
		sort.Strings(drift.Untyped)
		if drift.HasDrift() {
			rep.Drifted = append(rep.Drifted, drift)
		} else {
			rep.Clean = append(rep.Clean, k)
		}
	}

	for _, e := range endpoints {
		if !e.Returns.Structured() {
			continue
		}
		if _, ok := mapped[e.Key()]; ok {
			continue
		}
		rep.Unmapped = append(rep.Unmapped, e.Key())
	}

	sort.Slice(rep.Drifted, func(i, j int) bool {
		if rep.Drifted[i].Path != rep.Drifted[j].Path {
			return rep.Drifted[i].Path < rep.Drifted[j].Path
		}
		return rep.Drifted[i].Method < rep.Drifted[j].Method
	})
	sort.Strings(rep.Clean)
	sort.Strings(rep.Unmapped)
	sort.Strings(rep.StaleMappings)
	return rep
}

// collectJSONFields walks a struct type (dereferencing pointers/slices and
// following embedded/anonymous structs) and returns a map of lowercased json
// key → field type. The tag's name segment is used, or the field name when the
// tag has no name (encoding/json then matches case-insensitively, so keys are
// lowercased here). Fields tagged `json:"-"` and unexported untagged fields are
// skipped.
func collectJSONFields(t reflect.Type) map[string]reflect.Type {
	out := map[string]reflect.Type{}
	var visit func(t reflect.Type)
	visit = func(t reflect.Type) {
		t = deref(t)
		if t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("json")
			name := strings.Split(tag, ",")[0]
			if name == "-" {
				continue
			}
			if f.Anonymous && name == "" {
				// Embedded struct: promote its fields.
				visit(f.Type)
				continue
			}
			if name == "" {
				if f.PkgPath != "" {
					continue // unexported, untagged
				}
				name = f.Name
			}
			lname := strings.ToLower(name)
			if _, exists := out[lname]; !exists {
				out[lname] = f.Type
			}
		}
	}
	visit(t)
	return out
}

// isUntyped reports whether a Go field type throws away declared structure —
// i.e. it decodes an object into a map or an empty interface rather than a
// named struct.
func isUntyped(t reflect.Type) bool {
	t = deref(t)
	switch t.Kind() {
	case reflect.Map:
		// map[string]any / map[string]interface{} is untyped; map[string]SomeStruct is fine.
		return isUntyped(t.Elem())
	case reflect.Interface:
		return t.NumMethod() == 0
	default:
		return false
	}
}

func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}
	return t
}

func typeName(t reflect.Type) string {
	t = deref(t)
	if t.PkgPath() == "" {
		return t.String()
	}
	return t.Name()
}
