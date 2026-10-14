package prestige

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

const PrestigePolicySchemaName = "PrestigePolicyV1"

// PrestigeSchemas exports the existing closed policy grammar for C18. Positive
// state-valued threshold semantics, table cardinality and Economy resource binding
// remain owner/composition checks; no formulas or balance literals are retuned.
func PrestigeSchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	object := func(fields ...publicapi.Field) *publicapi.Schema {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
	}
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
	}
	ppm := func() *publicapi.Schema { return integer(0, 1_000_000) }
	modifiers := []publicapi.Field{}
	for _, kind := range []string{"acquihire", "acquisition", "collapse", "ipo", "scripted_first"} {
		modifiers = append(modifiers, field(kind, integer(0, decimal.MaxExactInteger)))
	}
	return []publicapi.NamedSchema{{Name: PrestigePolicySchemaName, Schema: object(
		field("schema_version", integer(1, 1)), field("value_resource_id", &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}),
		field("threshold", &publicapi.Schema{Kind: publicapi.SchemaString, Format: "canonical-decimal"}), field("exit_modifiers_ppm", object(modifiers...)),
		field("collapse_route_knowledge", integer(0, decimal.MaxExactInteger)), field("catchup_ceiling_ms", integer(1, 86_400_000)),
		field("offer_duration_ms", integer(1, MaximumOfferDurationMS)), field("decline_drift_ppm", ppm()),
		field("spawn_gate_ppm", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: ppm()}), field("advisor_per_run_ppm", ppm()), field("advisor_cap_ppm", ppm())),
	}}
}
