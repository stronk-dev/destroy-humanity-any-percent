package publicapi

import "testing"

func TestMergeRegistriesUnionsIdenticalSchemasAndRejectsConflicts(t *testing.T) {
	left := testRegistry(t)
	schemas := testSchemas()[:1]
	status := Operation{ID: "get_status", Method: "GET", Path: "/api/public/v1/status", Surface: SurfacePublicV1, Auth: AuthNone, Public: true,
		Responses: []Response{{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "APIError"}}}
	right, err := NewRegistry(schemas, []Operation{status})
	if err != nil {
		t.Fatal(err)
	}
	merged, err := MergeRegistries(left, right)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(merged.Operations()); got != 3 {
		t.Fatalf("merged operations = %d", got)
	}
	if len(merged.Schemas()) != len(testSchemas()) {
		t.Fatalf("identical shared schema was duplicated: %d", len(merged.Schemas()))
	}

	conflicting := testSchemas()[:1]
	conflicting[0].Schema.Fields = conflicting[0].Schema.Fields[:1]
	other, err := NewRegistry(conflicting, []Operation{status})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := MergeRegistries(left, other); err == nil {
		t.Fatal("conflicting shared schema accepted")
	}
	if _, err := MergeRegistries(left, left); err == nil {
		t.Fatal("duplicate operation IDs accepted")
	}
}
