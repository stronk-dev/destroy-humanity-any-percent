package publicapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func integerPointer(value int64) *int64 { return &value }

func testSchemas() []NamedSchema {
	return []NamedSchema{
		{Name: "APIError", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
			{Name: "category", Schema: &Schema{Kind: SchemaString, Enum: []string{"invalid", "unknown_id"}}, Required: true},
			{Name: "detail", Schema: &Schema{Kind: SchemaString}, Required: true},
		}}},
		{Name: "EpochPage", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
			{Name: "items", Schema: &Schema{Kind: SchemaArray, Items: &Schema{Kind: SchemaRef, Ref: "EpochRow"}}, Required: true},
			{Name: "next_cursor", Schema: &Schema{Kind: SchemaOneOf, Alternates: []*Schema{{Kind: SchemaNull}, {Kind: SchemaString}}}, Required: true},
		}}},
		{Name: "EpochRow", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
			{Name: "epoch_id", Schema: &Schema{Kind: SchemaInteger, Minimum: integerPointer(1)}, Required: true},
			{Name: "name", Schema: &Schema{Kind: SchemaString}, Required: true},
			{Name: "started_at", Schema: &Schema{Kind: SchemaString, Format: "date-time-ms"}, Required: true},
		}}},
		{Name: "TimeKey", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
			{Name: "run_id", Schema: &Schema{Kind: SchemaString}, Required: true},
			{Name: "time_ms", Schema: &Schema{Kind: SchemaInteger, Minimum: integerPointer(0)}, Required: true},
		}}},
	}
}

func testOperations() []Operation {
	return []Operation{
		{ID: "get_board", Method: "GET", Path: "/api/public/v1/boards/{category}", Surface: SurfacePublicV1, Auth: AuthNone, Public: true,
			Parameters: []Parameter{{Name: "category", Schema: &Schema{Kind: SchemaString, Format: "mechanical-id"}}}, CursorKey: "TimeKey", Responses: []Response{
				{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "EpochPage"},
				{Kind: ResponseSchema, Status: 400, ContentType: ContentJSON, SchemaRef: "APIError"},
			}},
		{ID: "get_epochs", Method: "GET", Path: "/api/public/v1/epochs", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: []Response{
			{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "EpochPage"},
			{Kind: ResponseSchema, Status: 400, ContentType: ContentJSON, SchemaRef: "APIError"},
		}},
	}
}

