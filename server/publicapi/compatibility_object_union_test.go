package publicapi

import (
	"fmt"
	"testing"
)

func objectGrowthFields(fields ...Field) *Schema { return &Schema{Kind: SchemaObject, Fields: fields} }
func objectGrowthInteger(minimum, maximum int64) *Schema {
	return &Schema{Kind: SchemaInteger, Minimum: &minimum, Maximum: &maximum}
}

func TestCompatibilityPinRejectsObjectGrowthInvalidatingOldBytes(t *testing.T) {
	for _, shape := range []string{"direct", "array", "union"} {
		for _, change := range []string{"optional_property", "required_demotion", "enum", "integer"} {
			t.Run(shape+"/"+change, func(t *testing.T) {
				field := &Schema{Kind: SchemaString}
				oldLeft := objectGrowthFields()
				oldRight := objectGrowthFields(Field{Name: "value", Schema: field, Required: true})
				nextLeft := objectGrowthFields(Field{Name: "value", Schema: field})
				value := `{"value":"b"}`
				switch change {
				case "required_demotion":
					oldLeft, oldRight = oldRight, objectGrowthFields()
					value = `{}`
				case "enum":
					oldLeft = objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString, Enum: []string{"a"}}, Required: true})
					oldRight = objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString, Enum: []string{"b"}}, Required: true})
					nextLeft = objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString, Enum: []string{"a", "b"}}, Required: true})
				case "integer":
					oldLeft = objectGrowthFields(Field{Name: "value", Schema: objectGrowthInteger(0, 1), Required: true})
					oldRight = objectGrowthFields(Field{Name: "value", Schema: objectGrowthInteger(2, 3), Required: true})
					nextLeft = objectGrowthFields(Field{Name: "value", Schema: objectGrowthInteger(0, 2), Required: true})
					value = `{"value":2}`
				}
				prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{oldLeft, oldRight}}, shape)
				pin, err := CanonicalOperationPins(prior)
				if err != nil {
					t.Fatal(err)
				}
				next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{nextLeft, oldRight}}, shape)
				if shape == "array" {
					value = "[" + value + "]"
				}
				body := []byte(fmt.Sprintf(`{"payload":%s}`, value))
				if err := prior.ValidateResponse("get_payload", 200, body); err != nil {
					t.Fatal(err)
				}
				if err := next.ValidateResponse("get_payload", 200, body); err == nil {
					t.Fatal("growth did not invalidate actual old bytes")
				}
				if err := CheckCompatibilityPin(pin, next); err == nil {
					t.Fatal("object growth broke an old exclusive response")
				}
			})
		}
	}
}

func TestCompatibilityPinPreservesSafeObjectGrowthAndLegacyOverlap(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprint(legacy), func(t *testing.T) {
			left := objectGrowthFields(Field{Name: "kind", Schema: &Schema{Kind: SchemaString, Enum: []string{"a"}}, Required: true})
			right := objectGrowthFields(Field{Name: "kind", Schema: &Schema{Kind: SchemaString, Enum: []string{"b"}}, Required: true})
			nextLeft := objectGrowthFields(Field{Name: "kind", Schema: &Schema{Kind: SchemaString, Enum: []string{"a", "c"}}, Required: true}, Field{Name: "note", Schema: &Schema{Kind: SchemaString}})
			if legacy {
				left.Fields[0].Schema.Enum = []string{"a", "x"}
				right.Fields[0].Schema.Enum = []string{"b", "x"}
				nextLeft.Fields[0].Schema.Enum = []string{"a", "c", "x"}
			}
			prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{left, right}}, "direct")
			pin, err := CanonicalOperationPins(prior)
			if err != nil {
				t.Fatal(err)
			}
			next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{nextLeft, right}}, "direct")
			for _, kind := range []string{"a", "b"} {
				body := []byte(fmt.Sprintf(`{"payload":{"kind":%q}}`, kind))
				if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) != nil {
					t.Fatal("safe growth lost old bytes")
				}
			}
			if err := CheckCompatibilityPin(pin, next); err != nil {
				t.Fatalf("safe object growth rejected: %v", err)
			}
		})
	}
}

