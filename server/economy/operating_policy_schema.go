package economy

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

// OperatingPolicySchemas describes the existing manual/offline policy shapes.
// Efficiency, burst and banking relationships stay with the owner, not the DSL.
func OperatingPolicySchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	positive := func() *publicapi.Schema {
		minimum, maximum := int64(1), int64(decimal.MaxExactInteger)
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
	}
	value := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "canonical-decimal"}
	}
	object := func(fields ...publicapi.Field) *publicapi.Schema {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
	}
	return []publicapi.NamedSchema{
		{Name: "EconomyManualPolicyV1", Schema: object(field("refill_milli_per_ms", positive()), field("bucket_cap_milli", positive()))},
		{Name: "EconomyOfflinePolicyV1", Schema: object(field("efficiency", value()), field("accrual_cap_ms", positive()), field("bank_ratio_numerator", positive()),
			field("bank_ratio_denominator", positive()), field("bank_cap_ms", positive()), field("burst_speed", value()), field("burst_max_duration_ms", positive()))},
	}
}