func testRegistry(t *testing.T) *Registry {
	t.Helper()
	registry, err := NewRegistry(testSchemas(), testOperations())
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func TestSchemaRegistryIsClosedAndValidatesRuntimeBytes(t *testing.T) {
	registry := testRegistry(t)
	valid := []byte(`{"items":[{"epoch_id":1,"name":"Phase 0","started_at":"2026-08-03T12:34:56.789Z"}],"next_cursor":null}`)
	if err := registry.ValidateResponse("get_epochs", 200, valid); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range [][]byte{
		[]byte(`{"items":[],"next_cursor":null,"private_founder_id":"x"}`),
		[]byte(`{"items":[{"epoch_id":0,"name":"x","started_at":"2026-08-03T12:34:56.789Z"}],"next_cursor":null}`),
		[]byte(`{"items":[],"next_cursor":false}`),
	} {
		if err := registry.ValidateResponse("get_epochs", 200, invalid); err == nil {
			t.Fatalf("invalid response accepted: %s", invalid)
		}
	}
}

func TestSchemaRegistryRejectsReferenceCycles(t *testing.T) {
	for name, schemas := range map[string][]NamedSchema{
		"direct": {{Name: "Loop", Schema: &Schema{Kind: SchemaRef, Ref: "Loop"}}},
		"indirect": {
			{Name: "First", Schema: &Schema{Kind: SchemaObject, Fields: []Field{{Name: "next", Schema: &Schema{Kind: SchemaRef, Ref: "Second"}, Required: true}}}},
			{Name: "Second", Schema: &Schema{Kind: SchemaArray, Items: &Schema{Kind: SchemaRef, Ref: "First"}}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ValidateSchemaDefinitions(schemas); !errors.Is(err, ErrInvalidSchema) {
				t.Fatalf("cycle accepted: %v", err)
			}
		})
	}
}

func TestSchemaRegistryRejectsAcyclicGraphBeyondRuntimeDepth(t *testing.T) {
	schemas := make([]NamedSchema, 66)
	for index := range schemas {
		name := fmt.Sprintf("S%02d", index)
		schema := &Schema{Kind: SchemaString}
		if index < len(schemas)-1 {
			schema = &Schema{Kind: SchemaRef, Ref: fmt.Sprintf("S%02d", index+1)}
		}
		schemas[index] = NamedSchema{Name: name, Schema: schema}
	}
	if _, err := ValidateSchemaDefinitions(schemas); !errors.Is(err, ErrInvalidSchema) {
		t.Fatalf("runtime-unusable schema chain accepted: %v", err)
	}
}

func TestRegistryOwnsImmutableSchemaAndOperationSnapshots(t *testing.T) {
	schemas := testSchemas()
	operations := testOperations()
	exact := []byte(`{"category":"invalid","detail":"body"}` + "\n")
	operations[1].Responses[1].ExactJSON = [][]byte{bytes.Clone(exact)}
	registry, err := NewRegistry(schemas, operations)
	if err != nil {
		t.Fatal(err)
	}
	schemas[2].Schema.Fields[0].Name = "corrupted"
	operations[1].Responses[0].SchemaRef = "APIError"
	operations[1].Responses[1].ExactJSON[0][0] = '!'
	schemaSnapshot := registry.Schemas()
	schemaSnapshot[2].Schema.Fields[0].Name = "also_corrupted"
	operationSnapshot := registry.Operations()
	operationSnapshot[1].Responses[0].SchemaRef = "APIError"
	operationSnapshot[1].Responses[1].ExactJSON[0][0] = '?'
	valid := []byte(`{"items":[{"epoch_id":1,"name":"Phase 0","started_at":"2026-08-03T12:34:56.789Z"}],"next_cursor":null}`)
	if err := registry.ValidateResponse("get_epochs", 200, valid); err != nil {
		t.Fatalf("registry changed after external mutation: %v", err)
	}
	if err := registry.ValidateResponse("get_epochs", http.StatusBadRequest, exact); err != nil {
		t.Fatalf("exact response authority changed after external mutation: %v", err)
	}
}

func TestRegistryExactJSONDiscriminatesSchemaValidCrossProductsAndExtraBytes(t *testing.T) {
	operations := testOperations()
	exact := []byte(`{"category":"invalid","detail":"body"}` + "\n")
	operations[1].Responses[1].ExactJSON = [][]byte{bytes.Clone(exact)}
	registry, err := NewRegistry(testSchemas(), operations)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.ValidateResponse("get_epochs", http.StatusBadRequest, exact); err != nil {
		t.Fatalf("exact response rejected: %v", err)
	}
	for _, invalid := range [][]byte{
		[]byte(`{"category":"unknown_id","detail":"body"}` + "\n"),
		append(bytes.Clone(exact), '!'),
	} {
		if err := registry.ValidateResponse("get_epochs", http.StatusBadRequest, invalid); err == nil {
			t.Fatalf("non-literal response accepted: %q", invalid)
		}
	}

	unsorted := testOperations()
	unsorted[1].Responses[1].ExactJSON = [][]byte{
		[]byte(`{"category":"unknown_id","detail":"body"}` + "\n"),
		bytes.Clone(exact),
	}
	if _, err := NewRegistry(testSchemas(), unsorted); err == nil {
		t.Fatal("unsorted exact JSON authority accepted")
	}
}

func TestCanonicalDecimalFormatUsesNumericCoreGrammar(t *testing.T) {
	definitions, err := ValidateSchemaDefinitions([]NamedSchema{{Name: "Decimal", Schema: &Schema{Kind: SchemaString, Format: "canonical-decimal"}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"0", "1e0", "-1e0", "1.23456789012e4000000000000000"} {
		encoded, _ := json.Marshal(value)
		if err := ValidateJSON("Decimal", encoded, definitions); err != nil {
			t.Fatalf("valid canonical decimal %q rejected: %v", value, err)
		}
	}
	for _, value := range []string{"10e0", "1.0e0", "1e+1", "1e9000000000000000", "-0"} {
		encoded, _ := json.Marshal(value)
		if err := ValidateJSON("Decimal", encoded, definitions); err == nil {
			t.Fatalf("noncanonical decimal %q accepted", value)
		}
	}
}

func TestRegistryRejectsDuplicateAndPublicAuthDrift(t *testing.T) {
	base := Operation{ID: "get_epochs", Method: "GET", Path: "/api/public/v1/epochs", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: []Response{{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "EpochPage"}}}
	if _, err := NewRegistry(testSchemas(), []Operation{base, {ID: "get_epochs_two", Method: base.Method, Path: base.Path, Surface: base.Surface, Auth: base.Auth, Public: true, Responses: base.Responses}}); err == nil {
		t.Fatal("duplicate route accepted")
	}
	base.Auth = AuthAccessToken
	if _, err := NewRegistry(testSchemas(), []Operation{base}); err == nil {
		t.Fatal("authenticated public operation accepted")
	}
}

func TestRegistryRequiresExactPathParameterDescriptors(t *testing.T) {
	base := testOperations()
	base[0].Parameters = []Parameter{{Name: "category", Schema: &Schema{Kind: SchemaString, Format: "mechanical-id"}}}
	if _, err := NewRegistry(testSchemas(), base); err != nil {
		t.Fatal(err)
	}
	for _, parameters := range [][]Parameter{
		nil,
		{{Name: "wrong", Schema: &Schema{Kind: SchemaString}}},
		{{Name: "category", Schema: &Schema{Kind: SchemaObject}}},
	} {
		invalid := testOperations()
		invalid[0].Parameters = parameters
		if _, err := NewRegistry(testSchemas(), invalid); err == nil {
			t.Fatalf("invalid path parameters accepted: %+v", parameters)
		}
	}
}

func TestRegistryResponseUnionValidatesSchemaAndRawBytes(t *testing.T) {
	operations := testOperations()
	operations[1].Responses = []Response{
		{Kind: ResponseRaw, Status: 200, ContentType: ContentGzip, ContentHashHeader: "X-Content-SHA256"},
		{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "EpochPage"},
		{Kind: ResponseSchema, Status: 400, ContentType: ContentJSON, SchemaRef: "APIError"},
	}
	registry, err := NewRegistry(testSchemas(), operations)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte("immutable gzip bytes")
	digest := fmt.Sprintf("%x", sha256.Sum256(body))
	if err := registry.ValidateRawResponse("get_epochs", 200, ContentGzip, body, digest); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct {
		contentType string
		body        []byte
		digest      string
	}{
		{ContentJSON, body, digest},
		{ContentGzip, append(append([]byte(nil), body...), '!'), digest},
		{ContentGzip, body, strings.Repeat("0", 64)},
	} {
		if err := registry.ValidateRawResponse("get_epochs", 200, mutation.contentType, mutation.body, mutation.digest); err == nil {
			t.Fatalf("invalid raw response accepted: %+v", mutation)
		}
	}
	if err := registry.ValidateResponse("get_epochs", 200, []byte(`{"items":[],"next_cursor":null}`)); err != nil {
		t.Fatal(err)
	}

	invalid := testOperations()
	invalid[1].Responses = []Response{{Kind: ResponseRaw, Status: 200, ContentType: "application/octet-stream", ContentHashHeader: "X-Content-SHA256"}}
	if _, err := NewRegistry(testSchemas(), invalid); err == nil {
		t.Fatal("generic raw media type accepted")
	}
}

func TestRegistryMountRejectsBindingDriftAndAppliesDeclaredAuth(t *testing.T) {
	operations := testOperations()
	operations[0].Auth, operations[0].Public, operations[0].Surface, operations[0].Path = AuthAccessToken, false, SurfacePrivateV1, "/api/v1/boards/{category}"
	registry, err := NewRegistry(testSchemas(), operations)
	if err != nil {
		t.Fatal(err)
	}
	newBindings := func() []Binding {
		return []Binding{
			{OperationID: "get_board", Handler: http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })},
			{OperationID: "get_epochs", Handler: http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) })},
		}
	}
	for _, mutate := range []func([]Binding) []Binding{
		func(values []Binding) []Binding { return values[:1] },
		func(values []Binding) []Binding { values[0].OperationID = "get_epochs"; return values },
		func(values []Binding) []Binding { values[0].Handler = nil; return values },
	} {
		if err := registry.Mount(chi.NewRouter(), mutate(newBindings()), map[AuthMode]Middleware{AuthAccessToken: func(next http.Handler) http.Handler { return next }}); err == nil {
			t.Fatal("binding drift mounted")
		}
	}

	router := chi.NewRouter()
	authenticated := false
	err = registry.Mount(router, newBindings(), map[AuthMode]Middleware{AuthAccessToken: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			authenticated = true
			next.ServeHTTP(response, request)
		})
	}})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/boards/any_percent", nil))
	if response.Code != http.StatusNoContent || !authenticated {
		t.Fatalf("status=%d authenticated=%t", response.Code, authenticated)
	}
}

