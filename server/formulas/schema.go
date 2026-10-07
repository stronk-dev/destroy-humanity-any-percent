package formulas

import (
	"errors"
	"math"
	"sort"
	"unicode/utf8"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/publicapi"
)

const SchemaName = "ProductionFormulasV14"

var ErrInvalidArtifact = errors.New("invalid production formula artifact")

// Schemas exports C18's owner descriptor for the existing version-14 artifact.
// Objects are closed, fields are required, and pinned is explicitly nullable.
// This validates the artifact's wire grammar, not the truth of its formulas or
// correspondence to a particular catalog; those remain the generator's job.
// Future artifact versions need their own arm, not a permissive JSON field.
func Schemas() []publicapi.NamedSchema {
	slotNames := make([]string, 0, len(multiplier.Order))
	for _, slot := range multiplier.Order {
		slotNames = append(slotNames, string(slot))
	}
	slots := enumeration(slotNames...)
	definitions := []publicapi.NamedSchema{
		{Name: SchemaName, Schema: object(
			property("axis_stack", reference("ProductionFormulaAxisStack")),
			property("commons", reference("ProductionFormulaCommons")),
			property("guild", reference("ProductionFormulaGuild")),
			property("meters", reference("ProductionFormulaMeters")),
			property("minigame_scaling", reference("ProductionFormulaMinigameScaling")),
			property("multiplier_slot_order", array(slots)),
			property("production_rate", text("")),
			property("purchasable_content", reference("ProductionFormulaPurchasableContent")),
			property("schema_version", integer(SchemaVersion, SchemaVersion)),
			property("source_fingerprint", text("sha256")),
			property("within_slot_order", enumeration(multiplier.WithinSlotOrder)),
		)},
		{Name: "ProductionFormulaAxisStack", Schema: object(
			property("factor", text("")), property("input", text("")),
			property("pinned", &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
				{Kind: publicapi.SchemaNull}, reference("ProductionFormulaAxisStackPinning"),
			}}),
			property("slot", text("")), property("stack", text("")), property("timing", text("")),
		)},
		{Name: "ProductionFormulaAxisStackPinning", Schema: object(
			property("cap_reason_key", text("mechanical-id")),
			property("input", enumeration(string(economy.AxisInputAttainmentRun), string(economy.AxisInputScoreRun), string(economy.AxisInputCloutRun))),
			property("input_cap", int64Schema()),
			property("maximum_product_at_cap", text("canonical-decimal")),
		)},
		{Name: "ProductionFormulaCommons", Schema: object(append(
			textProperties("compliance", "effective_health", "enclosure", "entry_participation_weight", "health", "modifier", "solidarity"),
			append(int64Properties(
				"cohort_health_weight_ppm", "cohort_merge_floor", "cohort_target_size", "collapse_health_ppm",
				"collective_exponent_ppm", "collective_weight_ppm", "default_tithe_ppm", "guild_health_weight_ppm",
				"health_decay_ppm_per_hour", "health_recovery_ppm_per_hour", "healthy_health_ppm", "maximum_tithe_ppm",
				"minimum_tithe_ppm", "npc_compliance_ppm", "npc_population_floor", "npc_weight_ppm", "population_tolerance_ppm",
				"server_health_weight_ppm", "solidarity_window_ms"),
				property("maximum_bonus", text("canonical-decimal")),
				property("source_weights", array(reference("ProductionFormulaSourceWeight"))))...)...),
		},
		{Name: "ProductionFormulaGuild", Schema: object(append(
			textProperties("clearing", "health", "stock_consumption", "tithe"),
			int64Properties("clearing_interval_ms", "clearing_rate_ppm", "consumption_bonus_ppm_per_unit",
				"guild_tithe_ppm", "guild_xp_target_per_founder", "npc_exchange_ppm", "stock_intake_cap")...)...),
		},
		{Name: "ProductionFormulaHardcap", Schema: object(
			property("count", int64Schema()), property("generator_id", text("mechanical-id")),
			property("reason_key", text("mechanical-id")),
		)},
		{Name: "ProductionFormulaMeters", Schema: object(textProperties(
			"attended_step", "band_events", "contribution_input", "decay", "hook_order", "ledger_fact_input")...),
		},
		{Name: "ProductionFormulaMinigameScaling", Schema: object(append(
			textProperties("fairness_gate", "fallback_rule", "grammar", "offline_grade_grammar", "offline_grade_rule",
				"payout_grammar", "payout_math", "rounding"),
			property("fallback_arms", array(enumeration(string(minigame.FallbackBot), string(minigame.FallbackNPCPartner), string(minigame.FallbackSolo)))),
			property("operation_order", array(text(""))), property("payout_order", array(text(""))))...),
		},
		{Name: "ProductionFormulaPurchasableContent", Schema: object(append(
			textProperties("ladders", "manual_output", "provisioning", "stock_rate", "synergy_linear", "synergy_log"),
			property("provision_tick_ms", int64Schema()),
			property("provisioned_hardcaps", array(reference("ProductionFormulaHardcap"))),
			property("synergy_pools", array(reference("ProductionFormulaSynergyPool"))))...),
		},
		{Name: "ProductionFormulaSourceWeight", Schema: object(
			property("forsworn", &publicapi.Schema{Kind: publicapi.SchemaBoolean}), property("slot", slots),
			property("source_id", text("mechanical-id")), property("weight_ppm", int64Schema()),
		)},
		{Name: "ProductionFormulaSynergyPool", Schema: object(
			property("curve", enumeration(string(economy.SynergyLinear), string(economy.SynergyLog))),
			property("id", text("mechanical-id")), property("slot", slots),
			property("sources", array(reference("ProductionFormulaSynergySource"))), property("target", text("")),
		)},
		{Name: "ProductionFormulaSynergySource", Schema: object(
			property("id_or_class", text("mechanical-id")),
			property("kind", enumeration(string(economy.SynergyGenerator), string(economy.SynergyUpgrade))),
			property("per_count_ppm", int64Schema()),
		)},
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	return definitions
}

// Validate rejects an unsupported version or malformed stored artifact without
// regenerating, normalizing or replacing any of its bytes.
func Validate(data []byte) error {
	if !utf8.Valid(data) {
		return ErrInvalidArtifact
	}
	definitions, err := publicapi.ValidateSchemaDefinitions(Schemas())
	if err == nil {
		err = publicapi.ValidateJSON(SchemaName, data, definitions)
	}
	if err != nil {
		return errors.Join(ErrInvalidArtifact, err)
	}
	return nil
}

func property(name string, schema *publicapi.Schema) publicapi.Field {
	return publicapi.Field{Name: name, Schema: schema, Required: true}
}
func object(fields ...publicapi.Field) *publicapi.Schema {
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
}
func text(format string) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaString, Format: format}
}
func reference(name string) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: name}
}
func array(items *publicapi.Schema) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: items}
}
func integer(minimum, maximum int64) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
}

// Signed int64 mirrors the existing generator model, not newly chosen balance
// limits. Catalog owners validate the actual numeric operating domains.
func int64Schema() *publicapi.Schema { return integer(math.MinInt64, math.MaxInt64) }
func enumeration(values ...string) *publicapi.Schema {
	sort.Strings(values)
	return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
}
func textProperties(names ...string) []publicapi.Field {
	fields := make([]publicapi.Field, 0, len(names))
	for _, name := range names {
		fields = append(fields, property(name, text("")))
	}
	return fields
}
func int64Properties(names ...string) []publicapi.Field {
	fields := make([]publicapi.Field, 0, len(names))
	for _, name := range names {
		fields = append(fields, property(name, int64Schema()))
	}
	return fields
}
