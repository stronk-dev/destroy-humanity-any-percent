package main

import "testing"

func TestVerifySourceRequiresLiveJSONFieldBinding(t *testing.T) {
	item := binding{GoFunction: "emit", JSONField: "reason_key", Key: "reason.example", SourceFile: "server/example.go"}
	valid := []byte(`package fixture
func emit() { _, _ = json.Marshal(map[string]any{"reason_key": "reason.example"}) }
`)
	if err := verifySource("fixture.go", valid, item); err != nil {
		t.Fatal(err)
	}
	invalid := []byte(`package fixture
// "reason.example"
func emit() {
  if false { _, _ = json.Marshal(map[string]any{"reason_key": "reason.example"}) }
  _, _ = json.Marshal(map[string]any{"other": "reason.example"})
}
`)
	if err := verifySource("fixture.go", invalid, item); err == nil {
		t.Fatal("comment/dead serialized literal satisfied a producer binding")
	}
}

func TestDecodeRegistryRejectsUnknownAndUnsortedRows(t *testing.T) {
	unknown := []byte(`{"schema_version":1,"references":[{"go_function":"emit","json_field":"reason_key","key":"reason.example","source_file":"server/example.go","extra":true}]}`)
	if _, err := decodeRegistry(unknown); err == nil {
		t.Fatal("unknown field accepted")
	}
	unsorted := []byte(`{"schema_version":1,"references":[{"go_function":"emit","json_field":"reason_key","key":"reason.z","source_file":"server/z.go"},{"go_function":"emit","json_field":"reason_key","key":"reason.a","source_file":"server/a.go"}]}`)
	if _, err := decodeRegistry(unsorted); err == nil {
		t.Fatal("unsorted rows accepted")
	}
}
