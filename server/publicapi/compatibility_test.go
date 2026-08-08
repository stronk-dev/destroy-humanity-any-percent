package publicapi

import "testing"

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
