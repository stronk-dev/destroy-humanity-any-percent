package publicapi

import (
	"reflect"
	"testing"
)

func TestSharedErrorSchemaRetainsClosedShapeAndFreshExports(t *testing.T) {
	first, second := APIErrorSchema(), APIErrorSchema()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("shared schema export is nondeterministic")
	}
	definitions, err := ValidateSchemaDefinitions([]NamedSchema{first})
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range []string{
		`{"category":"invalid","detail":"body"}`,
		`{"category":"unauthorized","detail":"access_token"}`,
		`{"category":"internal_invariant","detail":"public_api"}`,
		`{"category":"refresh_reused","detail":"session_family_revoked"}`,
	} {
		if err := ValidateJSON("APIError", []byte(data), definitions); err != nil {
			t.Fatalf("existing shared error rejected: %s: %v", data, err)
		}
	}
	for _, data := range []string{
		`{"category":"invalid"}`,
		`{"detail":"body"}`,
		`{"category":"invalid","detail":"body","account_id":"private"}`,
		`{"category":"invented","detail":"body"}`,
		`{"category":"invalid","detail":"invented"}`,
		`{"category":null,"detail":"body"}`,
		`{"category":"invalid","detail":false}`,
	} {
		if err := ValidateJSON("APIError", []byte(data), definitions); err == nil {
			t.Fatalf("open or malformed error accepted: %s", data)
		}
	}
	first.Schema.Fields[0].Schema.Enum[0] = "mutated"
	if !reflect.DeepEqual(second, APIErrorSchema()) {
		t.Fatal("caller mutation corrupted the shared owner")
	}
	if err := ValidateJSON("APIError", []byte(`{"category":"conflict","detail":"body"}`), definitions); err != nil {
		t.Fatal("caller mutation corrupted already-validated definitions")
	}
}