type timeKey struct {
	RunID  string `json:"run_id"`
	TimeMS int64  `json:"time_ms"`
}

func TestCursorValidatesMACBeforeParsingAndBindsQuery(t *testing.T) {
	current := []byte("0123456789abcdef0123456789abcdef")
	previous := []byte("abcdef0123456789abcdef0123456789")
	codec, err := NewCursorCodec(current, previous, testRegistry(t))
	if err != nil {
		t.Fatal(err)
	}
	filter := strings.Repeat("a", 64)
	token, err := codec.Encode("get_board", filter, timeKey{RunID: "run-9", TimeMS: 1234})
	if err != nil {
		t.Fatal(err)
	}
	var decoded timeKey
	if err := codec.Decode(token, "get_board", filter, &decoded); err != nil || decoded.RunID != "run-9" || decoded.TimeMS != 1234 {
		t.Fatalf("decoded=%+v err=%v", decoded, err)
	}
	if err := codec.Decode(token, "get_epochs", filter, &decoded); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("cross-operation cursor accepted")
	}
	if _, err := codec.Encode("unknown_operation", filter, timeKey{RunID: "run-9", TimeMS: 1234}); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("unknown operation accepted")
	}
	if _, err := codec.Encode("get_epochs", filter, timeKey{RunID: "run-9", TimeMS: 1234}); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("non-paginated operation accepted")
	}
	if _, err := codec.Encode("get_board", filter, map[string]any{"run_id": "run-9"}); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("missing cursor key field accepted")
	}
	var incomplete struct {
		RunID string `json:"run_id"`
	}
	if err := codec.Decode(token, "get_board", filter, &incomplete); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("cursor key target that cannot exact-reencode was accepted")
	}
	if err := codec.Decode(token, "get_board", strings.Repeat("b", 64), &decoded); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("cross-filter cursor accepted")
	}

	raw, _ := base64.RawURLEncoding.DecodeString(token)
	raw[len(raw)-1] ^= 1
	if err := codec.Decode(base64.RawURLEncoding.EncodeToString(raw), "get_board", filter, &decoded); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("tampered cursor accepted")
	}
	invalidBeforeMAC := append([]byte(`not-json`), make([]byte, sha256.Size)...)
	if err := codec.Decode(base64.RawURLEncoding.EncodeToString(invalidBeforeMAC), "get_board", filter, &decoded); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("invalid unsigned payload accepted")
	}
}

