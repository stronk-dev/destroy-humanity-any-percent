package pet

import (
	"sort"

	"cloud-clicker/server/publicapi"
)

const PetCatalogSchemaName = "PetCatalog"

// PetSchemas exports the complete C17 care policy and Soul SB22 v2 extension for
// API C18. Relations, cardinality/order, deterministic transitions, raw member
// uniqueness and Soul bundle eligibility still require owner/composition checks.
func PetSchemas() []publicapi.NamedSchema {
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
	enum := func(values ...string) *publicapi.Schema {
		sort.Strings(values)
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
	}
	array := func(item *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: item}
	}
	version := func(version int64) *publicapi.Schema {
		action := []publicapi.Field{
			field("action_id", &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}),
			field("stat_id", enum(petSchemaMembers(StatIDs())...)), field("delta_ppm", integer(1, 1_000_000)),
			field("cooldown_attended_ms", integer(0, maxExactInteger)), field("min_eligible_ppm", ppm()),
		}
		if version == 2 {
			action = append(action, field("soul_gate", enum("essential", "ordinary", "recovery")))
		}
		return object(field("schema_version", integer(version, version)),
			field("stat_policy", object(field("grid_ms", integer(1, maxExactInteger)),
				field("diminishing_threshold_ppm", ppm()), field("diminishing_factor_ppm", ppm()),
				field("stats", array(object(field("stat_id", enum(petSchemaMembers(StatIDs())...)),
					field("initial_ppm", ppm()), field("floor_ppm", ppm()), field("decay_ppm_per_grid", ppm())))))),
			field("actions", array(object(action...))),
			field("trust_policy", object(field("initial_ppm", ppm()), field("neutral_ppm", ppm()), field("floor_ppm", ppm()),
				field("cap_ppm", ppm()), field("gain_ppm_per_effective_action", ppm()), field("decay_ppm_per_grid", ppm()))),
			field("mood_policy", array(object(field("mood_member", enum(petSchemaMembers(Moods())...)), field("floor_ppm", ppm())))),
			field("behavior_policy", array(object(field("from_state", enum(petSchemaMembers(BehaviorStates())...)),
				field("event", enum(petSchemaMembers(BehaviorEvents())...)), field("to_state", enum(petSchemaMembers(BehaviorStates())...)),
				field("duration_grid_ticks", integer(1, maxExactInteger))))))
	}
	return []publicapi.NamedSchema{
		{Name: PetCatalogSchemaName, Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			{Kind: publicapi.SchemaRef, Ref: "PetCatalogV1"}, {Kind: publicapi.SchemaRef, Ref: "PetCatalogV2"},
		}}},
		{Name: "PetCatalogV1", Schema: version(1)}, {Name: "PetCatalogV2", Schema: version(2)},
	}
}

func petSchemaMembers[T ~string](members []T) []string {
	result := make([]string, len(members))
	for index, member := range members {
		result[index] = string(member)
	}
	return result
}
