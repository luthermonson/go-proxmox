package apidrift

import (
	"reflect"
	"strings"
	"testing"
)

// Fixture Go types for the diff, including an embedded struct to exercise
// json-tag promotion.

type fixtureEmbedded struct {
	Shared string `json:"shared,omitempty"`
}

type fixtureWidget struct {
	fixtureEmbedded
	ID   string `json:"id"`
	Size int    `json:"size,omitempty"`
	// "nested" is declared as an object by the schema but decoded into a map
	// here → should be flagged UNTYPED. "missingField" is absent entirely.
	Nested map[string]any `json:"nested,omitempty"`
	Hidden string         `json:"-"`
}

func fixtureEndpoints() []Endpoint {
	return []Endpoint{
		{
			Method: "GET", Path: "/widgets",
			Returns: Return{
				Kind: KindArray,
				Properties: map[string]Property{
					"id":           {Name: "id", Type: "string"},
					"size":         {Name: "size", Type: "integer"},
					"shared":       {Name: "shared", Type: "string"},
					"missingField": {Name: "missingField", Type: "string"},
					"nested": {Name: "nested", Type: "object", Properties: map[string]Property{
						"inner": {Name: "inner", Type: "string"},
					}},
				},
			},
		},
		{
			Method: "GET", Path: "/unmapped",
			Returns: Return{
				Kind:       KindObject,
				Properties: map[string]Property{"x": {Name: "x", Type: "string"}},
			},
		},
		{
			Method: "GET", Path: "/scalar",
			Returns: Return{Kind: KindScalar, ScalarType: "boolean"},
		},
	}
}

func TestDiffFindsMissingUntypedUnmappedAndStale(t *testing.T) {
	registry := []Mapping{
		{Method: "GET", Path: "/widgets", Type: reflect.TypeOf(fixtureWidget{})},
		{Method: "GET", Path: "/gone", Type: reflect.TypeOf(fixtureWidget{})}, // not in schema → stale
	}

	rep := Diff(fixtureEndpoints(), registry)

	if len(rep.Drifted) != 1 {
		t.Fatalf("want 1 drifted endpoint, got %d: %+v", len(rep.Drifted), rep.Drifted)
	}
	d := rep.Drifted[0]
	if got := strings.Join(d.Missing, ","); got != "missingField" {
		t.Errorf("Missing = %q, want \"missingField\"", got)
	}
	if got := strings.Join(d.Untyped, ","); got != "nested" {
		t.Errorf("Untyped = %q, want \"nested\"", got)
	}

	// "shared" is promoted from the embedded struct, so it must NOT be missing.
	for _, m := range d.Missing {
		if m == "shared" {
			t.Errorf("embedded field 'shared' was wrongly reported missing")
		}
	}

	// /unmapped has a structured return and no mapping → unmapped.
	if !contains(rep.Unmapped, "get /unmapped") {
		t.Errorf("want /unmapped in Unmapped, got %v", rep.Unmapped)
	}
	// /scalar is not structured → must not be reported unmapped.
	if contains(rep.Unmapped, "get /scalar") {
		t.Errorf("/scalar is scalar and should not be unmapped")
	}
	// /gone is a stale mapping.
	if !contains(rep.StaleMappings, "get /gone") {
		t.Errorf("want /gone in StaleMappings, got %v", rep.StaleMappings)
	}

	if !rep.HasDrift() {
		t.Errorf("report should report drift")
	}
}

func TestDiffCleanWhenStructMatches(t *testing.T) {
	type exact struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	endpoints := []Endpoint{{
		Method: "GET", Path: "/x",
		Returns: Return{Kind: KindObject, Properties: map[string]Property{
			"a": {Name: "a", Type: "string"},
			"b": {Name: "b", Type: "string"},
		}},
	}}
	rep := Diff(endpoints, []Mapping{{Method: "GET", Path: "/x", Type: reflect.TypeOf(exact{})}})
	if rep.HasDrift() {
		t.Fatalf("expected no drift, got %+v", rep.Drifted)
	}
	if len(rep.Clean) != 1 || rep.Clean[0] != "get /x" {
		t.Errorf("Clean = %v, want [get /x]", rep.Clean)
	}
}

func TestTypedMapIsNotUntyped(t *testing.T) {
	type inner struct {
		V string `json:"v"`
	}
	type outer struct {
		M map[string]inner `json:"m"`
	}
	endpoints := []Endpoint{{
		Method: "GET", Path: "/x",
		Returns: Return{Kind: KindObject, Properties: map[string]Property{
			"m": {Name: "m", Type: "object", Properties: map[string]Property{"v": {Name: "v", Type: "string"}}},
		}},
	}}
	rep := Diff(endpoints, []Mapping{{Method: "GET", Path: "/x", Type: reflect.TypeOf(outer{})}})
	if rep.HasDrift() {
		t.Errorf("map[string]struct should not be flagged untyped: %+v", rep.Drifted)
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
