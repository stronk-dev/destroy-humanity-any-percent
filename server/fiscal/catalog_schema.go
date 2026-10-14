package fiscal

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/publicapi"
)

const FiscalCatalogSchemaName = "FiscalCatalogV1"

// FiscalSchemas exports F9's closed structural grammar for C18. Clock/credit
// relationships, sequence horizon, row order/cardinality and pinned Economy
// declarations remain mandatory semantic owner checks.
func FiscalSchemas() []publicapi.NamedSchema {
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
	array := func(item *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: item}
	}
	slot := func() *publicapi.Schema {
		values := make([]string, len(multiplier.Order))
		for i, value := range multiplier.Order {
			values[i] = string(value)
		}
		sort.Strings(values)
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
	}
	return []publicapi.NamedSchema{{Name: FiscalCatalogSchemaName, Schema: object(
		field("schema_version", integer(CatalogSchemaVersion, CatalogSchemaVersion)),
		field("clock_policy", object(field("early_ms", positive()), field("guaranteed_ms", positive()),
			field("auto_ms", positive()), field("early_success_ppm", integer(0, PPM)))),
		field("credit_policy", object(field("credit_per_period", positive()), field("hardcap", positive()), field("hardcap_reason_key", id()))),
		field("hoard_policy", object(field("ppm_per_credit", integer(1, PPM)), field("cap_credits", positive()), field("slot", slot()),
			field("source_id", id()), field("target", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{"all"}}))),
		field("generator_level_rows", array(object(field("generator_id", id()), field("ppm_per_level", integer(1, PPM)),
			field("level_hardcap", positive()), field("hardcap_reason_key", id()), field("slot", slot()), field("source_id", id())))),
		field("unlock_rows", array(object(field("unlock_id", id()), field("cost", positive())))),
	)}}
}
