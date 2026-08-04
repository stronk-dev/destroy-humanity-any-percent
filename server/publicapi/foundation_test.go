package publicapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
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
		{ID: "get_board", Method: "GET", Path: "/api/public/v1/boards/{category}", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, CursorKey: "TimeKey", Responses: map[int]string{200: "EpochPage", 400: "APIError"}},
		{ID: "get_epochs", Method: "GET", Path: "/api/public/v1/epochs", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: map[int]string{200: "EpochPage", 400: "APIError"}},
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
	registry, err := NewRegistry(schemas, operations)
	if err != nil {
		t.Fatal(err)
	}
	schemas[2].Schema.Fields[0].Name = "corrupted"
	operations[1].Responses[200] = "APIError"
	schemaSnapshot := registry.Schemas()
	schemaSnapshot[2].Schema.Fields[0].Name = "also_corrupted"
	operationSnapshot := registry.Operations()
	operationSnapshot[1].Responses[200] = "APIError"
	valid := []byte(`{"items":[{"epoch_id":1,"name":"Phase 0","started_at":"2026-08-03T12:34:56.789Z"}],"next_cursor":null}`)
	if err := registry.ValidateResponse("get_epochs", 200, valid); err != nil {
		t.Fatalf("registry changed after external mutation: %v", err)
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
	base := Operation{ID: "get_epochs", Method: "GET", Path: "/api/public/v1/epochs", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: map[int]string{200: "EpochPage"}}
	if _, err := NewRegistry(testSchemas(), []Operation{base, {ID: "get_epochs_two", Method: base.Method, Path: base.Path, Surface: base.Surface, Auth: base.Auth, Public: true, Responses: base.Responses}}); err == nil {
		t.Fatal("duplicate route accepted")
	}
	base.Auth = AuthAccessToken
	if _, err := NewRegistry(testSchemas(), []Operation{base}); err == nil {
		t.Fatal("authenticated public operation accepted")
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
