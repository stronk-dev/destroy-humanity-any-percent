package publicapi

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloud-clicker/server/httpapi"

	"github.com/go-chi/chi/v5"
)

func TestRequestIDFailureBindingPreservesDomainsAndOwnsDeclarations(t *testing.T) {
	schemas := []NamedSchema{APIErrorSchema(), {Name: "OK", Schema: &Schema{Kind: SchemaObject}}}
	oldError := []byte("{\"category\":\"internal_invariant\",\"detail\":\"public_api\"}\n")
	success := Response{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "OK"}
	base, err := NewRegistry(schemas, []Operation{
		{ID: "absent", Method: "GET", Path: "/api/public/v1/absent", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: []Response{success}},
		{ID: "broad", Method: "GET", Path: "/api/public/v1/broad", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: []Response{success, {Kind: ResponseSchema, Status: 500, ContentType: ContentJSON, SchemaRef: "APIError"}}},
		{ID: "exact", Method: "GET", Path: "/api/public/v1/exact", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: []Response{success, {Kind: ResponseSchema, Status: 500, ContentType: ContentJSON, SchemaRef: "APIError", ExactJSON: [][]byte{oldError}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	before, err := CanonicalOperationPins(base)
	if err != nil {
		t.Fatal(err)
	}
	bound, err := BindRequestIDFailure(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"absent", "broad", "exact"} {
		if err := bound.ValidateResponse(id, 500, []byte(requestIDFailureJSON)); err != nil {
			t.Fatalf("%s missing global failure: %v", id, err)
		}
	}
	for _, id := range []string{"broad", "exact"} {
		if err := bound.ValidateResponse(id, 500, oldError); err != nil {
			t.Fatalf("%s lost existing error: %v", id, err)
		}
	}
	// This was already shape-valid in the broad500. Preserve its domain rather
	// than treating this compatibility control as a newly endorsed error pair.
	if err := bound.ValidateResponse("broad", 500, []byte(`{"category":"rate_limited","detail":"ip"}`)); err != nil {
		t.Fatal("binding narrowed the existing broad response", err)
	}
	after, _ := CanonicalOperationPins(base)
	if !bytes.Equal(before, after) {
		t.Fatal("binding mutated the original authority")
	}
	if err := CheckCompatibilityPin(before, bound); err != nil {
		t.Fatal("binding is not additive", err)
	}
	again, err := BindRequestIDFailure(bound)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := CanonicalOperationPins(bound)
	second, _ := CanonicalOperationPins(again)
	if !bytes.Equal(first, second) {
		t.Fatal("binding is not idempotent")
	}
	copy := RequestIDFailureResponse()
	copy.ExactJSON[0][0] = '!'
	if string(RequestIDFailureResponse().ExactJSON[0]) != requestIDFailureJSON {
		t.Fatal("failure declaration aliases caller-owned bytes")
	}
}

func TestRequestIDFailureBindingRefusesConflicting500(t *testing.T) {
	for _, response := range []Response{
		{Kind: ResponseSchema, Status: 500, ContentType: ContentJSON, SchemaRef: "Other"},
		{Kind: ResponseRaw, Status: 500, ContentType: ContentJSON, ContentHashHeader: "X-Content-SHA256"},
	} {
		registry, err := NewRegistry([]NamedSchema{APIErrorSchema(), {Name: "Other", Schema: &Schema{Kind: SchemaObject}}}, []Operation{{
			ID: "conflict", Method: http.MethodGet, Path: "/api/public/v1/conflict", Surface: SurfacePublicV1, Auth: AuthNone, Public: true,
			Responses: []Response{response},
		}})
		if err != nil {
			t.Fatal(err)
		}
		if bound, err := BindRequestIDFailure(registry); err == nil || bound != nil {
			t.Fatal("binding rewrote a conflicting JSON500 declaration")
		}
	}
	if bound, err := BindRequestIDFailure(nil); err == nil || bound != nil {
		t.Fatal("binding accepted a missing authority")
	}
}

func TestRequestIDFailureBindingPreservesGzip500(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write([]byte(`{"existing":"opaque error evidence"}`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data := compressed.Bytes()
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	raw := Response{Kind: ResponseRaw, Status: 500, ContentType: ContentGzip, ContentHashHeader: "X-Content-SHA256"}
	oldError := []byte("{\"category\":\"internal_invariant\",\"detail\":\"public_api\"}\n")
	for _, test := range []struct {
		name string
		json *Response
	}{
		{"raw only", nil},
		{"raw and broad JSON", &Response{Kind: ResponseSchema, Status: 500, ContentType: ContentJSON, SchemaRef: "APIError"}},
		{"raw and exact JSON", &Response{Kind: ResponseSchema, Status: 500, ContentType: ContentJSON, SchemaRef: "APIError", ExactJSON: [][]byte{oldError}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			responses := []Response{raw}
			if test.json != nil {
				responses = append(responses, *test.json)
			}
			base, err := NewRegistry([]NamedSchema{APIErrorSchema()}, []Operation{{
				ID: "get_evidence", Method: http.MethodGet, Path: "/api/public/v1/evidence", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: responses,
			}})
			if err != nil {
				t.Fatal(err)
			}
			before, err := CanonicalOperationPins(base)
			if err != nil {
				t.Fatal(err)
			}
			bound, err := BindRequestIDFailure(base)
			if err != nil {
				t.Fatalf("valid existing gzip response blocked binding: %v", err)
			}
			if err := bound.ValidateRawResponse("get_evidence", 500, ContentGzip, data, digest); err != nil {
				t.Fatal("lost existing raw response", err)
			}
			if err := bound.ValidateResponse("get_evidence", 500, []byte(requestIDFailureJSON)); err != nil {
				t.Fatal("missing identity error", err)
			}
			if test.json != nil {
				if err := bound.ValidateResponse("get_evidence", 500, oldError); err != nil {
					t.Fatal("lost existing JSON response", err)
				}
			}
			operation, _ := bound.Operation("get_evidence")
			if len(operation.Responses) != 2 {
				t.Fatal("raw declaration changed or response duplicated")
			}
			retained := operation.Responses[0]
			if retained.Kind != raw.Kind || retained.Status != raw.Status || retained.ContentType != raw.ContentType || retained.ContentHashHeader != raw.ContentHashHeader || retained.SchemaRef != "" || len(retained.ExactJSON) != 0 {
				t.Fatal("raw declaration changed")
			}
			after, err := CanonicalOperationPins(base)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("binding mutated input authority", err)
			}
			if err := CheckCompatibilityPin(before, bound); err != nil {
				t.Fatal("binding broke existing responses", err)
			}
			again, err := BindRequestIDFailure(bound)
			if err != nil {
				t.Fatal(err)
			}
			first, err := CanonicalOperationPins(bound)
			if err != nil {
				t.Fatal(err)
			}
			second, err := CanonicalOperationPins(again)
			if err != nil || !bytes.Equal(first, second) {
				t.Fatal("binding is not idempotent", err)
			}

			// Exercise both actual mounted representations. A valid caller ID reaches
			// the old raw handler; exhausted entropy produces only the JSON failure.
			called := 0
			router := chi.NewRouter()
			if err := bound.Mount(router, []Binding{{OperationID: "get_evidence", Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called++
				w.Header().Set("Content-Type", ContentGzip)
				w.Header().Set("X-Content-SHA256", digest)
				w.WriteHeader(500)
				_, _ = w.Write(data)
			})}}, nil); err != nil {
				t.Fatal(err)
			}
			ids, err := httpapi.NewRequestIDs(`^[A-Za-z0-9-]{1,64}$`, 64, bytes.NewReader(nil))
			if err != nil {
				t.Fatal(err)
			}
			handler := WithRequestID(router, ids, func() time.Time { return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC) })
			request := httptest.NewRequest(http.MethodGet, operation.Path, nil)
			request.Header.Set("X-Request-ID", "existing-raw")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != 500 || response.Header().Get("Content-Type") != ContentGzip || response.Header().Get("X-Content-SHA256") != digest || response.Header().Get("X-Request-ID") != "existing-raw" || !bytes.Equal(response.Body.Bytes(), data) || called != 1 {
				t.Fatal("existing mounted raw reply changed")
			}
			response = httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, operation.Path, nil))
			if response.Code != 500 || response.Header().Get("Content-Type") != ContentJSON || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Content-SHA256") != "" || response.Header().Get("X-Request-ID") != "" || response.Body.String() != requestIDFailureJSON || called != 1 {
				t.Fatal("identity failure did not retain its JSON boundary")
			}

			document, err := GenerateOpenAPI(bound, "Mixed error representations")
			if err != nil {
				t.Fatal(err)
			}
			var parsed struct {
				Paths map[string]map[string]struct {
					Responses map[string]struct{ Content map[string]json.RawMessage }
				}
			}
			if err := json.Unmarshal(document, &parsed); err != nil {
				t.Fatal(err)
			}
			content := parsed.Paths[operation.Path]["get"].Responses["500"].Content
			if len(content) != 2 || content[ContentGzip] == nil || content[ContentJSON] == nil {
				t.Fatal("OpenAPI lost a500 representation")
			}
			types, err := GenerateTypeScript(bound)
			if err != nil {
				t.Fatal(err)
			}
			for _, declaration := range []string{`status: 500, kind: "raw", contentType: "application/gzip", contentHashHeader: "X-Content-SHA256"`, `status: 500, kind: "schema", contentType: "application/json"`, `get_evidence: { status: 500; body: Uint8Array } | { status: 500; body: APIError };`} {
				if !bytes.Contains(types, []byte(declaration)) {
					t.Fatalf("generated client lost %q", declaration)
				}
			}
		})
	}
}
