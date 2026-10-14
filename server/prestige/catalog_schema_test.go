package prestige

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
)

func prestigeSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(PrestigeSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestPrestigeOwnerSchemaAdmitsCurrentStoredAndRetunedPoliciesWithoutChangingBytes(t *testing.T) {
	definitions := prestigeSchemaDefinitions(t)
	for _, path := range []string{
		"../../balance/prestige/phase0.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/prestige.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/prestige.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(PrestigePolicySchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if _, err := LoadPolicy(data); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed artifact bytes")
		}
	}
	data := editPolicyFixture(t, func(v map[string]any) {
		v["threshold"] = "2e12"
		v["catchup_ceiling_ms"], v["offer_duration_ms"] = 6000, MaximumOfferDurationMS
		v["collapse_route_knowledge"], v["decline_drift_ppm"] = 0, 0
		v["exit_modifiers_ppm"].(map[string]any)["ipo"] = json.Number("9007199254740991")
	})
	if err := publicapi.ValidateJSON(PrestigePolicySchemaName, data, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPolicy(data); err != nil {
		t.Fatal(err)
	}
}

func TestPrestigeOwnerSchemaClosedPolicyBounds(t *testing.T) {
	definitions := prestigeSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"missing root":       func(v map[string]any) { delete(v, "collapse_route_knowledge") },
		"private root":       func(v map[string]any) { v["account_id"] = "private" },
		"root alias":         func(v map[string]any) { v["THRESHOLD"] = v["threshold"]; delete(v, "threshold") },
		"version":            func(v map[string]any) { v["schema_version"] = 2 },
		"value ID":           func(v map[string]any) { v["value_resource_id"] = "bad-id" },
		"threshold grammar":  func(v map[string]any) { v["threshold"] = "1000000" },
		"null zero":          func(v map[string]any) { v["decline_drift_ppm"] = nil },
		"negative knowledge": func(v map[string]any) { v["collapse_route_knowledge"] = -1 },
		"unsafe knowledge":   func(v map[string]any) { v["collapse_route_knowledge"] = json.Number("9007199254740992") },
		"catchup":            func(v map[string]any) { v["catchup_ceiling_ms"] = 86400001 },
		"duration":           func(v map[string]any) { v["offer_duration_ms"] = MaximumOfferDurationMS + 1 },
		"drift":              func(v map[string]any) { v["decline_drift_ppm"] = 1000001 },
		"advisor fractional": func(v map[string]any) { v["advisor_per_run_ppm"] = .5 },
		"modifier missing":   func(v map[string]any) { delete(v["exit_modifiers_ppm"].(map[string]any), "ipo") },
		"modifier null":      func(v map[string]any) { v["exit_modifiers_ppm"].(map[string]any)["ipo"] = nil },
		"modifier unknown":   func(v map[string]any) { v["exit_modifiers_ppm"].(map[string]any)["fictional"] = 0 },
		"null gate":          func(v map[string]any) { v["spawn_gate_ppm"].([]any)[0] = nil },
		"gate bounds":        func(v map[string]any) { v["spawn_gate_ppm"].([]any)[0] = 1000001 },
	} {
		t.Run(name, func(t *testing.T) {
			data := editPolicyFixture(t, edit)
			if err := publicapi.ValidateJSON(PrestigePolicySchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted invalid shape")
			}
			if _, err := LoadPolicy(data); err == nil {
				t.Fatal("owner admitted invalid shape")
			}
		})
	}
}

func TestPrestigeOwnerSchemaRetainsSemanticChecksAndFreshExports(t *testing.T) {
	definitions := prestigeSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"nonpositive threshold": func(v map[string]any) { v["threshold"] = "0" },
		"gate cardinality":      func(v map[string]any) { v["spawn_gate_ppm"] = []int64{0} },
	} {
		t.Run(name, func(t *testing.T) {
			data := editPolicyFixture(t, edit)
			if err := publicapi.ValidateJSON(PrestigePolicySchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadPolicy(data); err == nil {
				t.Fatal("structural descriptor replaced semantic policy checks")
			}
		})
	}
	before, fresh := PrestigeSchemas(), PrestigeSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("non-deterministic exports")
	}
	*fresh[0].Schema.Fields[0].Schema.Maximum = 0
	if !reflect.DeepEqual(before, PrestigeSchemas()) {
		t.Fatal("caller mutation changed owner authority")
	}
	source := bytes.TrimSpace(policyFixture(t))
	for _, data := range [][]byte{
		bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1),
		append(bytes.Clone(source[:len(source)-1]), []byte(`,"schema_version":2,"schema_version":1}`)...),
	} {
		if bytes.Equal(source, data) {
			t.Fatal("member control did not change source")
		}
		if err := publicapi.ValidateJSON(PrestigePolicySchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadPolicy(data); err != nil {
			t.Fatal(err)
		}
	}
}
