package prestige

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func policyFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../balance/prestige/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func editPolicyFixture(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(policyFixture(t), &root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLoadPolicyRejectsCaseAliasedMembers(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal(policyFixture(t), &root); err != nil {
		t.Fatal(err)
	}
	for key := range root {
		t.Run(key, func(t *testing.T) {
			data := editPolicyFixture(t, func(v map[string]any) { v[strings.ToUpper(key)] = v[key]; delete(v, key) })
			if _, err := LoadPolicy(data); err == nil {
				t.Fatal("case alias admitted")
			}
		})
	}
}

func TestLoadPolicyRequiresExplicitNonNullZeroMembers(t *testing.T) {
	for _, key := range []string{"collapse_route_knowledge", "decline_drift_ppm", "advisor_per_run_ppm", "advisor_cap_ppm"} {
		for _, missing := range []bool{false, true} {
			t.Run(key+"/"+map[bool]string{false: "null", true: "missing"}[missing], func(t *testing.T) {
				data := editPolicyFixture(t, func(v map[string]any) {
					v[key] = nil
					if missing {
						delete(v, key)
					}
				})
				if _, err := LoadPolicy(data); err == nil {
					t.Fatal("missing/null value silently became zero")
				}
			})
		}
	}
	for _, kind := range []string{"acquihire", "acquisition", "collapse", "ipo", "scripted_first"} {
		t.Run("modifier/"+kind, func(t *testing.T) {
			data := editPolicyFixture(t, func(v map[string]any) { v["exit_modifiers_ppm"].(map[string]any)[kind] = nil })
			if _, err := LoadPolicy(data); err == nil {
				t.Fatal("null modifier silently became zero")
			}
		})
	}
	for index := range 10 {
		data := editPolicyFixture(t, func(v map[string]any) { v["spawn_gate_ppm"].([]any)[index] = nil })
		if _, err := LoadPolicy(data); err == nil {
			t.Errorf("null spawn gate at index %d silently became zero", index)
		}
	}
	source := bytes.TrimSpace(policyFixture(t))
	partial := append(bytes.Clone(source[:len(source)-1]), []byte(`,"exit_modifiers_ppm":{"acquihire":900000}}`)...)
	if _, err := LoadPolicy(partial); err == nil {
		t.Error("partial final modifier object merged with earlier keys")
	}
	zero := editPolicyFixture(t, func(v map[string]any) {
		for _, key := range []string{"collapse_route_knowledge", "decline_drift_ppm", "advisor_per_run_ppm", "advisor_cap_ppm"} {
			v[key] = 0
		}
		for key := range v["exit_modifiers_ppm"].(map[string]any) {
			v["exit_modifiers_ppm"].(map[string]any)[key] = 0
		}
		v["spawn_gate_ppm"] = []int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	})
	if _, err := LoadPolicy(zero); err != nil {
		t.Fatal("explicit zero refused:", err)
	}
}

func TestLoadPolicyExistingDurationRepresentationBoundary(t *testing.T) {
	maximum := int64((1<<63 - 1) / time.Millisecond)
	for _, value := range []int64{maximum, maximum + 1} {
		data := editPolicyFixture(t, func(v map[string]any) { v["offer_duration_ms"] = value })
		_, err := LoadPolicy(data)
		if (err == nil) != (value == maximum) {
			t.Fatalf("duration %d: %v", value, err)
		}
	}
}
