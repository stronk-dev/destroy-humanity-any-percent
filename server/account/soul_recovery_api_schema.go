package account

import (
	"net/http"

	"cloud-clicker/server/publicapi"
)

func soulRecoveryAPISchemas() []publicapi.NamedSchema {
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: apiInteger(minimum), Maximum: apiInteger(maximum)}
	}
	return []publicapi.NamedSchema{
		{Name: "SoulRecoveryFinishRequest", Schema: apiObject(
			apiField("session_id", apiString("uuid-v7")),
		)},
		{Name: "SoulRecoveryProgressRequest", Schema: apiObject(
			apiField("progress_token", apiString("uuid")),
			apiField("session_id", apiString("uuid-v7")),
		)},
		{Name: "SoulRecoveryProgressResponse", Schema: apiObject(
			apiField("attended_progress_ms", integer(0, apiMaxExactInteger)),
			apiField("eligible", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
			apiField("last_progress_server_ms", integer(0, apiMaxExactInteger)),
			apiField("required_duration_attended_ms", integer(1, apiMaxExactInteger)),
			apiField("session_id", apiString("uuid-v7")),
		)},
		{Name: "SoulRecoveryStartRequest", Schema: apiObject(
			apiField("activity_id", apiString("mechanical-id")),
		)},
		{Name: "SoulRecoveryStartResponse", Schema: apiObject(
			apiField("activity_id", apiString("mechanical-id")),
			apiField("attended_progress_ms", integer(0, apiMaxExactInteger)),
			apiField("last_progress_server_ms", integer(0, apiMaxExactInteger)),
			apiField("progress_token", apiString("uuid")),
			apiField("required_duration_attended_ms", integer(1, apiMaxExactInteger)),
			apiField("session_id", apiString("uuid-v7")),
			apiField("started_wall_ms", integer(0, apiMaxExactInteger)),
		)},
		{Name: "SoulRecoveryTerminalResponse", Schema: apiObject(
			apiField("action", apiString("", "cancel", "resolve")),
			apiField("activity_id", apiString("mechanical-id")),
			apiField("band_after", apiString("", "dimming", "hollow", "near_zero", "whole")),
			apiField("band_before", apiString("", "dimming", "hollow", "near_zero", "whole")),
			publicapi.Field{Required: false, Name: "cancelled_by", Schema: apiString("", "player", "watchdog")},
			apiField("company_revision", integer(1, apiMaxExactInteger)),
			apiField("founder_revision", integer(1, apiMaxExactInteger)),
			apiField("intent_id", apiString("uuid-v7")),
			apiField("outcome", apiString("", "applied")),
			apiField("session_id", apiString("uuid-v7")),
			apiField("soul_after", integer(0, 100)),
			apiField("soul_before", integer(0, 100)),
		)},
	}
}

func soulRecoveryAPIResponses(success string, rateLimited bool) []publicapi.Response {
	_ = rateLimited // every authenticated route can hit the shared account limiter
	return minigameAPIResponses(success, "")
}

func soulRecoveryAPIOperations() []publicapi.Operation {
	return []publicapi.Operation{
		{ID: "cancel_soul_recovery", Method: http.MethodPost, Path: "/api/v1/soul-recovery/cancel", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Request: "SoulRecoveryFinishRequest", Responses: soulRecoveryAPIResponses("SoulRecoveryTerminalResponse", false)},
		{ID: "progress_soul_recovery", Method: http.MethodPost, Path: "/api/v1/soul-recovery/progress", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Request: "SoulRecoveryProgressRequest", Responses: soulRecoveryAPIResponses("SoulRecoveryProgressResponse", true)},
		{ID: "resolve_soul_recovery", Method: http.MethodPost, Path: "/api/v1/soul-recovery/resolve", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Request: "SoulRecoveryFinishRequest", Responses: soulRecoveryAPIResponses("SoulRecoveryTerminalResponse", false)},
		{ID: "start_soul_recovery", Method: http.MethodPost, Path: "/api/v1/soul-recovery/start", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Request: "SoulRecoveryStartRequest", Responses: soulRecoveryAPIResponses("SoulRecoveryStartResponse", false)},
	}
}
