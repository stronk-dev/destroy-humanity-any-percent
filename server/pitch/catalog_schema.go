package pitch

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

const PitchCatalogSchemaName = "PitchCatalogV1"

// PitchSchemas exports the TP-C13/TP-C21/TP-C22 structural grammar for API C18.
// Row cardinality/order, positive Decimal semantics, copy bindings, chain partners
// and raw JSON member uniqueness remain mandatory owner admission checks.
func PitchSchemas() []publicapi.NamedSchema {
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
	id := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}
	}
	value := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "canonical-decimal"}
	}
	enum := func(values ...string) *publicapi.Schema {
		sort.Strings(values)
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
	}
	array := func(item *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: item}
	}
	policy := object(
		field("hand_size", integer(7, 7)), field("play_size", integer(4, 4)),
		field("hands_per_round", integer(3, 3)), field("hack_slots", integer(4, 4)),
		field("start_currency", integer(4, 4)), field("shop_size", integer(3, 3)),
		field("round_clear_currency", integer(3, 3)), field("best_exponent_hardcap", integer(1_000_000, 1_000_000)),
		field("hand_size_reason_key", id()), field("play_size_reason_key", id()),
		field("hands_per_round_reason_key", id()), field("hack_slots_reason_key", id()),
		field("rounds_reason_key", id()), field("exponent_reason_key", enum("cap.pitch_exponent")))
	effect := &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
		object(field("kind", enum("flat_add")), field("amount", value())),
		object(field("kind", enum("card_factor")), field("factor", value())),
		object(field("kind", enum("shape_factor")), field("shape", enum("pair", "full_hand")), field("factor", value())),
		object(field("kind", enum("chain_factor")), field("partner_hack_id", id()), field("factor", value())),
	}}
	return []publicapi.NamedSchema{{Name: PitchCatalogSchemaName, Schema: object(
		field("schema_version", integer(SchemaVersion, SchemaVersion)), field("policy", policy),
		field("metric_cards", array(object(field("card_id", id()), field("base_metric", value()),
			field("copies", integer(2, 2)), field("copy_key", id())))),
		field("growth_hacks", array(object(field("hack_id", id()), field("price", integer(0, decimal.MaxExactInteger)),
			field("draft_weight", integer(1, decimal.MaxExactInteger)), field("effect", effect), field("copy_key", id())))),
		field("funding_curve", array(object(field("round", integer(1, 8)), field("funding_target", value())))),
	)}}
}
