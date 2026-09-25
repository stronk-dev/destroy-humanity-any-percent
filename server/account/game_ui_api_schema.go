package account

import (
	"net/http"
	"sort"

	"cloud-clicker/server/publicapi"
)

func gameUIAPISchemas() []publicapi.NamedSchema {
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: apiInteger(minimum), Maximum: apiInteger(maximum)}
	}
	nullableCap := &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{apiRef("GameUIResourceCap"), {Kind: publicapi.SchemaNull}}}
	factValue := &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
		{Kind: publicapi.SchemaBoolean}, integer(-apiMaxExactInteger, apiMaxExactInteger), apiString(""),
	}}
	nullable := func(schema *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{schema, {Kind: publicapi.SchemaNull}}}
	}
	boolean := &publicapi.Schema{Kind: publicapi.SchemaBoolean}
	array := func(item string) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef(item)}
	}
	snapshotFields := func(version int, founderRevision, transitions bool) []publicapi.Field {
		fields := []publicapi.Field{
			apiField("constants_hash", apiString("sha256-prefixed")),
			apiField("evaluated_through_ms", integer(1, apiMaxExactInteger)),
			apiField("facts", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIFact")}),
		}
		if version >= 4 {
			fields = append(fields, apiField("features", apiRef("GameUIFeatures")))
		}
		if founderRevision {
			fields = append(fields, apiField("founder_revision", integer(1, apiMaxExactInteger)))
		}
		fields = append(fields,
			apiField("generators", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef(map[bool]string{true: "GameUIGeneratorV4", false: "GameUIGenerator"}[version >= 4])}),
			apiField("manual_action", apiRef("GameUIManualAction")),
			apiField("progress", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIProgress")}),
			apiField("resources", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIResource")}),
			apiField("revision", integer(1, apiMaxExactInteger)),
			apiField("run", apiRef("GameUIRun")),
			apiField("schema_version", integer(int64(version), int64(version))),
			apiField("server_now_ms", integer(1, apiMaxExactInteger)),
		)
		if transitions {
			fields = append(fields, apiField("transitions", apiRef("GameUITransitions")))
		}
		return append(fields, apiField("upgrades", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIUpgrade")}))
	}
	schemas := []publicapi.NamedSchema{
		{Name: "GameUIFact", Schema: apiObject(
			apiField("fact_id", apiString("mechanical-id")),
			apiField("value", factValue),
		)},
		{Name: "GameUIGenerator", Schema: apiObject(
			apiField("generator_id", apiString("mechanical-id")),
			apiField("max_affordable", integer(0, apiMaxExactInteger)),
			apiField("next_cost", apiString("canonical-decimal")),
			apiField("next_cost_resource_id", apiString("mechanical-id")),
			apiField("owned", integer(0, apiMaxExactInteger)),
			apiField("provisioned", integer(0, apiMaxExactInteger)),
			apiField("rate_contribution", apiString("canonical-decimal")),
		)},
		{Name: "GameUIGeneratorV4", Schema: apiObject(
			apiField("generator_id", apiString("mechanical-id")),
			apiField("max_affordable", integer(0, apiMaxExactInteger)),
			apiField("next_cost", apiString("canonical-decimal")),
			apiField("next_cost_resource_id", apiString("mechanical-id")),
			apiField("owned", integer(0, apiMaxExactInteger)),
			apiField("provision_cap", nullable(apiRef("GameUIIntCap"))),
			apiField("provisioned", integer(0, apiMaxExactInteger)),
			apiField("rate_contribution", apiString("canonical-decimal")),
		)},
		{Name: "GameUIIntCap", Schema: apiObject(
			apiField("amount", integer(0, apiMaxExactInteger)),
			apiField("reason_key", apiString("mechanical-id")),
		)},
		{Name: "GameUIFeatures", Schema: apiObject(
			apiField("achievements", nullable(apiRef("GameUIAchievementsArm"))),
			apiField("active_play", &publicapi.Schema{Kind: publicapi.SchemaNull}),
			apiField("fiscal", nullable(apiRef("GameUIFiscalArm"))),
			apiField("meters", nullable(apiRef("GameUIMetersArm"))),
			apiField("minigames", nullable(apiRef("GameUIMinigamesArm"))),
			apiField("pets", &publicapi.Schema{Kind: publicapi.SchemaNull}),
		)},
		{Name: "GameUIAchievementsArm", Schema: apiObject(
			apiField("rows", array("GameUIAchievementRow")),
			apiField("score", apiObject(
				apiField("lifetime", integer(0, apiMaxExactInteger)),
				apiField("run", integer(0, apiMaxExactInteger)),
			)),
		)},
		{Name: "GameUIAchievementRow", Schema: apiObject(
			apiField("achievement_id", apiString("mechanical-id")),
			apiField("condition_scope", apiString("", "career", "run")),
			apiField("copy_key", apiString("mechanical-id")),
			apiField("earned", nullable(apiString("", "lifetime", "run"))),
			apiField("proof_kind", apiString("", "burn", "possession", "provenance")),
			apiField("score_grant", integer(0, apiMaxExactInteger)),
		)},
		{Name: "GameUIMetersArm", Schema: apiObject(apiField("meters", array("GameUIMeterRow")))},
		{Name: "GameUIMeterRow", Schema: apiObject(
			apiField("band_id", apiString("mechanical-id")),
			apiField("bands", array("GameUIMeterBand")),
			apiField("max", integer(0, 100)),
			apiField("meter_id", apiString("mechanical-id")),
			apiField("min", integer(0, 100)),
			apiField("value", integer(0, 100)),
		)},
		{Name: "GameUIMeterBand", Schema: apiObject(
			apiField("band_id", apiString("mechanical-id")),
			apiField("floor_value", integer(0, 100)),
		)},
		{Name: "GameUIFiscalArm", Schema: apiObject(
			apiField("credit", integer(0, apiMaxExactInteger)),
			apiField("credit_cap", apiRef("GameUIIntCap")),
			apiField("credit_per_period", integer(0, apiMaxExactInteger)),
			apiField("generator_levels", array("GameUIFiscalLevel")),
			apiField("hoard", apiObject(
				apiField("cap_credits", integer(0, apiMaxExactInteger)),
				apiField("preview_ppm", integer(0, apiMaxExactInteger)),
				apiField("reason_note", apiString("", "next_run")),
			)),
			apiField("period", apiObject(
				apiField("auto_ms", integer(1, apiMaxExactInteger)),
				apiField("early_ms", integer(0, apiMaxExactInteger)),
				apiField("early_success_ppm", integer(0, 1_000_000)),
				apiField("guaranteed_ms", integer(0, apiMaxExactInteger)),
				apiField("opened_wall_ms", integer(0, apiMaxExactInteger)),
				apiField("seq", integer(0, apiMaxExactInteger)),
			)),
			apiField("sweep_preview", apiObject(
				apiField("credit_after", integer(0, apiMaxExactInteger)),
				apiField("credited", integer(0, apiMaxExactInteger)),
				apiField("periods", integer(0, apiMaxExactInteger)),
				apiField("saturated", boolean),
			)),
			apiField("unlocks", array("GameUIFiscalUnlock")),
		)},
		{Name: "GameUIFiscalLevel", Schema: apiObject(
			apiField("generator_id", apiString("mechanical-id")),
			apiField("level", integer(0, apiMaxExactInteger)),
			apiField("level_cap", apiRef("GameUIIntCap")),
			apiField("next_level_cost", nullable(integer(0, apiMaxExactInteger))),
			apiField("ppm_per_level", integer(0, apiMaxExactInteger)),
		)},
		{Name: "GameUIFiscalUnlock", Schema: apiObject(
			apiField("cost", integer(0, apiMaxExactInteger)),
			apiField("owned", boolean),
			apiField("unlock_id", apiString("mechanical-id")),
		)},
		{Name: "GameUIMinigamesArm", Schema: apiObject(apiField("rows", array("GameUIMinigameAvailability")))},
		{Name: "GameUIMinigameAvailability", Schema: apiObject(
			apiField("active_session", boolean),
			apiField("human_content_locked", boolean),
			apiField("minigame_id", apiString("mechanical-id")),
			apiField("unlocked", boolean),
		)},
		{Name: "GameUIManualAction", Schema: apiObject(
			apiField("action_id", apiString("mechanical-id")),
			apiField("bucket_cap_milli", integer(1, apiMaxExactInteger)),
			apiField("refill_milli_per_ms", integer(1, apiMaxExactInteger)),
			apiField("refilled_at_ms", integer(1, apiMaxExactInteger)),
			apiField("tokens_milli", integer(0, apiMaxExactInteger)),
		)},
		{Name: "GameUIProgress", Schema: apiObject(
			apiField("current", apiString("canonical-decimal")),
			apiField("stage_id", apiString("mechanical-id")),
			apiField("target", apiString("canonical-decimal")),
		)},
		{Name: "GameUIResource", Schema: apiObject(
			apiField("amount", apiString("canonical-decimal")),
			apiField("cap", nullableCap),
			apiField("rate_per_second", apiString("canonical-decimal")),
			apiField("resource_id", apiString("mechanical-id")),
		)},
		{Name: "GameUIResourceCap", Schema: apiObject(
			apiField("amount", apiString("canonical-decimal")),
			apiField("reason_key", apiString("mechanical-id")),
		)},
		{Name: "GameUIRun", Schema: apiObject(
			apiField("category", apiString("mechanical-id")),
			apiField("exit_count", integer(0, apiMaxExactInteger)),
			apiField("founder_id", apiString("uuid")),
			apiField("run_seq", integer(1, apiMaxExactInteger)),
			apiField("run_started_at_ms", integer(1, apiMaxExactInteger)),
			apiField("tier", integer(0, 9)),
		)},
		{Name: "GameUISnapshot", Schema: apiObject(snapshotFields(4, true, true)...)},
		{Name: "GameUISnapshotV3", Schema: apiObject(snapshotFields(3, true, true)...)},
		{Name: "GameUISnapshotV1", Schema: apiObject(snapshotFields(1, false, false)...)},
		{Name: "GameUISnapshotV2", Schema: apiObject(snapshotFields(2, true, false)...)},
		{Name: "GameUITransitionCrossGate", Schema: apiObject(
			apiField("eligible", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
			apiField("gate_id", apiString("mechanical-id")),
			apiField("route_id", &publicapi.Schema{Kind: publicapi.SchemaNull}),
		)},
		{Name: "GameUITransitionEligibility", Schema: apiObject(
			apiField("eligible", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
		)},
		{Name: "GameUITransitions", Schema: apiObject(
			apiField("cross_gate", &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{apiRef("GameUITransitionCrossGate"), {Kind: publicapi.SchemaNull}}}),
			apiField("wind_down", apiRef("GameUITransitionEligibility")),
		)},
		{Name: "GameUIUpgrade", Schema: apiObject(
			apiField("cost_amount", apiString("canonical-decimal")),
			apiField("cost_resource_id", apiString("mechanical-id")),
			apiField("eligible", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
			apiField("owned", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
			apiField("upgrade_id", apiString("mechanical-id")),
		)},
	}
	sort.Slice(schemas, func(left, right int) bool { return schemas[left].Name < schemas[right].Name })
	return schemas
}

func gameUIAPIOperations() []publicapi.Operation {
	return []publicapi.Operation{{ID: "get_game_ui_snapshot", Method: http.MethodGet, Path: "/api/v1/founder/state",
		Surface: publicapi.SurfacePrivateV1, Auth: publicapi.AuthAccessToken, Responses: minigameAPIResponses("GameUISnapshot", "")}}
}