func TestCompatibilityObjectGrowthAgreesWithCompleteFiniteRuntimePopulation(t *testing.T) {
	rows := []*Schema{objectGrowthFields()}
	for _, required := range []bool{false, true} {
		for _, field := range []*Schema{{Kind: SchemaString, Enum: []string{"a"}}, {Kind: SchemaString, Enum: []string{"b"}}, objectGrowthInteger(0, 0), objectGrowthInteger(1, 1), {Kind: SchemaBoolean}} {
			rows = append(rows, objectGrowthFields(Field{Name: "value", Schema: field, Required: required}))
		}
	}
	// Every old field domain is finite, and old objects allow only value.
	// These seven objects exhaust every byte-value class the old schemas admit.
	values := []string{`{}`, `{"value":"a"}`, `{"value":"b"}`, `{"value":0}`, `{"value":1}`, `{"value":true}`, `{"value":false}`}
	for leftIndex, left := range rows {
		for rightIndex, right := range rows {
			prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{left, right}}, "direct")
			pin, err := CanonicalOperationPins(prior)
			if err != nil {
				t.Fatal(err)
			}
			for _, demote := range []bool{false, true} {
				fields := append([]Field(nil), left.Fields...)
				if demote {
					for index := range fields {
						fields[index].Required = false
					}
				} else {
					fields = append([]Field{{Name: "note", Schema: &Schema{Kind: SchemaString}}}, fields...)
				}
				next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{objectGrowthFields(fields...), right}}, "direct")
				lost := false
				for _, value := range values {
					body := []byte(`{"payload":` + value + `}`)
					if prior.ValidateResponse("get_payload", 200, body) == nil && next.ValidateResponse("get_payload", 200, body) != nil {
						lost = true
					}
				}
				if rejected := CheckCompatibilityPin(pin, next) != nil; rejected != lost {
					t.Fatalf("pair %d/%d demote=%t rejected=%t lost=%t", leftIndex, rightIndex, demote, rejected, lost)
				}
			}
		}
	}
}

func TestCompatibilityObjectGrowthUsesOwnDefinitionsAndNestedArms(t *testing.T) {
	for _, wrapper := range []string{"named", "named_value", "nested_union", "nested_object"} {
		for _, safe := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/safe=%t", wrapper, safe), func(t *testing.T) {
				left, right := objectGrowthFields(), objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString}, Required: true})
				name := "value"
				if safe {
					name = "note"
				}
				nextLeft := objectGrowthFields(Field{Name: name, Schema: &Schema{Kind: SchemaString}})
				oldArms, nextArms := []*Schema{left, right}, []*Schema{nextLeft, right}
				var oldDefinitions, nextDefinitions []NamedSchema
				switch wrapper {
				case "named_value":
					left = objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaRef, Ref: "LeftValue"}, Required: true})
					right = objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaRef, Ref: "RightValue"}, Required: true})
					oldArms, nextArms = []*Schema{left, right}, []*Schema{left, right}
					oldDefinitions = []NamedSchema{{Name: "LeftValue", Schema: &Schema{Kind: SchemaString, Enum: []string{"a"}}}, {Name: "RightValue", Schema: &Schema{Kind: SchemaString, Enum: []string{"b"}}}}
					added := "b"
					if safe {
						added = "c"
					}
					nextDefinitions = []NamedSchema{{Name: "LeftValue", Schema: &Schema{Kind: SchemaString, Enum: []string{"a", added}}}, oldDefinitions[1]}
				case "named":
					oldDefinitions = []NamedSchema{{Name: "Left", Schema: left}, {Name: "Right", Schema: right}}
					nextDefinitions = []NamedSchema{{Name: "Left", Schema: nextLeft}, {Name: "Right", Schema: right}}
					oldArms = []*Schema{{Kind: SchemaRef, Ref: "Left"}, {Kind: SchemaRef, Ref: "Right"}}
					nextArms = oldArms
				case "nested_union":
					oldArms[0] = &Schema{Kind: SchemaOneOf, Alternates: []*Schema{left, {Kind: SchemaBoolean}}}
					nextArms[0] = &Schema{Kind: SchemaOneOf, Alternates: []*Schema{nextLeft, {Kind: SchemaBoolean}}}
				case "nested_object":
					for index, row := range oldArms {
						oldArms[index] = objectGrowthFields(Field{Name: "data", Schema: row, Required: true})
					}
					for index, row := range nextArms {
						nextArms[index] = objectGrowthFields(Field{Name: "data", Schema: row, Required: true})
					}
				}
				prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: oldArms}, "direct", oldDefinitions...)
				pin, err := CanonicalOperationPins(prior)
				if err != nil {
					t.Fatal(err)
				}
				next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: nextArms}, "direct", nextDefinitions...)
				value := `{"value":"b"}`
				if wrapper == "nested_object" {
					value = `{"data":` + value + `}`
				}
				body := []byte(`{"payload":` + value + `}`)
				if prior.ValidateResponse("get_payload", 200, body) != nil || (next.ValidateResponse("get_payload", 200, body) == nil) != safe {
					t.Fatal("wrapper fixture did not discriminate actual old bytes")
				}
				if err := CheckCompatibilityPin(pin, next); (err == nil) != safe {
					t.Fatalf("wrapper compatibility: %v", err)
				}
			})
		}
	}
}

