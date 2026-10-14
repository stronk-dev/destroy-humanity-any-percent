package publicapi

import (
	"fmt"
	"testing"
)

func compatibilityArrayOf(items *Schema) *Schema { return &Schema{Kind: SchemaArray, Items: items} }

func TestCompatibilityPinRejectsArrayArmGrowthInvalidatingOldBytes(t *testing.T) {
	for _, shape := range []string{"direct", "array", "union"} {
		for _, domain := range []string{"enum", "integer", "object"} {
			t.Run(shape+"/"+domain, func(t *testing.T) {
				left := &Schema{Kind: SchemaString, Enum: []string{"a"}}
				right := &Schema{Kind: SchemaString, Enum: []string{"b"}}
				grown := &Schema{Kind: SchemaString, Enum: []string{"a", "b"}}
				value := `["b"]`
				switch domain {
				case "integer":
					left, right, grown = objectGrowthInteger(0, 1), objectGrowthInteger(2, 3), objectGrowthInteger(0, 2)
					value = `[2]`
				case "object":
					left = objectGrowthFields()
					right = objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString}, Required: true})
					grown = objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString}})
					value = `[{"value":"b"}]`
				}
				prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{compatibilityArrayOf(left), compatibilityArrayOf(right)}}, shape)
				pin, err := CanonicalOperationPins(prior)
				if err != nil {
					t.Fatal(err)
				}
				next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{compatibilityArrayOf(grown), compatibilityArrayOf(right)}}, shape)
				if shape == "array" {
					value = "[" + value + "]"
				}
				body := []byte(fmt.Sprintf(`{"payload":%s}`, value))
				if err := prior.ValidateResponse("get_payload", 200, body); err != nil {
					t.Fatalf("old array witness is not admitted: %v", err)
				}
				if err := next.ValidateResponse("get_payload", 200, body); err == nil {
					t.Fatal("array growth did not invalidate the old reply")
				}
				if err := CheckCompatibilityPin(pin, next); err == nil {
					t.Fatal("array-arm growth breaks a previously exclusive reply")
				}
			})
		}
	}
}

func TestCompatibilityPinArrayGrowthCannotHideMixedElementLoss(t *testing.T) {
	array := func(values ...string) *Schema { return compatibilityArrayOf(&Schema{Kind: SchemaString, Enum: values}) }
	prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{array("a"), array("a", "b"), array("b")}}, "direct")
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{array("a", "b"), array("a", "b"), array("b")}}, "direct")
	for _, value := range []string{`[]`, `["a"]`, `["b"]`} {
		if prior.ValidateResponse("get_payload", 200, []byte(`{"payload":`+value+`}`)) == nil {
			t.Fatal("singleton/empty premise must be old-invalid")
		}
	}
	body := []byte(`{"payload":["a","b"]}`)
	if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) == nil {
		t.Fatal("mixed-element witness did not lose its old exclusive match")
	}
	if CheckCompatibilityPin(pin, next) == nil {
		t.Fatal("singleton coverage concealed a valid old mixed array")
	}
}

func TestCompatibilityPinPreservesSafeArrayGrowth(t *testing.T) {
	for _, shape := range []string{"direct", "array", "union"} {
		for _, legacyOverlap := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/legacy=%t", shape, legacyOverlap), func(t *testing.T) {
				left, right := []string{"a"}, []string{"b"}
				grownLeft, grownRight := []string{"a", "c"}, []string{"b", "c"}
				if legacyOverlap {
					left, right = []string{"a", "x"}, []string{"b", "x"}
					grownLeft, grownRight = []string{"a", "c", "x"}, []string{"b", "c", "x"}
				}
				union := func(a, b []string) *Schema {
					return &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
						compatibilityArrayOf(&Schema{Kind: SchemaString, Enum: a}), compatibilityArrayOf(&Schema{Kind: SchemaString, Enum: b}),
					}}
				}
				prior := compatibilityPayloadRegistry(t, union(left, right), shape)
				pin, err := CanonicalOperationPins(prior)
				if err != nil {
					t.Fatal(err)
				}
				for _, payload := range []*Schema{union(left, right), union(grownLeft, grownRight)} {
					next := compatibilityPayloadRegistry(t, payload, shape)
					for _, value := range []string{`["a"]`, `["b"]`, `["a","a"]`, `["b","b"]`} {
						if shape == "array" {
							value = "[" + value + "]"
						}
						body := []byte(`{"payload":` + value + `}`)
						if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) != nil {
							t.Fatal("safe array growth lost an old admitted reply")
						}
					}
					if err := CheckCompatibilityPin(pin, next); err != nil {
						t.Fatalf("unchanged/safe array growth refused: %v", err)
					}
				}
			})
		}
	}
}

