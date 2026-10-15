package publicapi

import (
	"bytes"
	"net/http"
	"testing"
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
	registry, err := NewRegistry([]NamedSchema{{Name: "Other", Schema: &Schema{Kind: SchemaObject}}}, []Operation{{
		ID: "conflict", Method: http.MethodGet, Path: "/api/public/v1/conflict", Surface: SurfacePublicV1, Auth: AuthNone, Public: true,
		Responses: []Response{{Kind: ResponseSchema, Status: 500, ContentType: ContentJSON, SchemaRef: "Other"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if bound, err := BindRequestIDFailure(registry); err == nil || bound != nil {
		t.Fatal("binding rewrote a conflicting500 declaration")
	}
	if bound, err := BindRequestIDFailure(nil); err == nil || bound != nil {
		t.Fatal("binding accepted a missing authority")
	}
}
