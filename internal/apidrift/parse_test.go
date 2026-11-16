package apidrift

import "testing"

// fixtureJS mimics the apidoc.js wrapper: `const <name> = [ ... ];` with the
// handler-code terminator the real file has, plus a couple of nested nodes.
const fixtureJS = `const apiSchema = [
   {
      "path" : "/widgets",
      "text" : "widgets",
      "info" : {
         "GET" : {
            "returns" : {
               "type" : "array",
               "items" : {
                  "type" : "object",
                  "properties" : {
                     "id" : { "type" : "string" },
                     "size" : { "type" : "integer" },
                     "nested" : {
                        "type" : "object",
                        "properties" : { "inner" : { "type" : "string" } }
                     }
                  }
               }
            }
         }
      },
      "children" : [
         {
            "path" : "/widgets/{id}",
            "text" : "{id}",
            "info" : {
               "GET" : {
                  "returns" : {
                     "type" : "object",
                     "properties" : {
                        "id" : { "type" : "string" },
                        "label" : { "type" : "string" }
                     }
                  }
               },
               "DELETE" : { "returns" : { "type" : "null" } }
            }
         }
      ]
   },
   {
      "path" : "/ping",
      "text" : "ping",
      "info" : {
         "GET" : { "returns" : { "type" : "boolean" } }
      }
   },
   {
      "path" : "/blank",
      "text" : "blank",
      "info" : {
         "GET" : { "returns" : { "type" : "object", "properties" : {} } }
      }
   }
]
;
function handler() { return m2c[method]; }
`

func TestParseAPIDoc(t *testing.T) {
	endpoints, err := ParseAPIDoc([]byte(fixtureJS))
	if err != nil {
		t.Fatalf("ParseAPIDoc: %v", err)
	}

	byKey := map[string]Endpoint{}
	for _, e := range endpoints {
		byKey[e.Key()] = e
	}

	// Array-of-objects: Properties come from items.properties.
	list := byKey["get /widgets"]
	if list.Returns.Kind != KindArray {
		t.Fatalf("GET /widgets: kind = %q, want array", list.Returns.Kind)
	}
	if !list.Returns.Structured() {
		t.Fatalf("GET /widgets should be structured")
	}
	for _, want := range []string{"id", "size", "nested"} {
		if _, ok := list.Returns.Properties[want]; !ok {
			t.Errorf("GET /widgets missing property %q", want)
		}
	}
	if !list.Returns.Properties["nested"].HasProperties() {
		t.Errorf("nested should declare sub-properties")
	}
	if list.Returns.Properties["id"].HasProperties() {
		t.Errorf("id is a scalar and should not declare sub-properties")
	}

	// Object return.
	obj := byKey["get /widgets/{id}"]
	if obj.Returns.Kind != KindObject {
		t.Fatalf("GET /widgets/{id}: kind = %q, want object", obj.Returns.Kind)
	}

	// Scalar return.
	if k := byKey["get /ping"].Returns.Kind; k != KindScalar {
		t.Errorf("GET /ping: kind = %q, want scalar", k)
	}

	// Propertyless object → empty, not structured.
	blank := byKey["get /blank"]
	if blank.Returns.Kind != KindEmpty {
		t.Errorf("GET /blank: kind = %q, want empty", blank.Returns.Kind)
	}
	if blank.Returns.Structured() {
		t.Errorf("GET /blank should not count as structured")
	}

	// null return → none.
	if k := byKey["delete /widgets/{id}"].Returns.Kind; k != KindNone {
		t.Errorf("DELETE /widgets/{id}: kind = %q, want none", k)
	}
}

func TestParsePlainJSONArray(t *testing.T) {
	// extractSchemaJSON should also accept a bare JSON array (cached fixture).
	raw := `[{"path":"/x","info":{"GET":{"returns":{"type":"object","properties":{"a":{"type":"string"}}}}}}]`
	endpoints, err := ParseAPIDoc([]byte(raw))
	if err != nil {
		t.Fatalf("ParseAPIDoc(plain array): %v", err)
	}
	if len(endpoints) != 1 || endpoints[0].Key() != "get /x" {
		t.Fatalf("unexpected parse result: %+v", endpoints)
	}
}
