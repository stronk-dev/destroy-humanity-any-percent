package publicapi

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
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
	}
}

func TestSchemaRegistryIsClosedAndValidatesRuntimeBytes(t *testing.T) {
	registry, err := NewRegistry(testSchemas(), []Operation{
		{ID: "get_epochs", Method: "GET", Path: "/api/public/v1/epochs", Surface: SurfacePublicV1, Auth: AuthNone, Public: true, Responses: map[int]string{200: "EpochPage", 400: "APIError"}},
	})
	if err != nil {
		t.Fatal(err)
	}
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
	codec, err := NewCursorCodec(current, previous)
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
	oldCodec, _ := NewCursorCodec(oldKey, nil)
	rotated, _ := NewCursorCodec(newKey, oldKey)
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
