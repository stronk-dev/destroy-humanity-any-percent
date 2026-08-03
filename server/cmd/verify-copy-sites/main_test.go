package main

import "testing"

func TestVerifySourceRequiresLiveJSONFieldBinding(t *testing.T) {
	item := binding{GoFunction: "emit", JSONField: "reason_key", Key: "reason.example", SourceFile: "server/example.go"}
	valid := []byte(`package fixture
func emit() {
  payload, _ := json.Marshal(map[string]any{"reason_key": "reason.example"})
  if _, err := tx.ExecContext(ctx, "INSERT", payload); err != nil { return }
}
`)
	if err := verifySource("fixture.go", valid, item); err != nil {
		t.Fatal(err)
	}
	invalid := []byte(`package fixture
// "reason.example"
func emit() {
  if 1 == 2 {
    payload, _ := json.Marshal(map[string]any{"reason_key": "reason.example"})
    if _, err := tx.ExecContext(ctx, "INSERT", payload); err != nil { return }
  }
  discarded, _ := json.Marshal(map[string]any{"reason_key": "reason.example"})
  _ = discarded
  payload, _ := json.Marshal(map[string]any{"other": "reason.example"})
  if _, err := tx.ExecContext(ctx, "INSERT", payload); err != nil { return }
}
`)
	if err := verifySource("fixture.go", invalid, item); err == nil {
		t.Fatal("dead, discarded, or wrong-field serialization satisfied a producer binding")
	}
}

func TestVerifySourceRejectsNeverInvokedFunctionLiteral(t *testing.T) {
	item := binding{GoFunction: "emit", JSONField: "reason_key", Key: "reason.example", SourceFile: "server/example.go"}
	invalid := []byte(`package fixture
func emit() {
  unused := func() {
    payload, _ := json.Marshal(map[string]any{"reason_key": "reason.example"})
    if _, err := tx.ExecContext(ctx, "INSERT", payload); err != nil { return }
  }
  _ = unused
}
`)
	if err := verifySource("fixture.go", invalid, item); err == nil {
		t.Fatal("never-invoked nested producer satisfied a binding")
	}
}

func TestVerifySourceRejectsUncheckedOrShadowedSink(t *testing.T) {
	item := binding{GoFunction: "emit", JSONField: "reason_key", Key: "reason.example", SourceFile: "server/example.go"}
	for _, invalid := range []string{
		`package fixture
func emit() {
  payload, _ := json.Marshal(map[string]any{"reason_key": "reason.example"})
  if _, err := tx.ExecContext(ctx, "INSERT", payload); false { _ = err }
}`,
		`package fixture
func emit() {
  payload, _ := json.Marshal(map[string]any{"reason_key": "reason.example"})
  payload = []byte("wrong")
  if _, err := tx.ExecContext(ctx, "INSERT", payload); err != nil { return }
}`,
	} {
		if err := verifySource("fixture.go", []byte(invalid), item); err == nil {
			t.Fatal("unchecked or shadowed payload sink satisfied a binding")
		}
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
