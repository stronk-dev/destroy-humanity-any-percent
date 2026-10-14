package pitch

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/publicapi"
)

func pitchSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(PitchSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func pitchCopyDeclarations() Declarations {
	keys := make(map[string]struct{})
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	return Declarations{CopyKeys: keys}
}

func editPitchFixture(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	data, _ := loadFixture(t)
	var root map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	return mustPitchJSON(t, root)
}

func TestPitchOwnerSchemaAdmitsCurrentStoredAndRetunedCatalogs(t *testing.T) {
	definitions := pitchSchemaDefinitions(t)
	for _, path := range []string{
		"../../balance/pitch.json", "../../balance/testdata/pitch-v1.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/pitch.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/pitch.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(PitchCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if _, err := LoadCatalog(data, pitchCopyDeclarations()); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation modified artifact bytes")
		}
	}
	data := editPitchFixture(t, func(v map[string]any) {
		v["metric_cards"].([]any)[0].(map[string]any)["base_metric"] = "0"
		rows := v["growth_hacks"].([]any)
		rows[0].(map[string]any)["price"] = 0
		rows[1].(map[string]any)["price"] = json.Number("9007199254740991")
		rows[0].(map[string]any)["draft_weight"] = json.Number("9007199254740991")
	})
	if err := publicapi.ValidateJSON(PitchCatalogSchemaName, data, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(data, pitchCopyDeclarations()); err != nil {
		t.Fatal(err)
	}
}

func TestPitchOwnerSchemaClosesAllFourEffectArmsAndBounds(t *testing.T) {
	definitions := pitchSchemaDefinitions(t)
	row := func(v map[string]any) map[string]any { return v["growth_hacks"].([]any)[0].(map[string]any) }
	for name, edit := range map[string]func(map[string]any){
		"missing root":        func(v map[string]any) { delete(v, "growth_hacks") },
		"private root":        func(v map[string]any) { v["player_id"] = "private" },
		"root alias":          func(v map[string]any) { v["POLICY"] = v["policy"]; delete(v, "policy") },
		"version":             func(v map[string]any) { v["schema_version"] = 2 },
		"hand hardcap":        func(v map[string]any) { v["policy"].(map[string]any)["hand_size"] = 8 },
		"reason key":          func(v map[string]any) { v["policy"].(map[string]any)["hand_size_reason_key"] = "bad-key" },
		"exponent reason":     func(v map[string]any) { v["policy"].(map[string]any)["exponent_reason_key"] = "cap.other" },
		"card copies":         func(v map[string]any) { v["metric_cards"].([]any)[0].(map[string]any)["copies"] = 3 },
		"card Decimal":        func(v map[string]any) { v["metric_cards"].([]any)[0].(map[string]any)["base_metric"] = "15" },
		"null price":          func(v map[string]any) { row(v)["price"] = nil },
		"unsafe price":        func(v map[string]any) { row(v)["price"] = json.Number("9007199254740992") },
		"negative price":      func(v map[string]any) { row(v)["price"] = -1 },
		"weight zero":         func(v map[string]any) { row(v)["draft_weight"] = 0 },
		"price fraction":      func(v map[string]any) { row(v)["price"] = .5 },
		"flat missing amount": func(v map[string]any) { row(v)["effect"] = map[string]any{"kind": "flat_add"} },
		"card extra amount": func(v map[string]any) {
			row(v)["effect"] = map[string]any{"kind": "card_factor", "factor": "1e0", "amount": "1e0"}
		},
		"shape successor": func(v map[string]any) {
			row(v)["effect"] = map[string]any{"kind": "shape_factor", "shape": "triple", "factor": "1e0"}
		},
		"chain missing partner": func(v map[string]any) { row(v)["effect"] = map[string]any{"kind": "chain_factor", "factor": "1e0"} },
		"unknown effect":        func(v map[string]any) { row(v)["effect"] = map[string]any{"kind": "future"} },
		"funding round":         func(v map[string]any) { v["funding_curve"].([]any)[0].(map[string]any)["round"] = 9 },
		"null target":           func(v map[string]any) { v["funding_curve"].([]any)[0].(map[string]any)["funding_target"] = nil },
	} {
		t.Run(name, func(t *testing.T) {
			data := editPitchFixture(t, edit)
			if err := publicapi.ValidateJSON(PitchCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("schema admitted malformed catalog")
			}
			if _, err := LoadCatalog(data, pitchCopyDeclarations()); err == nil {
				t.Fatal("owner admitted malformed catalog")
			}
		})
	}
}

func TestPitchOwnerSchemaRetainsSemanticAndRawMemberAdmission(t *testing.T) {
	definitions := pitchSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"card count":      func(v map[string]any) { v["metric_cards"] = v["metric_cards"].([]any)[:11] },
		"card order":      func(v map[string]any) { rows := v["metric_cards"].([]any); rows[0], rows[1] = rows[1], rows[0] },
		"negative metric": func(v map[string]any) { v["metric_cards"].([]any)[0].(map[string]any)["base_metric"] = "-1e0" },
		"zero factor": func(v map[string]any) {
			v["growth_hacks"].([]any)[0].(map[string]any)["effect"].(map[string]any)["factor"] = "0"
		},
		"absent chain": func(v map[string]any) {
			v["growth_hacks"].([]any)[6].(map[string]any)["effect"].(map[string]any)["partner_hack_id"] = "missing"
		},
		"self chain": func(v map[string]any) {
			v["growth_hacks"].([]any)[6].(map[string]any)["effect"].(map[string]any)["partner_hack_id"] = "stealth_mode"
		},
		"copy binding": func(v map[string]any) {
			v["metric_cards"].([]any)[0].(map[string]any)["copy_key"] = "pitch.card.beta_invite"
		},
		"unknown reason": func(v map[string]any) { v["policy"].(map[string]any)["hand_size_reason_key"] = "cap.missing" },
		"funding order":  func(v map[string]any) { v["funding_curve"].([]any)[1].(map[string]any)["funding_target"] = "1e0" },
	} {
		t.Run(name, func(t *testing.T) {
			data := editPitchFixture(t, edit)
			if err := publicapi.ValidateJSON(PitchCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data, pitchCopyDeclarations()); err == nil {
				t.Fatal("descriptor replaced mandatory owner checks")
			}
		})
	}
	source, _ := loadFixture(t)
	escaped := bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1)
	if bytes.Equal(source, escaped) {
		t.Fatal("escaped control did not change bytes")
	}
	if _, err := LoadCatalog(escaped, pitchCopyDeclarations()); err != nil {
		t.Fatal(err)
	}
	trimmed := bytes.TrimSpace(source)
	duplicate := append(bytes.Clone(trimmed[:len(trimmed)-1]), []byte(`,"schema_version":1}`)...)
	if err := publicapi.ValidateJSON(PitchCatalogSchemaName, duplicate, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(duplicate, pitchCopyDeclarations()); err == nil {
		t.Fatal("owner's unique-member gate disappeared")
	}
	if err := publicapi.ValidateJSON(PitchCatalogSchemaName, escaped, definitions); err != nil {
		t.Fatal(err)
	}
	before, fresh := PitchSchemas(), PitchSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("nondeterministic exports")
	}
	for _, field := range fresh[0].Schema.Fields {
		if field.Name == "schema_version" {
			*field.Schema.Maximum = 0
		}
	}
	if !reflect.DeepEqual(before, PitchSchemas()) {
		t.Fatal("caller mutation changed owner schema")
	}
}
