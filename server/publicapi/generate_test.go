package publicapi

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestOpenAPIResponseHeadersByRepresentation(t *testing.T) {
	type header struct {
		Required    bool              `json:"required"`
		Description string            `json:"description"`
		Schema      map[string]string `json:"schema"`
	}
	raw := func(media, name string) Response {
		return Response{Kind: ResponseRaw, Status: 200, ContentType: media, ContentHashHeader: name}
	}
	schema := Response{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "APIError"}
	expectHeader := func(required bool, media string) header {
		description := ""
		if media != "" {
			description = "Required for " + media + " responses."
		}
		return header{Required: required, Description: description, Schema: map[string]string{"format": "sha256", "type": "string"}}
	}
	for _, test := range []struct {
		name      string
		responses []Response
		headers   map[string]header
	}{
		{"single raw", []Response{raw(ContentGzip, "X-Content-SHA256")}, map[string]header{
			"X-Content-SHA256": expectHeader(true, ""),
		}},
		{"shared raw header", []Response{raw(ContentGzip, "X-Content-SHA256"), raw(ContentJSON, "X-Content-SHA256")}, map[string]header{
			"X-Content-SHA256": expectHeader(true, ""),
		}},
		{"case insensitive shared header", []Response{raw(ContentGzip, "X-Content-SHA256"), raw(ContentJSON, "X-content-sha256")}, map[string]header{
			"X-Content-SHA256": expectHeader(true, ""),
		}},
		{"distinct raw headers", []Response{raw(ContentGzip, "X-Gzip-SHA256"), raw(ContentJSON, "X-JSON-SHA256")}, map[string]header{
			"X-Gzip-SHA256": expectHeader(false, ContentGzip),
			"X-JSON-SHA256": expectHeader(false, ContentJSON),
		}},
		{"raw and schema", []Response{raw(ContentGzip, "X-Content-SHA256"), schema}, map[string]header{
			"X-Content-SHA256": expectHeader(false, ContentGzip),
		}},
		{"schema only", []Response{schema}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			operations := testOperations()
			// A schema response at another status must not make a success header optional.
			operations[1].Responses = append(test.responses, Response{Kind: ResponseSchema, Status: 404, ContentType: ContentJSON, SchemaRef: "APIError"})
			registry, err := NewRegistry(testSchemas(), operations)
			if err != nil {
				t.Fatal(err)
			}
			document, err := GenerateOpenAPI(registry, "Header test")
			if err != nil {
				t.Fatal(err)
			}
			var parsed struct {
				Paths map[string]map[string]struct {
					Responses map[string]struct {
						Headers map[string]header `json:"headers"`
					} `json:"responses"`
				} `json:"paths"`
			}
			if err := json.Unmarshal(document, &parsed); err != nil {
				t.Fatal(err)
			}
			responses := parsed.Paths[operations[1].Path]["get"].Responses
			if !reflect.DeepEqual(responses["200"].Headers, test.headers) {
				t.Errorf("success headers = %#v; want %#v", responses["200"].Headers, test.headers)
			}
			if responses["404"].Headers != nil {
				t.Errorf("raw success headers leaked into schema error: %#v", responses["404"].Headers)
			}
		})
	}
}

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

func TestGeneratedResponseMetadataPreservesRegisteredExactJSON(t *testing.T) {
	operations := testOperations()
	literals := [][]byte{
		[]byte("{\"category\":\"invalid\",\"detail\":\"body\"}\n"),
		[]byte("{\"category\":\"unknown_id\",\"detail\":\"a<>&🙂\"}\n"),
	}
	operations[1].Responses[1].ExactJSON = literals
	registry, err := NewRegistry(testSchemas(), operations)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := GenerateTypeScript(registry)
	if err != nil {
		t.Fatal(err)
	}
	// These are generated JS strings, not raw JSON inserted into JS syntax.
	// Decode the JSON-compatible string array and compare every literal byte.
	marker := []byte("exactJSON: ")
	start := bytes.Index(generated, marker)
	if start < 0 {
		t.Fatal("generated metadata omitted registered response literals")
	}
	var got []string
	decoder := json.NewDecoder(bytes.NewReader(generated[start+len(marker):]))
	if err := decoder.Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(literals) {
		t.Fatalf("generated %d literals, want %d", len(got), len(literals))
	}
	for index, literal := range literals {
		if got[index] != string(literal) {
			t.Fatalf("literal %d changed during generation", index)
		}
	}
	if bytes.Count(generated, marker) != 1 {
		t.Fatal("response literals leaked into unrelated operation/status metadata")
	}
}

func TestGeneratedClientSameStatusRepresentations(t *testing.T) {
	for _, kind := range []ResponseKind{ResponseRaw, ResponseSchema} {
		t.Run(string(kind), func(t *testing.T) {
			operations := testOperations()
			jsonResponse := Response{Kind: kind, Status: 200, ContentType: ContentJSON}
			if kind == ResponseRaw {
				jsonResponse.ContentHashHeader = "X-Content-SHA256"
			} else {
				jsonResponse.SchemaRef = "APIError"
			}
			operations[1].Responses = []Response{
				{Kind: ResponseRaw, Status: 200, ContentType: ContentGzip, ContentHashHeader: "X-Content-SHA256"},
				jsonResponse,
				{Kind: ResponseSchema, Status: 404, ContentType: ContentJSON, SchemaRef: "APIError"},
			}
			registry, err := NewRegistry(testSchemas(), operations)
			if err != nil {
				t.Fatalf("legal same-status representations refused: %v", err)
			}
			document, err := GenerateOpenAPI(registry, "Representation test")
			if err != nil {
				t.Fatal(err)
			}
			var parsed struct {
				Paths map[string]map[string]struct {
					Responses map[string]struct {
						Content map[string]json.RawMessage `json:"content"`
					} `json:"responses"`
				} `json:"paths"`
			}
			if err := json.Unmarshal(document, &parsed); err != nil {
				t.Fatal(err)
			}
			content := parsed.Paths[operations[1].Path]["get"].Responses["200"].Content
			if len(content) != 2 || content[ContentGzip] == nil || content[ContentJSON] == nil {
				t.Fatalf("OpenAPI lost a same-status representation: %s", document)
			}
			generated, err := GenerateTypeScript(registry)
			if err != nil {
				t.Fatal(err)
			}
			for _, representation := range []string{
				`{ status: 200, kind: "raw", contentType: "application/gzip", contentHashHeader: "X-Content-SHA256" }`,
				`{ status: 200, kind: "` + string(kind) + `", contentType: "application/json", contentHashHeader: "` + jsonResponse.ContentHashHeader + `" }`,
			} {
				if !bytes.Contains(generated, []byte(representation)) {
					t.Fatalf("generated client lost %s", representation)
				}
			}
		})
	}
}
