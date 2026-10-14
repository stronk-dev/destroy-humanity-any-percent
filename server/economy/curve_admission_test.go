package economy

import (
	"encoding/json"
	"strings"
	"testing"
)

func costCurveCandidate(t *testing.T, curve map[string]any) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(loadKernelFixture(t).Catalog, &root); err != nil {
		t.Fatal(err)
	}
	root["generator_classes"].([]any)[0].(map[string]any)["price"].(map[string]any)["curve"] = curve
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCostCurvesRejectAliasesAndWrongArmNullFields(t *testing.T) {
	for _, base := range []map[string]any{
		{"kind": "constant"}, {"kind": "linear", "step": "0"}, {"kind": "geometric", "ratio": "1e0"},
	} {
		kind := base["kind"].(string)
		if _, err := LoadCatalog(costCurveCandidate(t, base)); err != nil {
			t.Fatal("legal curve control rejected", err)
		}
		for key, value := range base {
			t.Run(kind+"/alias/"+key, func(t *testing.T) {
				curve := make(map[string]any, len(base))
				for key, value := range base {
					curve[key] = value
				}
				curve[strings.ToUpper(key)] = value
				delete(curve, key)
				if _, err := LoadCatalog(costCurveCandidate(t, curve)); err == nil {
					t.Fatal("nonliteral cost-curve key admitted")
				}
			})
		}
		for _, wrong := range []string{"step", "ratio"} {
			if _, exists := base[wrong]; exists {
				continue
			}
			t.Run(kind+"/wrong-arm-null/"+wrong, func(t *testing.T) {
				curve := make(map[string]any, len(base)+1)
				for key, value := range base {
					curve[key] = value
				}
				curve[wrong] = nil
				if _, err := LoadCatalog(costCurveCandidate(t, curve)); err == nil {
					t.Fatal("wrong-arm null looked absent")
				}
			})
		}
	}
}
