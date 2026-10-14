package publicapi

import (
	"fmt"
	"testing"
)

func integerUnionSchema(bounds [][2]int64) *Schema {
	result := &Schema{Kind: SchemaOneOf}
	for _, bound := range bounds {
		minimum, maximum := bound[0], bound[1]
		result.Alternates = append(result.Alternates, &Schema{Kind: SchemaInteger, Minimum: &minimum, Maximum: &maximum})
	}
	return result
}

func TestCompatibilityPinRejectsExistingIntegerArmOverlap(t *testing.T) {
	const floor, ceiling = int64(-1 << 63), int64(1<<63 - 1)
	const wide = int64(1<<53 + 1)
	for _, row := range []struct {
		name      string
		old, next [][2]int64
		witness   int64
	}{
		{"maximum", [][2]int64{{0, 2}, {4, 6}}, [][2]int64{{0, 4}, {4, 6}}, 4},
		{"minimum", [][2]int64{{0, 2}, {4, 6}}, [][2]int64{{0, 2}, {2, 6}}, 2},
		{"above-float-precision", [][2]int64{{wide - 1, wide}, {wide + 1, wide + 2}}, [][2]int64{{wide - 1, wide + 1}, {wide + 1, wide + 2}}, wide + 1},
		{"int64-floor", [][2]int64{{floor, floor + 1}, {floor + 2, floor + 3}}, [][2]int64{{floor, floor + 2}, {floor + 2, floor + 3}}, floor + 2},
		{"int64-ceiling", [][2]int64{{ceiling - 3, ceiling - 2}, {ceiling - 1, ceiling}}, [][2]int64{{ceiling - 3, ceiling - 1}, {ceiling - 1, ceiling}}, ceiling - 1},
	} {
		for _, shape := range []string{"direct", "array", "union"} {
			t.Run(row.name+"/"+shape, func(t *testing.T) {
				prior := compatibilityPayloadRegistry(t, integerUnionSchema(row.old), shape)
				pin, err := CanonicalOperationPins(prior)
				if err != nil {
					t.Fatal(err)
				}
				next := compatibilityPayloadRegistry(t, integerUnionSchema(row.next), shape)
				value := fmt.Sprint(row.witness)
				if shape == "array" {
					value = "[" + value + "]"
				}
				body := []byte(fmt.Sprintf(`{"payload":%s}`, value))
				if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) == nil {
					t.Fatal("integer fixture did not invalidate formerly valid bytes")
				}
				if err := CheckCompatibilityPin(pin, next); err == nil {
					t.Fatal("integer interval growth broke an old response")
				}
			})
		}
	}
}

func TestCompatibilityPinAllowsSafeIntegerArmGrowth(t *testing.T) {
	for _, row := range []struct {
		name      string
		old, next [][2]int64
	}{
		{"new-only-overlap", [][2]int64{{0, 1}, {4, 5}}, [][2]int64{{0, 3}, {2, 5}}},
		{"existing-invalid-overlap", [][2]int64{{0, 3}, {2, 5}}, [][2]int64{{-1, 3}, {2, 6}}},
		{"both-extremes", [][2]int64{{-1 << 63, -10}, {10, 1<<63 - 1}}, [][2]int64{{-1 << 63, -5}, {5, 1<<63 - 1}}},
	} {
		t.Run(row.name, func(t *testing.T) {
			prior := compatibilityPayloadRegistry(t, integerUnionSchema(row.old), "direct")
			pin, err := CanonicalOperationPins(prior)
			if err != nil {
				t.Fatal(err)
			}
			next := compatibilityPayloadRegistry(t, integerUnionSchema(row.next), "direct")
			if err := CheckCompatibilityPin(pin, next); err != nil {
				t.Fatalf("safe integer growth was narrowed: %v", err)
			}
		})
	}
}

func TestCompatibilityIntegerRegionsAgreeWithCompleteSmallRuntimePopulation(t *testing.T) {
	prior := compatibilityPayloadRegistry(t, integerUnionSchema([][2]int64{{-2, -1}, {1, 2}}), "direct")
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	for leftMaximum := int64(-1); leftMaximum <= 2; leftMaximum++ {
		for rightMinimum := int64(-2); rightMinimum <= 1; rightMinimum++ {
			next := compatibilityPayloadRegistry(t, integerUnionSchema([][2]int64{{-2, leftMaximum}, {rightMinimum, 2}}), "direct")
			wantRefusal := false
			// This exhausts the old population, not just selected witnesses:
			// its two intervals admit no integer outside [-2,2].
			for value := int64(-2); value <= 2; value++ {
				body := []byte(fmt.Sprintf(`{"payload":%d}`, value))
				if prior.ValidateResponse("get_payload", 200, body) == nil && next.ValidateResponse("get_payload", 200, body) != nil {
					wantRefusal = true
				}
			}
			if err := CheckCompatibilityPin(pin, next); (err != nil) != wantRefusal {
				t.Fatalf("regions/runtime disagree at leftMax=%d rightMin=%d: err=%v wantRefusal=%t", leftMaximum, rightMinimum, err, wantRefusal)
			}
		}
	}
}

func TestCompatibilityIntegerGrowthPreservesUnboundedAndNamedDomains(t *testing.T) {
	for _, referenced := range []bool{false, true} {
		t.Run(fmt.Sprint(referenced), func(t *testing.T) {
			minimum, maximum, grownMinimum := int64(2), int64(3), int64(1)
			oldBound := &Schema{Kind: SchemaInteger, Minimum: &minimum, Maximum: &maximum}
			newBound := &Schema{Kind: SchemaInteger, Minimum: &grownMinimum, Maximum: &maximum}
			oldArm, nextArm := oldBound, newBound
			var oldDefinitions, nextDefinitions []NamedSchema
			if referenced {
				oldArm, nextArm = &Schema{Kind: SchemaRef, Ref: "RangeAlias"}, &Schema{Kind: SchemaRef, Ref: "RangeAlias"}
				oldDefinitions = []NamedSchema{{Name: "RangeAlias", Schema: &Schema{Kind: SchemaRef, Ref: "RangeValue"}}, {Name: "RangeValue", Schema: oldBound}}
				nextDefinitions = []NamedSchema{{Name: "RangeAlias", Schema: &Schema{Kind: SchemaRef, Ref: "RangeValue"}}, {Name: "RangeValue", Schema: newBound}}
			}
			prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{{Kind: SchemaInteger}, oldArm}}, "direct", oldDefinitions...)
			pin, err := CanonicalOperationPins(prior)
			if err != nil {
				t.Fatal(err)
			}
			next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{{Kind: SchemaInteger}, nextArm}}, "direct", nextDefinitions...)
			body := []byte(`{"payload":1}`)
			if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) == nil {
				t.Fatal("unbounded/reference fixture did not invalidate old bytes")
			}
			if err := CheckCompatibilityPin(pin, next); err == nil {
				t.Fatal("new minimum stole a formerly valid unbounded integer")
			}
		})
	}
}
