package publicapi

import (
	"fmt"
	"testing"
)

func TestCompatibilityPinRejectsAddedResponseArmInvalidatingOldBytes(t *testing.T) {
	for _, shape := range []string{"direct", "array", "union"} {
		for _, addition := range []string{"duplicate", "broader"} {
			t.Run(shape+"/"+addition, func(t *testing.T) {
				old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
					{Kind: SchemaString, Enum: []string{"a"}},
					{Kind: SchemaString, Enum: []string{"b"}},
				}}
				prior := compatibilityPayloadRegistry(t, old, shape)
				pin, err := CanonicalOperationPins(prior)
				if err != nil {
					t.Fatal(err)
				}
				value := `"a"`
				if shape == "array" {
					value = `["a"]`
				}
				body := []byte(fmt.Sprintf(`{"payload":%s}`, value))
				if err := prior.ValidateResponse("get_payload", 200, body); err != nil {
					t.Fatalf("old response fixture is invalid: %v", err)
				}
				extra := &Schema{Kind: SchemaString}
				if addition == "duplicate" {
					extra.Enum = []string{"a"}
				}
				nextPayload := &Schema{Kind: SchemaOneOf, Alternates: append(append([]*Schema(nil), old.Alternates...), extra)}
				next := compatibilityPayloadRegistry(t, nextPayload, shape)
				if err := next.ValidateResponse("get_payload", 200, body); err == nil {
					t.Fatal("overlapping addition did not invalidate the actual old response")
				}
				if err := CheckCompatibilityPin(pin, next); err == nil {
					t.Fatal("compatibility admitted an addition that rejects previously valid response bytes")
				}
			})
		}
	}
}

func TestCompatibilityPinPreservesDisjointResponseUnionGrowth(t *testing.T) {
	for _, shape := range []string{"direct", "array", "union", "request", "shared"} {
		t.Run(shape, func(t *testing.T) {
			old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
				{Kind: SchemaString, Enum: []string{"a"}},
				{Kind: SchemaString, Enum: []string{"b"}},
			}}
			prior := compatibilityPayloadRegistry(t, old, shape)
			pin, err := CanonicalOperationPins(prior)
			if err != nil {
				t.Fatal(err)
			}
			for _, extra := range []*Schema{{Kind: SchemaBoolean}, {Kind: SchemaString, Enum: []string{"c"}}} {
				nextPayload := &Schema{Kind: SchemaOneOf, Alternates: append(append([]*Schema(nil), old.Alternates...), extra)}
				next := compatibilityPayloadRegistry(t, nextPayload, shape)
				err := CheckCompatibilityPin(pin, next)
				wantRefusal := shape == "request" || shape == "shared"
				if (err != nil) != wantRefusal {
					t.Fatalf("disjoint addition: err=%v wantRefusal=%t", err, wantRefusal)
				}
			}
		})
	}
}

func TestCompatibilityPinResponseUnionReferenceGrowth(t *testing.T) {
	for _, tag := range []string{"a", "c"} {
		t.Run(tag, func(t *testing.T) {
			definitions := []NamedSchema{
				{Name: "OldAlias", Schema: &Schema{Kind: SchemaRef, Ref: "OldValue"}},
				{Name: "OldValue", Schema: &Schema{Kind: SchemaString, Enum: []string{"a"}}},
				{Name: "NewAlias", Schema: &Schema{Kind: SchemaRef, Ref: "NewValue"}},
				{Name: "NewValue", Schema: &Schema{Kind: SchemaString, Enum: []string{tag}}},
			}
			old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
				{Kind: SchemaRef, Ref: "OldAlias"}, {Kind: SchemaBoolean},
			}}
			prior := compatibilityPayloadRegistry(t, old, "direct", definitions...)
			pin, err := CanonicalOperationPins(prior)
			if err != nil {
				t.Fatal(err)
			}
			next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf,
				Alternates: append(append([]*Schema(nil), old.Alternates...), &Schema{Kind: SchemaRef, Ref: "NewAlias"})}, "direct", definitions...)
			body := []byte(`{"payload":"a"}`)
			if err := prior.ValidateResponse("get_payload", 200, body); err != nil {
				t.Fatal(err)
			}
			wantRefusal := tag == "a"
			if (next.ValidateResponse("get_payload", 200, body) != nil) != wantRefusal {
				t.Fatal("reference fixture did not exercise old-byte ambiguity")
			}
			if err := CheckCompatibilityPin(pin, next); (err != nil) != wantRefusal {
				t.Fatalf("reference growth: err=%v wantRefusal=%t", err, wantRefusal)
			}
		})
	}
}

