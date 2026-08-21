package account

import (
	"net/http"

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
	snapshotFields := func(version int, founderRevision bool) []publicapi.Field {
		fields := []publicapi.Field{
			apiField("constants_hash", apiString("sha256-prefixed")),
			apiField("evaluated_through_ms", integer(1, apiMaxExactInteger)),
			apiField("facts", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIFact")}),
		}
		if founderRevision {
			fields = append(fields, apiField("founder_revision", integer(1, apiMaxExactInteger)))
		}
		return append(fields,
			apiField("generators", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIGenerator")}),
			apiField("manual_action", apiRef("GameUIManualAction")),
			apiField("progress", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIProgress")}),
			apiField("resources", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIResource")}),
			apiField("revision", integer(1, apiMaxExactInteger)),
			apiField("run", apiRef("GameUIRun")),
			apiField("schema_version", integer(int64(version), int64(version))),
			apiField("server_now_ms", integer(1, apiMaxExactInteger)),
			apiField("upgrades", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GameUIUpgrade")}),
		)
	}
	return []publicapi.NamedSchema{
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
		{Name: "GameUISnapshot", Schema: apiObject(snapshotFields(2, true)...)},
		{Name: "GameUISnapshotV1", Schema: apiObject(snapshotFields(1, false)...)},
		{Name: "GameUIUpgrade", Schema: apiObject(
			apiField("cost_amount", apiString("canonical-decimal")),
			apiField("cost_resource_id", apiString("mechanical-id")),
			apiField("eligible", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
			apiField("owned", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
			apiField("upgrade_id", apiString("mechanical-id")),
		)},
	}
}

func gameUIAPIOperations() []publicapi.Operation {
	return []publicapi.Operation{{ID: "get_game_ui_snapshot", Method: http.MethodGet, Path: "/api/v1/founder/state",
		Surface: publicapi.SurfacePrivateV1, Auth: publicapi.AuthAccessToken, Responses: minigameAPIResponses("GameUISnapshot", "")}}
}
