package activeplay

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

const ActivePlayCatalogSchemaName = "ActivePlayCatalogV1"

// ActivePlaySchemas exports A9's closed structural artifact grammar for C18.
// Bindings, Decimal relationships, array cardinality/order, weight totals and
// scheduler horizons remain mandatory semantic owner checks.
func ActivePlaySchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	object := func(fields ...publicapi.Field) *publicapi.Schema {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
	}
	text := func(format string) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: format}
	}
	literal := func(value string) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{value}}
	}
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
	}
	positive := func() *publicapi.Schema { return integer(1, decimal.MaxExactInteger) }
	array := func(item *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: item}
	}
	effect := func(kind string, fields ...publicapi.Field) *publicapi.Schema {
		return object(append([]publicapi.Field{field("effect_row_id", text("mechanical-id")), field("kind", literal(kind)), field("weight", positive())}, fields...)...)
	}
	return []publicapi.NamedSchema{{Name: ActivePlayCatalogSchemaName, Schema: object(
		field("schema_version", integer(CatalogSchemaVersion, CatalogSchemaVersion)),
		field("schedule_policy", object(field("sampler_version", literal(SamplerVersion)), field("substream_label", literal(SpawnSubstream)),
			field("minimum_interval_ms", positive()), field("scale_ms", positive()), field("lifetime_ms", positive()), field("max_due_transitions", positive()))),
		field("effects", array(&publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			effect("production_frenzy", field("factor", text("canonical-decimal")), field("duration_ms", positive()), field("targets", array(literal("generator_production")))),
			effect("click_frenzy", field("factor", text("canonical-decimal")), field("duration_ms", positive()), field("action_ids", array(text("mechanical-id")))),
			effect("building_special", field("per_owned_ppm", integer(1, 1_000_000)), field("duration_ms", positive()), field("eligible_generator_ids", array(text("mechanical-id")))),
			effect("lucky_payout", field("lucky_bank_frac", text("canonical-decimal")), field("lucky_rate_cap", text("canonical-decimal")),
				field("epsilon", text("canonical-decimal")), field("resource_id", text("mechanical-id")), field("hardcap_reason_key", text("mechanical-id"))),
		}})),
		field("combo_policy", object(field("combo_cap", text("canonical-decimal")), field("hardcap_reason_key", text("mechanical-id")))),
	)}}
}
