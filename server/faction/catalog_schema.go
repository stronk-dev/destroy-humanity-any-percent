package faction

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/publicapi"
)

const FactionCatalogSchemaName = "FactionCatalogV1"

// FactionSchemas exports F1/FC's closed Phase-0 grammar for C18. The Hamiltonian
// cycle, row cardinality, copy binding, faction-specific Compact relationship and
// currently empty modifier list remain mandatory owner/composition checks.
func FactionSchemas() []publicapi.NamedSchema {
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
	enum := func(values []string) *publicapi.Schema {
		values = append([]string(nil), values...)
		sort.Strings(values)
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
	}
	array := func(item *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: item}
	}
	slots := make([]string, len(multiplier.Order))
	for index, slot := range multiplier.Order {
		slots[index] = string(slot)
	}
	compact := &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
		{Kind: publicapi.SchemaNull},
		object(field("auto_sign", &publicapi.Schema{Kind: publicapi.SchemaBoolean}), field("tithe_ppm", integer(0, 1_000_000))),
	}}
	return []publicapi.NamedSchema{{Name: FactionCatalogSchemaName, Schema: object(
		field("schema_version", integer(CatalogSchemaVersion, CatalogSchemaVersion)),
		field("stock_cap", integer(1, decimal.MaxExactInteger)), field("stock_interval_ms", integer(1, 86_400_000)),
		field("factions", array(object(field("id", enum(phase0FactionIDs)), field("produces", enum(phase0Resources)),
			field("consumes", enum(phase0Resources)), field("compact", compact),
			field("modifier_slots", array(object(field("slot", enum(slots)), field("ppm", &publicapi.Schema{Kind: publicapi.SchemaInteger})))),
			field("incorporation_copy_key", &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"})))),
	)}}
}
