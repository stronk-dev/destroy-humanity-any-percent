package economy

import "cloud-clicker/server/publicapi"

const CostCurveSchemaName = "EconomyCostCurveV1"

// CostCurveSchemas is the K2 pricing component of C18, not a complete Economy
// artifact descriptor. Canonical state, nonnegative step and ratio >= 1 still
// require the owner; this export never chooses a price or evaluates a curve.
func CostCurveSchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	kind := func(value CurveKind) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{string(value)}}
	}
	value := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "canonical-decimal"}
	}
	return []publicapi.NamedSchema{{Name: CostCurveSchemaName, Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
		{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{field("kind", kind(CurveConstant))}},
		{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{field("kind", kind(CurveLinear)), field("step", value())}},
		{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{field("kind", kind(CurveGeometric)), field("ratio", value())}},
	}}}}
}