func TestCompatibilityPinResponseUnionTaggedObjectGrowth(t *testing.T) {
	row := func(tag string) *Schema {
		return &Schema{Kind: SchemaObject, Fields: []Field{
			{Name: "name", Required: true, Schema: &Schema{Kind: SchemaString, Enum: []string{tag}}},
		}}
	}
	old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{row("a"), row("b")}}
	prior := compatibilityPayloadRegistry(t, old, "direct")
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{"a", "c"} {
		next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf,
			Alternates: append(append([]*Schema(nil), old.Alternates...), row(tag))}, "direct")
		wantRefusal := tag == "a"
		body := []byte(`{"payload":{"name":"a"}}`)
		if err := prior.ValidateResponse("get_payload", 200, body); err != nil {
			t.Fatal(err)
		}
		if (next.ValidateResponse("get_payload", 200, body) != nil) != wantRefusal {
			t.Fatal("tagged fixture did not exercise old-byte ambiguity")
		}
		if err := CheckCompatibilityPin(pin, next); (err != nil) != wantRefusal {
			t.Fatalf("tagged growth: err=%v wantRefusal=%t", err, wantRefusal)
		}
	}
}

func TestCompatibilityPinResponseUnionExactIntegerBoundary(t *testing.T) {
	boundary := int64(1<<53 + 1)
	old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
		{Kind: SchemaInteger, Maximum: &boundary}, {Kind: SchemaBoolean},
	}}
	prior := compatibilityPayloadRegistry(t, old, "direct")
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	for _, minimum := range []int64{boundary, boundary + 1} {
		next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf,
			Alternates: append(append([]*Schema(nil), old.Alternates...), &Schema{Kind: SchemaInteger, Minimum: &minimum})}, "direct")
		body := []byte(fmt.Sprintf(`{"payload":%d}`, boundary))
		if err := prior.ValidateResponse("get_payload", 200, body); err != nil {
			t.Fatal(err)
		}
		wantRefusal := minimum == boundary
		if (next.ValidateResponse("get_payload", 200, body) != nil) != wantRefusal {
			t.Fatal("integer fixture did not exercise the exact old-byte boundary")
		}
		if err := CheckCompatibilityPin(pin, next); (err != nil) != wantRefusal {
			t.Fatalf("integer growth: err=%v wantRefusal=%t", err, wantRefusal)
		}
	}
}

func TestCompatibilityPinResponseUnionArraysStillOverlapAtEmpty(t *testing.T) {
	old := &Schema{Kind: SchemaOneOf, Alternates: []*Schema{
		{Kind: SchemaArray, Items: &Schema{Kind: SchemaString}}, {Kind: SchemaNull},
	}}
	prior := compatibilityPayloadRegistry(t, old, "direct")
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	next := compatibilityPayloadRegistry(t, &Schema{Kind: SchemaOneOf,
		Alternates: append(append([]*Schema(nil), old.Alternates...), &Schema{Kind: SchemaArray, Items: &Schema{Kind: SchemaInteger}})}, "direct")
	body := []byte(`{"payload":[]}`)
	if err := prior.ValidateResponse("get_payload", 200, body); err != nil {
		t.Fatal(err)
	}
	if err := next.ValidateResponse("get_payload", 200, body); err == nil {
		t.Fatal("empty arrays were assumed disjoint because their item types differ")
	}
	if err := CheckCompatibilityPin(pin, next); err == nil {
		t.Fatal("array addition invalidated the old empty array")
	}
}
