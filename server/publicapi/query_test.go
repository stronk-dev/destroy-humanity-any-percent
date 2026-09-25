package publicapi

import (
	"bytes"
	"errors"
	"net/url"
	"testing"
)

func pagedQuery() []QueryParameter {
	return []QueryParameter{
		{Name: "cursor", Schema: &Schema{Kind: SchemaString}},
		{Name: "limit", Schema: &Schema{Kind: SchemaInteger, Minimum: integerPointer(1), Maximum: integerPointer(100)}},
	}
}

func queryOperations() []Operation {
	operations := testOperations()
	operations[0].Query = pagedQuery()
	return operations
}

// plainOperations removes every query/cursor declaration, modelling a
// registry that predates query parameters.
func plainOperations() []Operation {
	operations := testOperations()
	for index := range operations {
		operations[index].Query, operations[index].CursorKey = nil, ""
	}
	return operations
}

func plainRegistry(t *testing.T) *Registry {
	t.Helper()
	registry, err := NewRegistry(testSchemas(), plainOperations())
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func TestRegistryRequiresExactQueryParameterDescriptors(t *testing.T) {
	if _, err := NewRegistry(testSchemas(), queryOperations()); err != nil {
		t.Fatal(err)
	}
	for name, query := range map[string][]QueryParameter{
		"unsorted":                            {pagedQuery()[1], pagedQuery()[0]},
		"duplicate":                           {pagedQuery()[0], pagedQuery()[0]},
		"invalid name":                        {{Name: "Cursor", Schema: &Schema{Kind: SchemaString}}, pagedQuery()[1]},
		"object schema":                       {pagedQuery()[0], {Name: "limit", Schema: &Schema{Kind: SchemaObject}}},
		"nil schema":                          {pagedQuery()[0], {Name: "limit"}},
		"path name collision":                 {{Name: "category", Schema: &Schema{Kind: SchemaString}}, pagedQuery()[0], pagedQuery()[1]},
		"cursor key without cursor parameter": {pagedQuery()[1]},
		"required cursor":                     {{Name: "cursor", Schema: &Schema{Kind: SchemaString}, Required: true}, pagedQuery()[1]},
	} {
		t.Run(name, func(t *testing.T) {
			operations := queryOperations()
			operations[0].Query = query
			if _, err := NewRegistry(testSchemas(), operations); err == nil {
				t.Fatal("invalid query parameters accepted")
			}
		})
	}
	// A cursor parameter without a cursor key has nothing to decode against.
	operations := queryOperations()
	operations[1].Query = []QueryParameter{{Name: "cursor", Schema: &Schema{Kind: SchemaString}}}
	if _, err := NewRegistry(testSchemas(), operations); err == nil {
		t.Fatal("cursor parameter without cursor key accepted")
	}
}

func TestRegistryQuerySnapshotsAreImmutable(t *testing.T) {
	operations := queryOperations()
	registry, err := NewRegistry(testSchemas(), operations)
	if err != nil {
		t.Fatal(err)
	}
	operations[0].Query[1].Name = "mutated"
	snapshot, _ := registry.Operation("get_board")
	snapshot.Query[0].Name = "mutated"
	again, _ := registry.Operation("get_board")
	if again.Query[0].Name != "cursor" || again.Query[1].Name != "limit" {
		t.Fatalf("query descriptors are aliased: %+v", again.Query)
	}
}

func TestParseQueryValidatesDeclaredParametersExactly(t *testing.T) {
	registry, err := NewRegistry(testSchemas(), queryOperations())
	if err != nil {
		t.Fatal(err)
	}
	values, err := registry.ParseQuery("get_board", url.Values{"limit": {"25"}, "cursor": {"abc"}})
	if err != nil || values["limit"] != int64(25) || values["cursor"] != "abc" {
		t.Fatalf("valid query rejected: %v %+v", err, values)
	}
	if values, err := registry.ParseQuery("get_board", url.Values{}); err != nil || len(values) != 0 {
		t.Fatalf("absent optional query rejected: %v %+v", err, values)
	}
	for _, query := range []url.Values{
		{"limit": {"0"}}, {"limit": {"101"}}, {"limit": {"010"}}, {"limit": {"+5"}}, {"limit": {"5x"}}, {"limit": {""}},
		{"limit": {"5", "6"}}, {"cursor": {""}},
	} {
		_, err := registry.ParseQuery("get_board", query)
		var invalid *InvalidQueryError
		if !errors.As(err, &invalid) {
			t.Fatalf("invalid query %v accepted or untyped: %v", query, err)
		}
	}
	_, err = registry.ParseQuery("get_board", url.Values{"limit": {"0"}})
	var invalid *InvalidQueryError
	if !errors.As(err, &invalid) || invalid.Parameter != "limit" {
		t.Fatalf("invalid query does not name its parameter: %v", err)
	}
	if _, err := registry.ParseQuery("unknown", url.Values{}); err == nil {
		t.Fatal("unknown operation query accepted")
	}
}

func TestGeneratedArtifactsCarryQueryParameters(t *testing.T) {
	registry, err := NewRegistry(testSchemas(), queryOperations())
	if err != nil {
		t.Fatal(err)
	}
	openapi, err := GenerateOpenAPI(registry, "test")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range [][]byte{[]byte(`"in": "query"`), []byte(`"name": "limit"`), []byte(`"maximum": 100`)} {
		if !bytes.Contains(openapi, want) {
			t.Fatalf("OpenAPI lacks %s", want)
		}
	}
	types, err := GenerateTypeScript(registry)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`queryParameters: ["cursor", "limit"]`, `query: { cursor?: string; limit?: number }`} {
		if !bytes.Contains(types, []byte(want)) {
			t.Fatalf("TypeScript lacks %s:\n%s", want, types)
		}
	}
	// Operations without query parameters keep their prior generated shape.
	plain, err := GenerateTypeScript(plainRegistry(t))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(plain, []byte("query")) {
		t.Fatalf("query-free registry grew query output:\n%s", plain)
	}
}

