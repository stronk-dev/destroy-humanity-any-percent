package minigame

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestScalingPolicyRequiresExactMandatoryFields(t *testing.T) {
	data, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var catalog map[string]json.RawMessage
	var definitions []map[string]json.RawMessage
	if json.Unmarshal(data, &catalog) != nil || json.Unmarshal(catalog["minigames"], &definitions) != nil || len(definitions) != 1 {
		t.Fatal("invalid canonical catalog")
	}
	// A legitimate zero minimum exposes null/missing values that Go otherwise
	// silently defaults. A missing maximum would change the resolved 3 to 0.
	const baseline = `{"schema_version":1,"scaling_inputs":[{"destination":"option.count","destination_class":"breadth","source_kind":"literal","source_ref":"3","op":"identity","operand":0,"clamp_min":0,"clamp_max":8}]}`
	loadCatalog := func(policy []byte) (*Catalog, error) {
		definitions[0]["scaling"] = policy
		catalog["minigames"] = scalingTestJSON(t, definitions)
		return LoadCatalog(scalingTestJSON(t, catalog))
	}
	assertValid := func(raw []byte) {
		t.Helper()
		policy, err := LoadScalingPolicy(raw, true)
		if err != nil {
			t.Fatal(err)
		}
		resolved, err := policy.Resolve(ScalingContext{PurchasedGeneratorCounts: map[string]int64{},
			FounderCarryCounters: map[string]int64{}, AttendedQualityGrades: map[string]int64{}})
		if err != nil || resolved["option.count"] != 3 {
			t.Fatalf("canonical scaling changed: %v err=%v", resolved, err)
		}
		if _, err := loadCatalog(raw); err != nil {
			t.Fatalf("canonical catalog refused: %v", err)
		}
	}
	assertValid([]byte(baseline))
	assertValid([]byte(strings.Replace(baseline, `"destination"`, `"\u0064estination"`, 1)))
	for _, scope := range []struct {
		name string
		keys []string
	}{
		{"root", []string{"schema_version", "scaling_inputs"}},
		{"row", []string{"destination", "destination_class", "source_kind", "source_ref", "op", "operand", "clamp_min", "clamp_max"}},
	} {
		for _, key := range scope.keys {
			for _, variant := range []string{"missing", "null", "wrong_case", "case_collision"} {
				t.Run(scope.name+"/"+key+"/"+variant, func(t *testing.T) {
					var root map[string]json.RawMessage
					var rows []map[string]json.RawMessage
					if json.Unmarshal([]byte(baseline), &root) != nil || json.Unmarshal(root["scaling_inputs"], &rows) != nil {
						t.Fatal("invalid baseline policy")
					}
					fields := root
					if scope.name == "row" {
						fields = rows[0]
					}
					switch variant {
					case "missing":
						delete(fields, key)
					case "null":
						fields[key] = json.RawMessage("null")
					case "wrong_case", "case_collision":
						fields[strings.ToUpper(key)] = fields[key]
						if variant == "wrong_case" {
							delete(fields, key)
						}
					}
					if scope.name == "row" {
						root["scaling_inputs"] = scalingTestJSON(t, rows)
					}
					bad := scalingTestJSON(t, root)
					if policy, err := LoadScalingPolicy(bad, true); !errors.Is(err, ErrInvalidScalingPolicy) || policy != nil {
						t.Errorf("malformed policy admitted: %s err=%v", bad, err)
					}
					if catalog, err := loadCatalog(bad); !errors.Is(err, ErrInvalidCatalog) || catalog != nil {
						t.Errorf("malformed nested catalog admitted: %v", err)
					}
				})
			}
		}
	}
}

func scalingTestJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestScalingSourceReferenceVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/minigame/scaling-source-refs-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var population struct {
		Cases []struct {
			Name    string                     `json:"name"`
			Changes map[string]json.RawMessage `json:"changes"`
			Valid   bool                       `json:"valid"`
		} `json:"cases"`
	}
	if json.Unmarshal(data, &population) != nil || len(population.Cases) != 29 {
		t.Fatal("incomplete scaling source-reference population")
	}
	base, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range population.Cases {
		t.Run(row.Name, func(t *testing.T) {
			var catalog map[string]json.RawMessage
			var definitions []map[string]json.RawMessage
			var policy map[string]json.RawMessage
			var inputs []map[string]json.RawMessage
			if json.Unmarshal(base, &catalog) != nil || json.Unmarshal(catalog["minigames"], &definitions) != nil ||
				json.Unmarshal(definitions[0]["scaling"], &policy) != nil || json.Unmarshal(policy["scaling_inputs"], &inputs) != nil {
				t.Fatal("invalid canonical catalog")
			}
			for key, value := range row.Changes {
				inputs[0][key] = value
			}
			policy["scaling_inputs"] = scalingTestJSON(t, inputs)
			policyBytes := scalingTestJSON(t, policy)
			loadedPolicy, policyErr := LoadScalingPolicy(policyBytes, true)
			definitions[0]["scaling"] = policyBytes
			catalog["minigames"] = scalingTestJSON(t, definitions)
			loadedCatalog, catalogErr := LoadCatalog(scalingTestJSON(t, catalog))
			if row.Valid {
				if policyErr != nil || catalogErr != nil || loadedPolicy == nil || loadedCatalog == nil {
					t.Fatalf("valid source refused: policy=%v catalog=%v", policyErr, catalogErr)
				}
			} else if !errors.Is(policyErr, ErrInvalidScalingPolicy) || !errors.Is(catalogErr, ErrInvalidCatalog) || loadedPolicy != nil || loadedCatalog != nil {
				t.Fatalf("invalid source admitted: policy=%v catalog=%v", policyErr, catalogErr)
			}
		})
	}
}

