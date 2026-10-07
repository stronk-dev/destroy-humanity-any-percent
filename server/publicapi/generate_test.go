package publicapi

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestGenerateOpenAPIAndTypeScriptFromImmutableRegistry(t *testing.T) {
	registry := testRegistry(t)
	first, err := GenerateOpenAPI(registry, "Test API")
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateOpenAPI(registry, "Test API")
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("OpenAPI generation is not byte-deterministic")
	}
	var document map[string]any
	if json.Unmarshal(first, &document) != nil || document["openapi"] != "3.1.0" {
		t.Fatalf("invalid OpenAPI: %s", first)
	}
	text := string(first)
	for _, required := range []string{
		`"additionalProperties": false`, `"operationId": "get_board"`,
		`"name": "category"`, `"$ref": "#/components/schemas/EpochPage"`,
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("OpenAPI omitted %s", required)
		}
	}
	types, err := GenerateTypeScript(registry)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"export type EpochPage =", "next_cursor: null | string", "export const operations =",
		`get_board: { auth: "none", method: "GET"`, "export interface OperationTypes",
		`export const apiErrorCategories = ["invalid", "unknown_id"] as const;`,
	} {
		if !bytes.Contains(types, []byte(required)) {
			t.Fatalf("TypeScript omitted %q\n%s", required, types)
		}
	}
	pins, err := CanonicalOperationPins(registry)
	if err != nil || !bytes.Contains(pins, []byte(`"schemas"`)) || !bytes.Contains(pins, []byte(`"get_board"`)) {
		t.Fatalf("invalid compatibility pins: %v\n%s", err, pins)
	}
}

func TestGeneratedClientPreservesRawSuccessAndStatusDescriptors(t *testing.T) {
	operations := testOperations()
	operations[1].Responses = []Response{
		{Kind: ResponseRaw, Status: 200, ContentType: ContentGzip, ContentHashHeader: "X-Content-SHA256"},
		{Kind: ResponseSchema, Status: 404, ContentType: ContentJSON, SchemaRef: "APIError"},
	}
	registry, err := NewRegistry(testSchemas(), operations)
	if err != nil {
		t.Fatal(err)
	}
	types, err := GenerateTypeScript(registry)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		`get_epochs: { path: {  }; request: null; response: Uint8Array | APIError };`,
		`get_epochs: { status: 200; body: Uint8Array } | { status: 404; body: APIError };`,
		`{ status: 200, kind: "raw", contentType: "application/gzip", contentHashHeader: "X-Content-SHA256" }`,
		`{ status: 404, kind: "schema", contentType: "application/json", contentHashHeader: "" }`,
		`export function createAPIClient(`,
	} {
		if !bytes.Contains(types, []byte(required)) {
			t.Errorf("generated client omitted %q", required)
		}
	}
}