func TestCompatibilityPinRejectsQueryChanges(t *testing.T) {
	prior, err := NewRegistry(testSchemas(), queryOperations())
	if err != nil {
		t.Fatal(err)
	}
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckCompatibilityPin(pin, prior); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func([]Operation){
		"removal":   func(operations []Operation) { operations[0].Query = operations[0].Query[:1] },
		"narrowing": func(operations []Operation) { operations[0].Query[1].Schema.Maximum = integerPointer(50) },
		"required":  func(operations []Operation) { operations[0].Query[1].Required = true },
	} {
		t.Run(name, func(t *testing.T) {
			operations := queryOperations()
			mutate(operations)
			next, err := NewRegistry(testSchemas(), operations)
			if err != nil {
				t.Fatal(err)
			}
			if err := CheckCompatibilityPin(pin, next); err == nil {
				t.Fatal("query change accepted")
			}
		})
	}
	// The pre-query pin shape (no query key) stays byte-identical for query-free operations.
	plain, err := CanonicalOperationPins(plainRegistry(t))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(plain, []byte(`"query"`)) {
		t.Fatalf("query-free pin grew a query key:\n%s", plain)
	}
}

func TestQueryIntegersAreCanonicalBaseTen(t *testing.T) {
	schema := &Schema{Kind: SchemaInteger, Minimum: integerPointer(-5), Maximum: integerPointer(5)}
	for _, raw := range []string{"-0", "00", "-01"} {
		if _, err := decodeQueryValue(schema, raw); err == nil {
			t.Fatalf("non-canonical integer %q accepted", raw)
		}
	}
	for raw, want := range map[string]int64{"0": 0, "-3": -3, "5": 5} {
		if value, err := decodeQueryValue(schema, raw); err != nil || value != want {
			t.Fatalf("canonical integer %q = %v, %v", raw, value, err)
		}
	}
}
