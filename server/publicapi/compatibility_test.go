package publicapi

import (
	"strings"
	"testing"
)

func compatibilityPayloadRegistry(t *testing.T, payload *Schema, shape string) *Registry {
	t.Helper()
	response := &Schema{Kind: SchemaRef, Ref: "Payload"}
	if shape == "array" {
		response = &Schema{Kind: SchemaArray, Items: response}
	} else if shape == "union" {
		response = &Schema{Kind: SchemaOneOf, Alternates: []*Schema{{Kind: SchemaNull}, response}}
	}
	schemas := []NamedSchema{{Name: "Envelope", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
		{Name: "payload", Schema: response, Required: true},
	}}}, {Name: "Payload", Schema: payload}}
	operation := Operation{ID: "get_payload", Method: "GET", Path: "/api/public/v1/payload",
		Surface: SurfacePublicV1, Auth: AuthNone, Public: true,
		Responses: []Response{{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "Envelope"}}}
	if shape == "request" || shape == "shared" {
		operation.Method, operation.Path, operation.Surface, operation.Public = "POST", "/api/v1/payload", SurfacePrivateV1, false
		operation.Request = "Payload"
		if shape == "request" {
			schemas[0].Schema.Fields[0].Schema = &Schema{Kind: SchemaNull}
		}
	}
	registry, err := NewRegistry(schemas, []Operation{operation})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func TestCompatibilityPinRejectsOptionalRequiredPromotion(t *testing.T) {
	for _, shape := range []string{"direct", "array", "union", "request", "shared"} {
		t.Run(shape, func(t *testing.T) {
			payload := func(required bool) *Schema {
				return &Schema{Kind: SchemaObject, Fields: []Field{{Name: "value", Schema: &Schema{Kind: SchemaString}, Required: required}}}
			}
			pin, err := CanonicalOperationPins(compatibilityPayloadRegistry(t, payload(false), shape))
			if err != nil {
				t.Fatal(err)
			}
			if err := CheckCompatibilityPin(pin, compatibilityPayloadRegistry(t, payload(false), shape)); err != nil {
				t.Fatalf("unchanged optional field rejected: %v", err)
			}
			if err := CheckCompatibilityPin(pin, compatibilityPayloadRegistry(t, payload(true), shape)); err == nil {
				t.Fatal("existing optional field became required inside v1")
			} else if !strings.Contains(err.Error(), "value") {
				t.Fatalf("requiredness failure did not identify field value: %v", err)
			}
		})
	}
}

func TestCompatibilityPinComparesInt64BoundsExactly(t *testing.T) {
	const twoTo53 = int64(1 << 53)
	for _, test := range []struct {
		name          string
		bound         string
		before, after int64
		refused       bool
	}{
		{"maximum narrows above float precision", "maximum", twoTo53 + 1, twoTo53, true},
		{"minimum narrows above float precision", "minimum", twoTo53, twoTo53 + 1, true},
		{"maximum narrows at int64 limit", "maximum", 1<<63 - 1, 1<<63 - 2, true},
		{"minimum narrows at int64 limit", "minimum", -1 << 63, -1<<63 + 1, true},
		{"maximum widens above float precision", "maximum", twoTo53, twoTo53 + 1, false},
		{"minimum widens above float precision", "minimum", twoTo53 + 1, twoTo53, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			payload := func(bound int64) *Schema {
				value := &Schema{Kind: SchemaInteger}
				if test.bound == "minimum" {
					value.Minimum = &bound
				} else {
					value.Maximum = &bound
				}
				return &Schema{Kind: SchemaObject, Fields: []Field{{Name: "counter", Schema: value, Required: true}}}
			}
			pin, err := CanonicalOperationPins(compatibilityPayloadRegistry(t, payload(test.before), "array"))
			if err != nil {
				t.Fatal(err)
			}
			err = CheckCompatibilityPin(pin, compatibilityPayloadRegistry(t, payload(test.after), "array"))
			if (err != nil) != test.refused {
				t.Fatalf("bound %d -> %d: error=%v, want refused=%v", test.before, test.after, err, test.refused)
			}
		})
	}
}

func TestCompatibilityPinAllowsOnlyAdditiveV1Changes(t *testing.T) {
	prior := testRegistry(t)
	pin, err := CanonicalOperationPins(prior)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckCompatibilityPin(pin, prior); err != nil {
		t.Fatal(err)
	}

	additiveSchemas := testSchemas()
	additiveSchemas[0].Schema.Fields = append(additiveSchemas[0].Schema.Fields,
		Field{Name: "request_id", Schema: &Schema{Kind: SchemaString}, Required: false})
	additiveOperations := testOperations()
	additiveOperations = append(additiveOperations, Operation{ID: "get_status", Method: "GET", Path: "/api/public/v1/status",
		Surface: SurfacePublicV1, Auth: AuthNone, Public: true,
		Responses: []Response{{Kind: ResponseSchema, Status: 200, ContentType: ContentJSON, SchemaRef: "APIError"}}})
	additive, err := NewRegistry(additiveSchemas, additiveOperations)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckCompatibilityPin(pin, additive); err != nil {
		t.Fatalf("additive response property/operation rejected: %v", err)
	}

	for name, mutate := range map[string]func([]NamedSchema, []Operation){
		"operation removal": func(_ []NamedSchema, operations []Operation) { operations[0].Method = "POST" },
		"response property removal": func(schemas []NamedSchema, _ []Operation) {
			schemas[1].Schema.Fields = schemas[1].Schema.Fields[:1]
		},
		"constraint narrowing": func(schemas []NamedSchema, _ []Operation) {
			maximum := int64(10)
			schemas[2].Schema.Fields[0].Schema.Maximum = &maximum
		},
	} {
		t.Run(name, func(t *testing.T) {
			schemas, operations := testSchemas(), testOperations()
			mutate(schemas, operations)
			registry, err := NewRegistry(schemas, operations)
			if err != nil {
				t.Fatal(err)
			}
			if err := CheckCompatibilityPin(pin, registry); err == nil {
				t.Fatal("incompatible v1 change accepted")
			}
		})
	}
}