func TestCursorKeyRotationAndCanonicalBoardVariables(t *testing.T) {
	oldKey := []byte("old-0123456789abcdef0123456789ab")
	newKey := []byte("new-0123456789abcdef0123456789ab")
	registry := testRegistry(t)
	oldCodec, _ := NewCursorCodec(oldKey, nil, registry)
	rotated, _ := NewCursorCodec(newKey, oldKey, registry)
	filter := strings.Repeat("c", 64)
	token, err := oldCodec.Encode("get_board", filter, timeKey{RunID: "run-1", TimeMS: 1})
	if err != nil {
		t.Fatal(err)
	}
	var decoded timeKey
	if err := rotated.Decode(token, "get_board", filter, &decoded); err != nil {
		t.Fatal(err)
	}

	faction := "open_source"
	encoded, err := EncodeBoardVariables(BoardVariables{Advisor: 1, Commons: 0, Faction: &faction, Glitched: 1})
	if err != nil {
		t.Fatal(err)
	}
	variables, err := DecodeBoardVariables(encoded)
	if err != nil || variables.Faction == nil || *variables.Faction != faction {
		t.Fatalf("variables=%+v err=%v", variables, err)
	}
	noncanonical := base64.RawURLEncoding.EncodeToString([]byte("{\"commons\":0,\"advisor\":1,\"faction\":\"open_source\",\"glitched\":1}"))
	if _, err := DecodeBoardVariables(noncanonical); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("noncanonical variables accepted")
	}
	if _, err := DecodeBoardVariables(encoded + "="); !errors.Is(err, ErrInvalidCursor) {
		t.Fatal("padded variables accepted")
	}
}

func TestBoardFilterHashHasOneCanonicalAuthority(t *testing.T) {
	faction := "open_source"
	base := BoardFilter{Category: "any_percent", Variables: BoardVariables{Advisor: 1, Commons: 0, Faction: &faction, Glitched: 1}, EpochID: 7, MandateLevel: 3, Limit: 50}
	encoded, err := EncodeBoardFilter(base)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"category":"any_percent","epoch":7,"limit":50,"mandate":3,"variables":{"advisor":1,"commons":0,"faction":"open_source","glitched":1}}` {
		t.Fatalf("unexpected canonical filter: %s", encoded)
	}
	digest, err := BoardFilterSHA256(base)
	if err != nil || !filterHashPattern.MatchString(digest) {
		t.Fatalf("digest=%q err=%v", digest, err)
	}
	mutations := []BoardFilter{base, base, base, base, base}
	mutations[0].Category = "ethical_percent"
	mutations[1].EpochID++
	mutations[2].Limit++
	mutations[3].MandateLevel++
	mutations[4].Variables.Commons = 1
	for _, mutation := range mutations {
		other, err := BoardFilterSHA256(mutation)
		if err != nil || other == digest {
			t.Fatalf("filter dimension was not bound: %+v digest=%q err=%v", mutation, other, err)
		}
	}
}