func TestCompatibilityObjectGrowthPreservesExactIntegerIntersections(t *testing.T) {
	for _, boundary := range []int64{compatibilityIntegerFloor, 1 << 53, compatibilityIntegerCeiling - 1} {
		left := objectGrowthFields(Field{Name: "value", Schema: objectGrowthInteger(boundary, boundary), Required: true})
		right := objectGrowthFields(Field{Name: "value", Schema: objectGrowthInteger(boundary+1, boundary+1), Required: true})
		prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{left, right}}, "direct")
		pin, err := CanonicalOperationPins(prior)
		if err != nil {
			t.Fatal(err)
		}
		nextLeft := objectGrowthFields(Field{Name: "value", Schema: objectGrowthInteger(boundary, boundary+1), Required: true})
		next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{nextLeft, right}}, "direct")
		body := []byte(fmt.Sprintf(`{"payload":{"value":%d}}`, boundary+1))
		if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) == nil {
			t.Fatal("exact boundary did not invalidate actual old bytes")
		}
		if err := CheckCompatibilityPin(pin, next); err == nil {
			t.Fatalf("exact boundary %d was rounded away", boundary)
		}
	}
}

func TestCompatibilityObjectGrowthMayCoverAlreadyInvalidThirdArmOverlap(t *testing.T) {
	left := objectGrowthFields()
	right := objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString, Enum: []string{"b"}}, Required: true})
	third := objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString, Enum: []string{"b", "c"}}, Required: true})
	prior := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{left, right, third}}, "direct")
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	nextLeft := objectGrowthFields(Field{Name: "value", Schema: &Schema{Kind: SchemaString, Enum: []string{"b"}}})
	next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf, Alternates: []*Schema{nextLeft, right, third}}, "direct")
	for _, value := range []string{`{}`, `{"value":"c"}`} {
		body := []byte(`{"payload":` + value + `}`)
		if prior.ValidateResponse("get_payload", 200, body) != nil || next.ValidateResponse("get_payload", 200, body) != nil {
			t.Fatal("third-arm control lost old bytes")
		}
	}
	body := []byte(`{"payload":{"value":"b"}}`)
	if prior.ValidateResponse("get_payload", 200, body) == nil || next.ValidateResponse("get_payload", 200, body) == nil {
		t.Fatal("third-arm overlap was not already invalid")
	}
	if err := CheckCompatibilityPin(pin, next); err != nil {
		t.Fatalf("already-invalid third-arm overlap rejected: %v", err)
	}
}
