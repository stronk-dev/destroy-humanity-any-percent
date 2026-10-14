package minigame

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

const MinigameCatalogSchemaName = "MinigameCatalog"

// MinigameSchemas exports the closed C27-C34/Soul SB22 artifact grammar for C18.
// Ranked scaling remains breadth/presentation-only. Ordering, identity/bindings,
// policy relationships and raw-member uniqueness remain owner/composition checks.
func MinigameSchemas() []publicapi.NamedSchema {
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
	nonnegative := func() *publicapi.Schema { return integer(0, decimal.MaxExactInteger) }
	signed := func() *publicapi.Schema { return integer(-decimal.MaxExactInteger, decimal.MaxExactInteger) }
	ppm := func() *publicapi.Schema { return integer(0, payoutPPM) }
	id := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}
	}
	version := func() *publicapi.Schema { return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "semver"} }
	enum := func(values ...string) *publicapi.Schema {
		sort.Strings(values)
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
	}
	array := func(item *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: item}
	}
	oneOf := func(arms ...*publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: arms}
	}
	ref := func(name string) *publicapi.Schema { return &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: name} }
	scaling := object(field("schema_version", integer(1, 1)), field("scaling_inputs", array(object(
		field("destination", id()), field("destination_class", enum(string(DestinationBreadth), string(DestinationPresentation))),
		field("source_kind", enum(string(ScalingLiteral), string(ScalingTier), string(ScalingPurchasedGeneratorCount), string(ScalingFounderCarryCounter), string(ScalingAttendedQualityGrade))),
		field("source_ref", &publicapi.Schema{Kind: publicapi.SchemaString}), field("op", enum(string(ScalingIdentity), string(ScalingAdd), string(ScalingMul), string(ScalingFloorDiv))),
		field("operand", signed()), field("clamp_min", signed()), field("clamp_max", signed())))))
	payout := object(field("credited_resource_id", id()), field("sends_per_day", nonnegative()), field("per_send_cap", nonnegative()),
		field("conversion_ppm", ppm()), field("payout_score_fact_id", id()), field("cap_reason_key", id()))
	fallback := oneOf(object(field("kind", enum(string(FallbackSolo)))),
		object(field("kind", enum(string(FallbackBot))), field("bot_ref", object(field("policy_id", id()), field("version", version()))), field("rate_reduction_ppm", ppm())),
		object(field("kind", enum(string(FallbackNPCPartner))), field("npc_profile", object(field("profile_id", id()), field("version", version()))), field("rate_reduction_ppm", ppm())))
	quality := object(field("score_fact", id()), field("grade_curve", array(object(field("score_threshold", nonnegative()), field("grade_ppm", ppm())))),
		field("decay_grid_ms", integer(1, decimal.MaxExactInteger)), field("decay_ppm_per_grid", ppm()), field("neutral_floor_ppm", ppm()), field("automation_destination", id()))
	rating := object(field("starting_elo", signed()), field("elo_floor", signed()), field("elo_ceiling", signed()), field("provisional_games", nonnegative()), field("season_member", id()))
	exit := field("exit_history_at_least", nonnegative())
	exit.Required = false
	unlock := oneOf(object(field("kind", enum("always"))),
		object(field("kind", enum("fact_equals")), field("fact_id", id()), field("value", oneOf(&publicapi.Schema{Kind: publicapi.SchemaBoolean}, &publicapi.Schema{Kind: publicapi.SchemaString}, signed()))),
		object(field("kind", enum("fiscal_unlock")), field("unlock_id", id())),
		object(field("kind", enum("tier_at_least")), field("tier", integer(0, 9)), exit))
	catalog := func(schemaVersion int64) *publicapi.Schema {
		fields := []publicapi.Field{
			field("minigame_id", id()), field("engine_ref", id()), field("engine_version", version()),
			field("modes", array(enum(string(ModeSolo), string(ModeAsyncSnapshot)))), field("result_score_fact_ids", array(id())),
			field("scaling", ref("MinigameRankedScalingV1")), field("payout", ref("MinigamePayoutPolicyV1")), field("fallback", ref("MinigameFallbackPolicyV1")),
			field("offline_quality", ref("MinigameOfflineQualityPolicyV1")), field("rating_policy", ref("MinigameRatingPolicyV1")), field("unlock_condition", ref("MinigameUnlockConditionV1")),
		}
		if schemaVersion == 3 {
			fields = append(fields, field("soul_gate", enum("human_hobby", "unrelated")))
		}
		return object(field("schema_version", integer(schemaVersion, schemaVersion)), field("rating_seasons", array(id())), field("minigames", array(object(fields...))))
	}
	definitions := []publicapi.NamedSchema{
		{Name: MinigameCatalogSchemaName, Schema: oneOf(ref("MinigameCatalogV2"), ref("MinigameCatalogV3"))},
		{Name: "MinigameCatalogV2", Schema: catalog(2)}, {Name: "MinigameCatalogV3", Schema: catalog(3)},
		{Name: "MinigameRankedScalingV1", Schema: scaling}, {Name: "MinigamePayoutPolicyV1", Schema: payout},
		{Name: "MinigameFallbackPolicyV1", Schema: fallback}, {Name: "MinigameOfflineQualityPolicyV1", Schema: quality},
		{Name: "MinigameRatingPolicyV1", Schema: rating}, {Name: "MinigameUnlockConditionV1", Schema: unlock},
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	return definitions
}
