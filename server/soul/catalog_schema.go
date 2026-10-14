package soul

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

const SoulCatalogSchemaName = "SoulCatalogV1"

// SoulSchemas exports the existing SB10/SB17/SB27 and SR-C2 structural grammar
// for C18. Copy/band/domain relationships, unique keys, epoch-vs-fixture rules,
// row order/cardinality and catch-up/watchdog binding remain owner checks.
func SoulSchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	object := func(fields ...publicapi.Field) *publicapi.Schema {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
	}
	id := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}
	}
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
	}
	positive := func() *publicapi.Schema { return integer(1, decimal.MaxExactInteger) }
	value := func() *publicapi.Schema { return integer(0, decimal.MaxExactInteger) }
	boolean := func() *publicapi.Schema { return &publicapi.Schema{Kind: publicapi.SchemaBoolean} }
	enum := func(values ...string) *publicapi.Schema {
		sort.Strings(values)
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
	}
	array := func(item *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: item}
	}
	return []publicapi.NamedSchema{{Name: SoulCatalogSchemaName, Schema: object(
		field("schema_version", integer(1, 1)),
		field("policy", object(field("soul_floor", value()), field("soul_initial", value()), field("soul_max", value()),
			field("recovery_beat_ceiling_ms", positive()), field("max_session_wall_ms", positive()))),
		field("bands", array(object(field("band_member", enum(string(BandNearZero), string(BandHollow), string(BandDimming), string(BandWhole))),
			field("min_inclusive", value()), field("max_inclusive", value()), field("human_content_locked", boolean()), field("reason_key", id())))),
		field("debit_sources", array(object(field("source_id", id()), field("owner_kind", enum(string(OwnerEvent), string(OwnerLongevity), string(OwnerContract), string(OwnerFixture))),
			field("amount", positive()), field("may_exhaust", boolean()), field("single_use", boolean()), field("curtain_copy_key", id())))),
		field("recovery_activities", array(object(field("activity_id", id()), field("duration_attended_ms", positive()), field("recovery_amount", positive()),
			field("toy_kind", enum(string(ToyDefrag), string(ToyRepot), string(ToyServerRoom))), field("reason_key", id()),
			field("title_copy_key", id()), field("description_copy_key", id()), field("disclosure_copy_key", id())))),
		field("ending_policy", object(field("whole_variant", enum(string(EndingEarnestAscension))), field("depleted_variant", enum(string(EndingTrainingData))))),
	)}}
}
