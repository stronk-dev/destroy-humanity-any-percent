package publicapi

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
)

func wireRequestOperation() Operation {
	return Operation{ID: "submit_example", Method: "POST", Path: "/api/v1/example", Surface: SurfacePrivateV1, Auth: AuthAccessToken,
		Request: "APIError", Responses: []Response{{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "APIError"}}}
}

func TestWireSchemaRejectsMalformedUTF8WithoutRepair(t *testing.T) {
	registry := testRegistry(t)
	request, err := NewRegistry(testSchemas(), []Operation{wireRequestOperation()})
	if err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []struct {
		name  string
		bytes []byte
	}{
		{"orphan", []byte{0x80}}, {"two-byte-truncated", []byte{0xc2}},
		{"three-byte-truncated", []byte{0xe2, 0x82}}, {"four-byte-truncated", []byte{0xf0, 0x9f, 0x99}},
		{"two-byte-overlong", []byte{0xc0, 0xaf}}, {"three-byte-overlong", []byte{0xe0, 0x80, 0xaf}},
		{"four-byte-overlong", []byte{0xf0, 0x80, 0x80, 0xaf}}, {"encoded-surrogate", []byte{0xed, 0xa0, 0x80}},
		{"above-Unicode-maximum", []byte{0xf4, 0x90, 0x80, 0x80}},
	} {
		body := append([]byte(`{"category":"invalid","detail":"before`), invalid.bytes...)
		body = append(body, []byte(`after"}`)...)
		success := append([]byte(`{"items":[{"epoch_id":1,"name":"before`), invalid.bytes...)
		success = append(success, []byte(`after","started_at":"2026-08-03T12:34:56.789Z"}],"next_cursor":null}`)...)
		for _, check := range []struct {
			name string
			call func() error
		}{
			{"request", func() error { return request.ValidateRequest("submit_example", body) }},
			{"success", func() error { return registry.ValidateResponse("get_epochs", 200, success) }},
			{"error", func() error { return registry.ValidateResponse("get_epochs", 400, body) }},
		} {
			t.Run(invalid.name+"/"+check.name, func(t *testing.T) {
				if err := check.call(); !errors.Is(err, ErrInvalidSchema) {
					t.Fatalf("malformed UTF-8 wire bytes were not refused: %v", err)
				}
			})
		}
		t.Run(invalid.name+"/literal", func(t *testing.T) {
			operation := wireRequestOperation()
			operation.Responses[0].ExactJSON = [][]byte{body}
			if _, err := NewRegistry(testSchemas(), []Operation{operation}); err == nil {
				t.Fatal("malformed UTF-8 literal accepted at construction")
			}
		})
	}
}

func TestWireEncodingPreservesUnicodeAndOwnerByteBoundaries(t *testing.T) {
	registry := testRegistry(t)
	request, err := NewRegistry(testSchemas(), []Operation{wireRequestOperation()})
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"aé中🙂\U0010ffff", "literal replacement \ufffd", `escaped \ud800`, `escaped \u00e9`} {
		body := []byte(`{"category":"invalid","detail":"` + text + `"}`)
		if request.ValidateRequest("submit_example", body) != nil || registry.ValidateResponse("get_epochs", 400, body) != nil {
			t.Fatal("existing valid Unicode/JSON escape refused")
		}
		operation := wireRequestOperation()
		operation.Responses[0].ExactJSON = [][]byte{body}
		if _, err := NewRegistry(testSchemas(), []Operation{operation}); err != nil {
			t.Fatalf("valid Unicode literal refused: %v", err)
		}
	}
	// This change belongs only to HTTP JSON admission. Shared artifact decoding
	// retains encoding/json semantics, and raw transport still proves byte identity.
	body := []byte("{\"category\":\"invalid\",\"detail\":\"\xff\"}")
	definitions, err := ValidateSchemaDefinitions(testSchemas())
	if err != nil || ValidateJSON("APIError", body, definitions) != nil {
		t.Fatal("shared owner decoder was silently tightened")
	}
	raw, err := NewRegistry(nil, []Operation{{ID: "get_raw", Method: "GET", Path: "/api/public/v1/raw", Surface: SurfacePublicV1, Auth: AuthNone, Public: true,
		Responses: []Response{{Kind: ResponseRaw, Status: 200, ContentType: ContentJSON, ContentHashHeader: "X-Content-SHA256"}}}})
	if err != nil || raw.ValidateRawResponse("get_raw", 200, ContentJSON, body, fmt.Sprintf("%x", sha256.Sum256(body))) != nil {
		t.Fatal("raw identity boundary was replaced by JSON decoding")
	}
	withBOM := append([]byte{0xef, 0xbb, 0xbf}, []byte(`{"category":"invalid","detail":"body"}`)...)
	if request.ValidateRequest("submit_example", withBOM) == nil || registry.ValidateResponse("get_epochs", 400, withBOM) == nil {
		t.Fatal("previous Go BOM refusal changed")
	}
	if !bytes.Equal(body, []byte("{\"category\":\"invalid\",\"detail\":\"\xff\"}")) {
		t.Fatal("validation mutated the evidence bytes")
	}
}
