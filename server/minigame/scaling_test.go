package minigame

import (
	"errors"
	"testing"
)

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
		"unknown key":        `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"identity","operand":0,"clamp_min":0,"clamp_max":9,"formula":"tier"}]}`,
		"duplicate target":   `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"identity","operand":0,"clamp_min":0,"clamp_max":9},{"destination":"x","destination_class":"presentation","source_kind":"literal","source_ref":"1","op":"identity","operand":0,"clamp_min":0,"clamp_max":9}]}`,
		"bad literal":        `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"literal","source_ref":"01","op":"identity","operand":0,"clamp_min":0,"clamp_max":9}]}`,
		"identity operand":   `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"identity","operand":1,"clamp_min":0,"clamp_max":9}]}`,
		"negative divisor":   `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"tier","source_ref":"tier","op":"floordiv","operand":-2,"clamp_min":0,"clamp_max":9}]}`,
		"unknown carry path": `{"schema_version":1,"scaling_inputs":[{"destination":"x","destination_class":"breadth","source_kind":"founder_carry_counter","source_ref":"soul","op":"identity","operand":0,"clamp_min":0,"clamp_max":9}]}`,
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
