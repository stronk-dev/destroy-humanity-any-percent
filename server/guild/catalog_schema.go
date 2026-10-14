package guild

import (
	"sort"

	"cloud-clicker/server/publicapi"
)

const GuildCatalogSchemaName = "GuildCatalogV1"

// GuildSchemas exports the currently admitted GB/GD3/GD5 Phase-0 policy for C18.
// It does not widen literal admission or authorize a retune, new membership rule,
// clearing formula, Health contract or multiplier effect.
func GuildSchemas() []publicapi.NamedSchema {
	literal := func(name string, value int64) publicapi.Field {
		minimum, maximum := value, value
		return publicapi.Field{Name: name, Required: true, Schema: &publicapi.Schema{
			Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum,
		}}
	}
	fields := []publicapi.Field{
		literal("schema_version", CatalogSchemaVersion), literal("guild_tithe_ppm", 20_000),
		literal("clearing_rate_ppm", 500_000), literal("npc_exchange_ppm", 250_000),
		literal("stock_intake_cap", 120), literal("consumption_bonus_ppm_per_unit", 0),
		literal("max_members", 50), literal("min_members", 2), literal("grace_days", 7),
		literal("guild_xp_target_per_founder", 250_000), literal("clearing_interval_ms", 300_000),
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return []publicapi.NamedSchema{{Name: GuildCatalogSchemaName, Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}}}
}
