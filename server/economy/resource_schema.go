package economy

import (
	"sort"

	"cloud-clicker/server/publicapi"
)

const ResourceSchemaName = "EconomyResourceV1"

// ResourceSchemas is K1's shared resource component, not a complete C18 Economy
// artifact. State Decimal semantics and minimum/initial/cap relations remain
// mandatory owner checks; this export never selects a balance or gameplay cap.
func ResourceSchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	object := func(fields ...publicapi.Field) *publicapi.Schema {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
	}
	value := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "canonical-decimal"}
	}
	return []publicapi.NamedSchema{{Name: ResourceSchemaName, Schema: object(
		field("id", &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}),
		field("scope", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{string(ScopeCompany), string(ScopeFounder), string(ScopeGuild), string(ScopeWorld)}}),
		field("numeric_kind", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{"decimal"}}),
		field("initial", value()), field("minimum", value()),
		field("hardcap", &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			{Kind: publicapi.SchemaNull}, object(field("amount", value()), field("reason_key", &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"})),
		}}),
	)}}
}
