package commons

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/publicapi"
)

const CommonsCatalogSchemaName = "CommonsCatalogV1"

// CommonsSchemas exports C1/C2's closed catalog grammar for API C18. Weight sums,
// threshold/rate/count relationships, positive bonus semantics, source uniqueness
// and declared Economy source/slot bindings remain mandatory owner checks.
func CommonsSchemas() []publicapi.NamedSchema {
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
	ppm := func() *publicapi.Schema { return integer(0, PPM) }
	positive := func() *publicapi.Schema { return integer(1, decimal.MaxExactInteger) }
	var slots []string
	for _, slot := range multiplier.Order {
		if slot != multiplier.SlotCommons {
			slots = append(slots, string(slot))
		}
	}
	sort.Strings(slots)
	return []publicapi.NamedSchema{{Name: CommonsCatalogSchemaName, Schema: object(
		field("schema_version", integer(CatalogSchemaVersion, CatalogSchemaVersion)),
		field("source_weights", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: object(
			field("source_id", &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}),
			field("slot", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: slots}), field("weight_ppm", integer(1, PPM)),
			field("forsworn", &publicapi.Schema{Kind: publicapi.SchemaBoolean}))}),
		field("default_tithe_ppm", ppm()), field("minimum_tithe_ppm", ppm()), field("maximum_tithe_ppm", ppm()),
		field("guild_health_weight_ppm", ppm()), field("cohort_health_weight_ppm", ppm()), field("server_health_weight_ppm", ppm()),
		field("collective_weight_ppm", ppm()), field("collective_exponent_ppm", integer(PPM, 10*PPM)),
		field("collapse_health_ppm", ppm()), field("healthy_health_ppm", ppm()),
		field("maximum_bonus", &publicapi.Schema{Kind: publicapi.SchemaString, Format: "canonical-decimal"}),
		field("health_recovery_ppm_per_hour", ppm()), field("health_decay_ppm_per_hour", ppm()),
		field("solidarity_window_ms", positive()), field("cohort_target_size", positive()), field("cohort_merge_floor", positive()),
		field("npc_population_floor", positive()), field("npc_weight_ppm", ppm()), field("npc_compliance_ppm", ppm()), field("population_tolerance_ppm", ppm()),
	)}}
}