func TestCompatibilityPinArrayGrowthFollowsEachSidesNamedDefinitions(t *testing.T) {
	definitions := func(left []string) []NamedSchema {
		return []NamedSchema{
			{Name: "LeftAlias", Schema: &Schema{Kind: SchemaRef, Ref: "LeftArray"}},
			{Name: "LeftArray", Schema: compatibilityArrayOf(&Schema{Kind: SchemaRef, Ref: "LeftItems"})},
			{Name: "LeftItems", Schema: &Schema{Kind: SchemaString, Enum: left}},
			{Name: "RightArray", Schema: compatibilityArrayOf(&Schema{Kind: SchemaString, Enum: []string{"b"}})},
		}
	}
	payload := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{{Kind: SchemaRef, Ref: "LeftAlias"}, {Kind: SchemaRef, Ref: "RightArray"}}}
	prior := compatibilityPayloadRegistry(t, payload, "direct", definitions([]string{"a"})...)
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	for _, added := range []string{"b", "c"} {
		next := compatibilityPayloadRegistry(t, payload, "direct", definitions([]string{"a", added})...)
		body := []byte(`{"payload":["b"]}`)
		refuse := added == "b"
		if prior.ValidateResponse("get_payload", 200, body) != nil || (next.ValidateResponse("get_payload", 200, body) != nil) != refuse || (CheckCompatibilityPin(pin, next) != nil) != refuse {
			t.Fatalf("named-array growth does not preserve actual exclusive matches: %s", added)
		}
	}
}

func TestCompatibilityPinArrayGrowthRetainsThirdArmCoveredOverlap(t *testing.T) {
	array := func(values ...string) *Schema { return compatibilityArrayOf(&Schema{Kind: SchemaString, Enum: values}) }
	prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{array("a"), array("b"), array("b")}}, "direct")
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{array("a", "b"), array("b"), array("b")}}, "direct")
	for _, value := range []string{`[]`, `["b"]`} {
		if prior.ValidateResponse("get_payload", 200, []byte(`{"payload":`+value+`}`)) == nil {
			t.Fatal("third-arm overlap was not already invalid")
		}
	}
	for _, value := range []string{`["a"]`, `["a","a"]`} {
		body := []byte(`{"payload":` + value + `}`)
		if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) != nil {
			t.Fatal("third-arm fixture lost an old valid array")
		}
	}
	if err := CheckCompatibilityPin(pin, next); err != nil {
		t.Fatalf("third-old-arm-covered array intersection refused: %v", err)
	}
}

func TestCompatibilityArrayGrowthAgreesWithFiniteRuntimeLanguage(t *testing.T) {
	array := func(mask int) *Schema {
		values := []string{}
		for index, value := range []string{"a", "b"} {
			if mask&(1<<index) != 0 {
				values = append(values, value)
			}
		}
		return compatibilityArrayOf(&Schema{Kind: SchemaString, Enum: values})
	}
	// Membership depends on the element set, not order or multiplicity. These
	// cover every set over the two admitted literals, plus duplicate/order controls.
	values := []string{`[]`, `["a"]`, `["b"]`, `["a","a"]`, `["a","b"]`, `["b","a"]`, `["b","b"]`}
	for a := 1; a <= 3; a++ {
		for b := 1; b <= 3; b++ {
			for c := 1; c <= 3; c++ {
				masks := []int{a, b, c}
				oldArms := []*Schema{array(a), array(b), array(c)}
				prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: oldArms}, "direct")
				pin, err := CanonicalOperationPins(prior)
				if err != nil {
					t.Fatal(err)
				}
				for grown := 0; grown < 3; grown++ {
					for added := 1; added <= 2; added++ {
						nextArms := append([]*Schema(nil), oldArms...)
						nextArms[grown] = array(masks[grown] | added)
						next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: nextArms}, "direct")
						lost := false
						for _, value := range values {
							body := []byte(`{"payload":` + value + `}`)
							if prior.ValidateResponse("get_payload", 200, body) == nil && next.ValidateResponse("get_payload", 200, body) != nil {
								lost = true
							}
						}
						if (CheckCompatibilityPin(pin, next) != nil) != lost {
							t.Fatalf("array proof differs from runtime: masks=%v grown=%d added=%d lost=%t", masks, grown, added, lost)
						}
					}
				}
			}
		}
	}
}