const scalingFixture = `{"schema_version":1,"scaling_inputs":[
{"destination":"challenge","destination_class":"power","source_kind":"literal","source_ref":"7","op":"mul","operand":3,"clamp_min":0,"clamp_max":100},
{"destination":"era","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"add","operand":1,"clamp_min":1,"clamp_max":9},
{"destination":"roster","destination_class":"breadth","source_kind":"purchased_generator_count","source_ref":"generator.beige_tower","op":"floordiv","operand":2,"clamp_min":0,"clamp_max":20},
{"destination":"experience","destination_class":"presentation","source_kind":"founder_carry_counter","source_ref":"age_ms","op":"identity","operand":0,"clamp_min":0,"clamp_max":1000},
{"destination":"offline_quality","destination_class":"presentation","source_kind":"attended_quality_grade","source_ref":"combat.duel","op":"add","operand":-100,"clamp_min":400000,"clamp_max":1000000}]}`

func TestScalingPolicyResolvesClosedSourcesAndNegativeFloor(t *testing.T) {
	policy, err := LoadScalingPolicy([]byte(scalingFixture), false)
	if err != nil {
		t.Fatal(err)
	}
	values, err := policy.Resolve(ScalingContext{Tier: 3,
		PurchasedGeneratorCounts: map[string]int64{"generator.beige_tower": 9},
		FounderCarryCounters:     map[string]int64{"age_ms": 5000},
		AttendedQualityGrades:    map[string]int64{"combat.duel": 750000}})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{"challenge": 21, "era": 4, "roster": 4, "experience": 1000, "offline_quality": 749900}
	for key, expected := range want {
		if values[key] != expected {
			t.Fatalf("%s=%d want=%d values=%v", key, values[key], expected, values)
		}
	}
	negative := `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"presentation","source_kind":"literal","source_ref":"-5","op":"floordiv","operand":2,"clamp_min":-10,"clamp_max":10}]}`
	policy, err = LoadScalingPolicy([]byte(negative), false)
	if err != nil {
		t.Fatal(err)
	}
	values, err = policy.Resolve(ScalingContext{PurchasedGeneratorCounts: map[string]int64{}, FounderCarryCounters: map[string]int64{}, AttendedQualityGrades: map[string]int64{}})
	if err != nil || values["x"] != -3 {
		t.Fatalf("negative floor=%v err=%v", values, err)
	}
}

func TestScalingPolicyRejectsRankedPowerAndInventedGrammar(t *testing.T) {
	if _, err := LoadScalingPolicy([]byte(scalingFixture), true); !errors.Is(err, ErrInvalidScalingPolicy) {
		t.Fatalf("ranked power err=%v", err)
	}
	cases := map[string]string{
		"unknown key":         `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"identity","operand":0,"clamp_min":0,"clamp_max":9,"formula":"tier"}]}`,
		"duplicate target":    `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"identity","operand":0,"clamp_min":0,"clamp_max":9},{"destination":"x","destination_class":"presentation","source_kind":"literal","source_ref":"1","op":"identity","operand":0,"clamp_min":0,"clamp_max":9}]}`,
		"bad literal":         `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"literal","source_ref":"01","op":"identity","operand":0,"clamp_min":0,"clamp_max":9}]}`,
		"identity operand":    `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"identity","operand":1,"clamp_min":0,"clamp_max":9}]}`,
		"negative divisor":    `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"floordiv","operand":-2,"clamp_min":0,"clamp_max":9}]}`,
		"unknown carry path":  `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"founder_carry_counter","source_ref":"soul","op":"identity","operand":0,"clamp_min":0,"clamp_max":9}]}`,
		"duplicate operation": `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"add","op":"mul","operand":1,"clamp_min":0,"clamp_max":9}]}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadScalingPolicy([]byte(raw), false); !errors.Is(err, ErrInvalidScalingPolicy) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestScalingPolicyUsesExactMathBeforeClamp(t *testing.T) {
	raw := `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"literal","source_ref":"9007199254740991","op":"mul","operand":9007199254740991,"clamp_min":0,"clamp_max":9007199254740991}]}`
	policy, err := LoadScalingPolicy([]byte(raw), false)
	if err != nil {
		t.Fatal(err)
	}
	values, err := policy.Resolve(ScalingContext{PurchasedGeneratorCounts: map[string]int64{}, FounderCarryCounters: map[string]int64{}, AttendedQualityGrades: map[string]int64{}})
	if err != nil || values["x"] != 9007199254740991 {
		t.Fatalf("exact multiply/clamp=%v err=%v", values, err)
	}
}
