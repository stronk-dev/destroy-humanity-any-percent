package formulas

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClientFormulaSchemaIsGeneratedFromOwner(t *testing.T) {
	generated, err := ClientSchemaModule()
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile(filepath.Join("..", "..", "client", "src", "api", "generated", "formula-schema.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(generated, committed) {
		t.Fatal("client formula metadata drift; run make formulas")
	}
	for _, named := range Schemas() {
		checkClientBounds(t, clientDescriptor(named.Schema))
	}
}

func checkClientBounds(t *testing.T, schema clientSchema) {
	t.Helper()
	if schema.Minimum != nil && schema.Maximum != nil && *schema.Minimum == "-9223372036854775808" && *schema.Maximum != "9223372036854775807" {
		t.Fatal("rounded signed int64 metadata")
	}
	for _, field := range schema.Fields {
		checkClientBounds(t, field.Schema)
	}
	if schema.Items != nil {
		checkClientBounds(t, *schema.Items)
	}
	for _, arm := range schema.Alternates {
		checkClientBounds(t, arm)
	}
}

func TestFormulaSharedParityControls(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "formula-artifact-parity-v14.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Cases []struct {
			ID      string `json:"id"`
			Replace string `json:"replace"`
			With    string `json:"with"`
			Valid   bool   `json:"valid"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(fixture, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Cases) == 0 {
		t.Fatal("empty parity population")
	}
	baseline := publishedArtifact(t)
	for _, control := range envelope.Cases {
		t.Run(control.ID, func(t *testing.T) {
			if bytes.Count(baseline, []byte(control.Replace)) != 1 {
				t.Fatal("control must replace exactly one published field")
			}
			data := bytes.Replace(baseline, []byte(control.Replace), []byte(control.With), 1)
			if err := Validate(data); (err == nil) != control.Valid {
				t.Fatalf("valid=%v err=%v", control.Valid, err)
			}
		})
	}
}

func TestFormulaJSONNestingMatchesDecoderWhenEarlierValueIsOverwritten(t *testing.T) {
	for _, control := range []struct {
		depth int
		valid bool
	}{{9998, true}, {9999, false}} {
		source := strings.Replace(string(publishedArtifact(t)), `"pinned": null`, `"pinned": `+strings.Repeat("[", control.depth)+`null`+strings.Repeat("]", control.depth)+`, "pinned": null`, 1)
		if err := Validate([]byte(source)); (err == nil) != control.valid {
			t.Fatalf("depth=%d valid=%v err=%v", control.depth, control.valid, err)
		}
	}
}
