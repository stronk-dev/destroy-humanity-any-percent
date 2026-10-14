package curriculum

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

const CurriculumCatalogSchemaName = "CurriculumCatalogV1"

// CurriculumSchemas exports the ruled first-hour grammar for C18. Copy and
// Economy/Gate binding, branch order/cardinality and positive Decimal relations
// remain owner checks; candidate prices/counts/bonuses are still data.
func CurriculumSchemas() []publicapi.NamedSchema {
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
	branch := func(name string, starter *publicapi.Schema, extras ...publicapi.Field) *publicapi.Schema {
		return object(append([]publicapi.Field{field("branch", literal(name)), field("route_knowledge_bonus", integer(0, decimal.MaxExactInteger)),
			field("starter_package", starter), field("title_key", text("mechanical-id")), field("body_key", text("mechanical-id"))}, extras...)...)
	}
	branches := &publicapi.Schema{Kind: publicapi.SchemaArray, Items: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
		branch("acquihire", object(field("kind", literal("resource_grant")), field("resource_id", text("mechanical-id")), field("amount", text("canonical-decimal"))),
			field("minimum_purchased_generators", positive()), field("minimum_owned_upgrades", positive())),
		branch("burnout", object(field("kind", literal("generated_generators")), field("generator_id", text("mechanical-id")), field("count", positive())),
			field("cheapest_price_factor", text("canonical-decimal"))),
		branch("pivot", object(field("kind", literal("preowned_upgrade")), field("upgrade_id", text("mechanical-id")))),
	}}}
	return []publicapi.NamedSchema{{Name: CurriculumCatalogSchemaName, Schema: object(
		field("schema_version", integer(SchemaVersion, SchemaVersion)),
		field("first_failure", object(field("run_seq", integer(1, 1)), field("founder_exit_count", integer(0, 0)), field("attended_ms", integer(900000, 900000)),
			field("gate_id", literal("gate.t0_to_t1")), field("exit_type", literal("scripted_first")), field("evaluation", literal("first_player_company_command_after_accrual")),
			field("requested_command_effect", literal("replaced_by_terminal_transition")), field("next_run_key", text("mechanical-id")), field("branches", branches))),
	)}}
}
