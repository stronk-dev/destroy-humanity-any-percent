package meters

import (
	"sort"

	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/publicapi"
)

const MeterCatalogSchemaName = "MeterCatalogV1"

// MeterSchemas exports C18's existing closed structural v1 contract. Exact
// eleven-row membership/order, source uniqueness, band ordering and reseed
// relationships remain semantic loader checks; values are not retuned here.
func MeterSchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	object := func(fields ...publicapi.Field) *publicapi.Schema {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
	}
	enum := func(values ...string) *publicapi.Schema {
		sort.Strings(values)
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
	}
	id := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}
	}
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
	}
	array := func(items *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: items}
	}
	nonzero := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{integer(-MaximumValue, -1), integer(1, MaximumValue)}}
	}
	slots := make([]string, len(multiplier.Order))
	for i, slot := range multiplier.Order {
		slots[i] = string(slot)
	}
	input := &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
		object(field("kind", enum(string(InputLedgerFact))), field("fact_kind", id()), field("delta", nonzero())),
		object(field("kind", enum(string(InputContributionSlot))), field("slot", enum(slots...)), field("source_id", id()), field("delta_per_attended_hour", nonzero())),
	}}
	decay := &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
		{Kind: publicapi.SchemaNull},
		object(field("toward_value", integer(MinimumValue, MaximumValue)), field("rate_per_attended_hour", integer(1, MaximumValue))),
	}}
	row := object(field("id", enum(append([]string(nil), requiredMeterIDs...)...)), field("scope", enum("company")),
		field("min_value", integer(MinimumValue, MinimumValue)), field("max_value", integer(MaximumValue, MaximumValue)),
		field("initial_value", integer(MinimumValue, MaximumValue)),
		field("bands", array(object(field("id", id()), field("floor_value", integer(MinimumValue, MaximumValue))))),
		field("inputs", array(input)), field("decay", decay))
	return []publicapi.NamedSchema{{Name: MeterCatalogSchemaName, Schema: object(
		field("schema_version", integer(CatalogSchemaVersion, CatalogSchemaVersion)), field("meters", array(row)),
		field("trust_reseed", object(field("base_value", integer(MinimumValue, MaximumValue)),
			field("notoriety_numerator", integer(0, maximumExactInteger)), field("notoriety_denominator", integer(1, maximumExactInteger)),
			field("floor_value", integer(MinimumValue, MaximumValue)), field("ceiling_value", integer(MinimumValue, MaximumValue)))),
	)}}
}
