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

// Renaming a placeholder does not create a different HTTP route. Chi stores
// only one endpoint for these identical patterns; accepting both loses an
// operation (and can replace its authentication middleware) during mounting.
func TestRegistryRejectsDuplicateParameterizedRoutes(t *testing.T) {
	for _, paths := range [][2]string{
		{"/api/v1/example/{first}", "/api/v1/example/{second}"},
		{"/api/v1/example/{first}/nested/{second}", "/api/v1/example/{second}/nested/{first}"},
		{"/api/v1/example/{first}/nested/{last}", "/api/v1/example/{second}/nested/{last}"},
	} {
		t.Run(paths[0]+" versus "+paths[1], func(t *testing.T) {
			operations := routeIdentityOperations(paths[0], paths[1])
			registry, err := NewRegistry(routeIdentitySchemas(), operations)
			if errors.Is(err, ErrInvalidOperation) && registry == nil {
				return
			}
			if err != nil || registry == nil {
				t.Fatalf("unexpected construction result: registry=%t err=%v", registry != nil, err)
			}
			// Keep the actual dispatch diagnostic for old-source review overlays.
			// This is not a production route or a claim of a deployed auth bypass.
			router := routeIdentityRouter(t, registry)
			requestPath := "/api/v1/example/item"
			if len(operations[0].Parameters) == 2 {
				requestPath += "/nested/tail"
			}
			reply := httptest.NewRecorder()
			router.ServeHTTP(reply, httptest.NewRequest(http.MethodGet, requestPath, nil))
			t.Fatalf("duplicate parameterized route admitted: unauthenticated status=%d body=%s (first operation requires authentication)", reply.Code, reply.Body.String())
		})
	}
}

func TestRegistryDistinctRoutesPreserveDispatchAndAuthentication(t *testing.T) {
	for _, row := range []struct {
		name, first, second, firstURL, secondURL, secondMethod string
		publicSecond                                           bool
	}{
		{"different methods", "/api/v1/example/{first}", "/api/v1/example/{second}", "/api/v1/example/item", "/api/v1/example/next", http.MethodPost, false},
		{"different static suffixes", "/api/v1/example/{first}/one", "/api/v1/example/{second}/two", "/api/v1/example/item/one", "/api/v1/example/next/two", http.MethodGet, false},
		{"different depths", "/api/v1/example/{first}", "/api/v1/example/{second}/nested/{last}", "/api/v1/example/item", "/api/v1/example/next/nested/tail", http.MethodGet, false},
		{"different static prefixes", "/api/v1/example/{first}", "/api/v1/other/{second}", "/api/v1/example/item", "/api/v1/other/next", http.MethodGet, false},
		{"different surfaces", "/api/v1/example/{first}", "/api/public/v1/example/{second}", "/api/v1/example/item", "/api/public/v1/example/next", http.MethodGet, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			operations := routeIdentityOperations(row.first, row.second)
			operations[1].Method = row.secondMethod
			if row.publicSecond {
				operations[1].Surface, operations[1].Public = SurfacePublicV1, true
			}
			registry, err := NewRegistry(routeIdentitySchemas(), operations)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := GenerateOpenAPI(registry, "Route identity test"); err != nil {
				t.Fatal(err)
			}
			if _, err := GenerateTypeScript(registry); err != nil {
				t.Fatal(err)
			}
			router := routeIdentityRouter(t, registry)
			first := httptest.NewRecorder()
			router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, row.firstURL, nil))
			if first.Code != http.StatusUnauthorized {
				t.Fatalf("distinct route bypassed first operation authentication: %d %s", first.Code, first.Body.String())
			}
			for _, request := range []struct {
				method, path, authorization, body, operation string
			}{
				{http.MethodGet, row.firstURL, "Bearer fixture", `{"value":"get_first:item"}`, "get_first"},
				{row.secondMethod, row.secondURL, "", `{"value":"get_second:next"}`, "get_second"},
			} {
				if row.name == "different depths" && request.operation == "get_second" {
					request.body = `{"value":"get_second:next:tail"}`
				}
				reply := httptest.NewRecorder()
				r := httptest.NewRequest(request.method, request.path, nil)
				r.Header.Set("Authorization", request.authorization)
				router.ServeHTTP(reply, r)
				if reply.Code != http.StatusOK || reply.Body.String() != request.body {
					t.Fatalf("distinct dispatch changed: %s %s got%d %s want%s", request.method, request.path, reply.Code, reply.Body.String(), request.body)
				}
				if err := registry.ValidateResponse(request.operation, reply.Code, reply.Body.Bytes()); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestMergeRegistriesRejectsDuplicateParameterizedRoutes(t *testing.T) {
	operations := routeIdentityOperations("/api/v1/example/{first}", "/api/v1/example/{second}")
	first, err := NewRegistry(routeIdentitySchemas(), operations[:1])
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewRegistry(routeIdentitySchemas(), operations[1:])
	if err != nil {
		t.Fatal(err)
	}
	for _, sources := range [][]*Registry{{first, second}, {second, first}} {
		merged, err := MergeRegistries(sources...)
		if !errors.Is(err, ErrInvalidOperation) || merged != nil {
			t.Fatalf("merged generation authority admitted colliding routes: registry=%t err=%v", merged != nil, err)
		}
	}
}

func routeIdentitySchemas() []NamedSchema {
	return []NamedSchema{{Name: "Reply", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
		{Name: "value", Required: true, Schema: &Schema{Kind: SchemaString}},
	}}}}
}

func routeIdentityOperations(firstPath, secondPath string) []Operation {
	operations := []Operation{}
	for index, path := range []string{firstPath, secondPath} {
		operation := Operation{ID: []string{"get_first", "get_second"}[index], Method: http.MethodGet,
			Path: path, Surface: SurfacePrivateV1, Auth: AuthNone,
			Responses: []Response{{Kind: ResponseSchema, Status: http.StatusOK, ContentType: ContentJSON, SchemaRef: "Reply"}}}
		if index == 0 {
			operation.Auth = AuthAccessToken
		}
		// Use the same already-validated template grammar as ordinary registry rows.
		for _, segment := range strings.Split(path, "/") {
			if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
				operation.Parameters = append(operation.Parameters, Parameter{Name: segment[1 : len(segment)-1], Schema: &Schema{Kind: SchemaString}})
			}
		}
		operations = append(operations, operation)
	}
	return operations
}

func routeIdentityRouter(t *testing.T, registry *Registry) chi.Router {
	t.Helper()
	router := chi.NewRouter()
	bindings := []Binding{}
	for _, operation := range registry.Operations() {
		bindings = append(bindings, Binding{OperationID: operation.ID, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", ContentJSON)
			value := operation.ID
			for _, parameter := range operation.Parameters {
				value += ":" + chi.URLParam(r, parameter.Name)
			}
			fmt.Fprintf(w, `{"value":%q}`, value)
		})})
	}
	if err := registry.Mount(router, bindings, map[AuthMode]Middleware{AuthAccessToken: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer fixture" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}}); err != nil {
		t.Fatal(err)
	}
	return router
}
