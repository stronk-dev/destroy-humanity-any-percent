package publicapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func pathParameterDefinitions(target *Schema) []NamedSchema {
	return []NamedSchema{
		{Name: "Reply", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
			{Name: "value", Schema: &Schema{Kind: SchemaString}, Required: true},
		}}},
		{Name: "Target", Schema: target},
		{Name: "TargetAlias", Schema: &Schema{Kind: SchemaRef, Ref: "Target"}},
		{Name: "TargetAliasAgain", Schema: &Schema{Kind: SchemaRef, Ref: "TargetAlias"}},
	}
}

func pathParameterOperation(schema *Schema) Operation {
	return Operation{ID: "get_example", Method: http.MethodGet, Path: "/api/public/v1/example/{value}",
		Surface: SurfacePublicV1, Public: true, Auth: AuthNone,
		Parameters: []Parameter{{Name: "value", Schema: schema}},
		Responses:  []Response{{Kind: ResponseSchema, Status: http.StatusOK, ContentType: ContentJSON, SchemaRef: "Reply"}},
	}
}

// A reference must not bypass the same scalar rule as the inline descriptor.
// These are construction regressions, not new runtime URI-value constraints.
func TestRegistryPathParametersRejectReferencedNonScalars(t *testing.T) {
	for name, target := range map[string]*Schema{
		"object": {Kind: SchemaObject, Fields: []Field{{Name: "value", Required: true, Schema: &Schema{Kind: SchemaString}}}},
		"array":  {Kind: SchemaArray, Items: &Schema{Kind: SchemaString}},
		"null":   {Kind: SchemaNull},
		"union":  {Kind: SchemaOneOf, Alternates: []*Schema{{Kind: SchemaString}, {Kind: SchemaNull}}},
	} {
		for _, ref := range []string{"Target", "TargetAliasAgain"} {
			t.Run(name+"/"+ref, func(t *testing.T) {
				registry, err := NewRegistry(pathParameterDefinitions(target), []Operation{pathParameterOperation(&Schema{Kind: SchemaRef, Ref: ref})})
				if !errors.Is(err, ErrInvalidOperation) || registry != nil {
					t.Fatalf("non-scalar path reference admitted: registry=%t err=%v", registry != nil, err)
				}
			})
		}
	}
}

func TestRegistryPathParametersPreserveScalarReferencesAndDispatch(t *testing.T) {
	for _, row := range []struct {
		name, value string
		schema      *Schema
	}{
		{"string", "item.alpha", &Schema{Kind: SchemaString, Format: "mechanical-id"}},
		{"integer", "7", &Schema{Kind: SchemaInteger, Minimum: integerPointer(0), Maximum: integerPointer(10)}},
		{"boolean", "true", &Schema{Kind: SchemaBoolean}},
	} {
		for _, ref := range []string{"Target", "TargetAliasAgain"} {
			t.Run(row.name+"/"+ref, func(t *testing.T) {
				operation := pathParameterOperation(&Schema{Kind: SchemaRef, Ref: ref})
				registry, err := NewRegistry(pathParameterDefinitions(row.schema), []Operation{operation})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := GenerateOpenAPI(registry, "Path parameter test"); err != nil {
					t.Fatal(err)
				}
				generated, err := GenerateTypeScript(registry)
				if err != nil || !strings.Contains(string(generated), "value: "+ref) {
					t.Fatalf("named path type was lost: %v", err)
				}
				router := chi.NewRouter()
				calls := 0
				if err := registry.Mount(router, []Binding{{OperationID: operation.ID, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					w.Header().Set("Content-Type", ContentJSON)
					fmt.Fprintf(w, `{"value":%q}`, chi.URLParam(r, "value"))
				})}}, nil); err != nil {
					t.Fatal(err)
				}
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/public/v1/example/"+row.value, nil))
				if calls != 1 || recorder.Code != http.StatusOK || recorder.Body.String() != fmt.Sprintf(`{"value":%q}`, row.value) {
					t.Fatalf("scalar dispatch changed: calls=%d status=%d body=%s", calls, recorder.Code, recorder.Body.String())
				}
				if err := registry.ValidateResponse(operation.ID, recorder.Code, recorder.Body.Bytes()); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestRegistryPathParametersRejectRepeatedNames(t *testing.T) {
	for _, path := range []string{"/api/public/v1/example/{value}/{value}", "/api/public/v1/example/{value}/nested/{value}"} {
		t.Run(path, func(t *testing.T) {
			operation := pathParameterOperation(&Schema{Kind: SchemaString})
			operation.Path = path
			operation.Parameters = append(operation.Parameters, Parameter{Name: "value", Schema: &Schema{Kind: SchemaString}})
			registry, err := NewRegistry(pathParameterDefinitions(&Schema{Kind: SchemaString}), []Operation{operation})
			if !errors.Is(err, ErrInvalidOperation) || registry != nil {
				t.Fatalf("repeated path name admitted: registry=%t err=%v", registry != nil, err)
			}
		})
	}
}

func TestRegistryPathParametersPreserveDistinctOrderedNames(t *testing.T) {
	// Path order, not lexical sorting. Equal schemas do not alias distinct names.
	operation := pathParameterOperation(&Schema{Kind: SchemaString})
	operation.Path = "/api/public/v1/example/{z}/{a}"
	operation.Parameters = []Parameter{{Name: "z", Schema: &Schema{Kind: SchemaString}}, {Name: "a", Schema: &Schema{Kind: SchemaString}}}
	registry, err := NewRegistry(pathParameterDefinitions(&Schema{Kind: SchemaString}), []Operation{operation})
	if err != nil {
		t.Fatal(err)
	}
	router := chi.NewRouter()
	if err := registry.Mount(router, []Binding{{OperationID: operation.ID, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentJSON)
		fmt.Fprintf(w, `{"value":%q}`, chi.URLParam(r, "z")+","+chi.URLParam(r, "a"))
	})}}, nil); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/public/v1/example/first/second", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"value":"first,second"}` {
		t.Fatal("distinct parameters lost their path ordering")
	}
	if err := registry.ValidateResponse(operation.ID, recorder.Code, recorder.Body.Bytes()); err != nil {
		t.Fatal(err)
	}
}
