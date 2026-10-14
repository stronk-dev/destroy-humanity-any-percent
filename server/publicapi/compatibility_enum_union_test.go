package publicapi

import (
	"fmt"
	"testing"
)

func TestCompatibilityPinRejectsExistingArmEnumGrowthInvalidatingOldBytes(t *testing.T) {
	for _, shape := range []string{"direct", "array", "union"} {
		for _, grow := range []int{0, 1} {
			t.Run(fmt.Sprintf("%s/arm%d", shape, grow), func(t *testing.T) {
				old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
					{Kind: SchemaString, Enum: []string{"a"}}, {Kind: SchemaString, Enum: []string{"b"}},
				}}
				prior := compatibilityPayloadRegistry(t, old, shape)
				pin, err := CanonicalOperationPins(prior)
				if err != nil {
					t.Fatal(err)
				}
				nextArms := []*Schema{{Kind: SchemaString, Enum: []string{"a"}}, {Kind: SchemaString, Enum: []string{"b"}}}
				nextArms[grow].Enum = []string{"a", "b"}
				next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: nextArms}, shape)
				value := `"b"`
				if grow == 1 {
					value = `"a"`
				}
				if shape == "array" {
					value = "[" + value + "]"
				}
				body := []byte(fmt.Sprintf(`{"payload":%s}`, value))
				if err := prior.ValidateResponse("get_payload", 200, body); err != nil {
					t.Fatal(err)
				}
				if err := next.ValidateResponse("get_payload", 200, body); err == nil {
					t.Fatal("enum fixture did not invalidate old bytes")
				}
				if err := CheckCompatibilityPin(pin, next); err == nil {
					t.Fatal("existing-arm widening invalidated a previously valid reply")
				}
			})
		}
	}
}

func TestCompatibilityPinAllowsExistingArmSafeEnumGrowth(t *testing.T) {
	for _, legacyOverlap := range []bool{false, true} {
		t.Run(fmt.Sprint(legacyOverlap), func(t *testing.T) {
			left, right := []string{"a"}, []string{"b"}
			if legacyOverlap {
				left, right = []string{"a", "x"}, []string{"b", "x"}
			}
			old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{{Kind: SchemaString, Enum: left}, {Kind: SchemaString, Enum: right}}}
			prior := compatibilityPayloadRegistry(t, old, "direct")
			pin, err := CanonicalOperationPins(prior)
			if err != nil {
				t.Fatal(err)
			}
			newLeft, newRight := []string{"a", "c"}, []string{"b", "c"}
			if legacyOverlap {
				newLeft, newRight = []string{"a", "c", "x"}, []string{"b", "c", "x"}
			}
			next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
				{Kind: SchemaString, Enum: newLeft}, {Kind: SchemaString, Enum: newRight},
			}}, "direct")
			for _, value := range []string{"a", "b"} {
				body := []byte(fmt.Sprintf(`{"payload":%q}`, value))
				if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) != nil {
					t.Fatal("safe fixture lost an old value")
				}
			}
			// c is ambiguous only in the new schema, and x was already invalid:
			// neither was part of the previous admitted response population.
			if err := CheckCompatibilityPin(pin, next); err != nil {
				t.Fatalf("safe enum growth was narrowed: %v", err)
			}
		})
	}
}

func TestCompatibilityPinEnumGrowthCannotStealOldUnboundedString(t *testing.T) {
	old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
		{Kind: SchemaString}, {Kind: SchemaString, Enum: []string{"a"}},
	}}
	prior := compatibilityPayloadRegistry(t, old, "direct")
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
		{Kind: SchemaString}, {Kind: SchemaString, Enum: []string{"a", "c"}},
	}}, "direct")
	body := []byte(`{"payload":"c"}`)
	if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) == nil {
		t.Fatal("new-only enum member did not exercise a formerly valid unbounded string")
	}
	if err := CheckCompatibilityPin(pin, next); err == nil {
		t.Fatal("enum grew across an old unbounded string domain")
	}
}

func TestCompatibilityPinExistingArmEnumGrowthThroughNamedDefinitions(t *testing.T) {
	definitions := func(left []string) []NamedSchema {
		return []NamedSchema{
			{Name: "LeftAlias", Schema: &Schema{Kind: SchemaRef, Ref: "LeftValue"}},
			{Name: "LeftValue", Schema: &Schema{Kind: SchemaString, Enum: left}},
			{Name: "RightValue", Schema: &Schema{Kind: SchemaString, Enum: []string{"b"}}},
		}
	}
	payload := func() *Schema {
		return &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
			{Kind: SchemaRef, Ref: "LeftAlias"}, {Kind: SchemaRef, Ref: "RightValue"},
		}}
	}
	prior := compatibilityPayloadRegistry(t, payload(), "direct", definitions([]string{"a"})...)
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	for _, added := range []string{"b", "c"} {
		next := compatibilityPayloadRegistry(t, payload(), "direct", definitions([]string{"a", added})...)
		body := []byte(`{"payload":"b"}`)
		wantRefusal := added == "b"
		if prior.ValidateResponse("get_payload", 200, body) != nil || (next.ValidateResponse("get_payload", 200, body) != nil) != wantRefusal {
			t.Fatal("named enum fixture did not preserve/invalidate the old value")
		}
		if err := CheckCompatibilityPin(pin, next); (err != nil) != wantRefusal {
			t.Fatalf("named enum growth: err=%v wantRefusal=%t", err, wantRefusal)
		}
	}
}

func TestCompatibilityPinStringEnumGrowthUsesActualFormatDomain(t *testing.T) {
	const initial = "01986666-1111-7111-8111-111111111111"
	for _, added := range []string{"01986666-2222-7222-8222-222222222222", "not-a-uuid"} {
		old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
			{Kind: SchemaString}, {Kind: SchemaString, Format: "uuid", Enum: []string{initial}},
		}}
		prior := compatibilityPayloadRegistry(t, old, "direct")
		pin, err := CanonicalOperationPins(prior)
		if err != nil {
			t.Fatal(err)
		}
		next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
			{Kind: SchemaString}, {Kind: SchemaString, Format: "uuid", Enum: []string{initial, added}},
		}}, "direct")
		body := []byte(fmt.Sprintf(`{"payload":%q}`, added))
		wantRefusal := added != "not-a-uuid"
		if prior.ValidateResponse("get_payload", 200, body) != nil || (next.ValidateResponse("get_payload", 200, body) != nil) != wantRefusal {
			t.Fatal("format-domain fixture did not preserve/invalidate the actual old string")
		}
		if err := CheckCompatibilityPin(pin, next); (err != nil) != wantRefusal {
			t.Fatalf("format-aware enum growth: err=%v wantRefusal=%t", err, wantRefusal)
		}
	}
}
