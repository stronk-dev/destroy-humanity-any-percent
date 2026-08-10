package account

import (
	"net/http"
	"sort"

	"cloud-clicker/server/publicapi"
)

const apiMaxExactInteger = int64(9_007_199_254_740_991)

func apiInteger(value int64) *int64 { return &value }
func apiRef(name string) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: name}
}
func apiString(format string, values ...string) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaString, Format: format, Enum: values}
}
func apiField(name string, schema *publicapi.Schema) publicapi.Field {
	return publicapi.Field{Name: name, Schema: schema, Required: true}
}
func apiObject(fields ...publicapi.Field) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
}

func minigameAPISchemas() []publicapi.NamedSchema {
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: apiInteger(minimum), Maximum: apiInteger(maximum)}
	}
	stringArray := &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiString("")}
	return []publicapi.NamedSchema{
		{Name: "APIError", Schema: apiObject(
			apiField("category", apiString("", "conflict", "idempotency_conflict", "internal_invariant", "invalid", "not_configured", "not_eligible", "rate_limited", "unauthorized", "unknown_id")),
			apiField("detail", apiString("", "access_token", "account", "body", "duplicate_card", "exclusive_activity", "fiscal_unlock_required", "founder", "founder_state", "game_ui_snapshot", "hack_slots_full", "hand_too_large", "human_content_locked", "illegal_phase", "insufficient_currency", "ip", "minigame_api", "minigame_command", "minigame_create", "minigame_revision", "minigame_session", "minigame_tenant", "recovery_progress", "recovery_session", "recovery_token", "session_id", "soul_recovery_cancel", "soul_recovery_not_ready", "soul_recovery_progress", "soul_recovery_resolve", "soul_recovery_start", "unknown_card", "unknown_offer")),
		)},
		{Name: "MinigameCommandRequest", Schema: apiObject(
			apiField("command", apiRef("MinigameTenantCommand")),
			apiField("command_id", apiString("opaque-id")),
			apiField("expected_revision", integer(1, apiMaxExactInteger)),
		)},
		{Name: "MinigameCreateRequest", Schema: apiObject(
			apiField("idempotency_key", apiString("opaque-id")),
		)},
		{Name: "MinigameCurrentActive", Schema: apiObject(
			apiField("kind", apiString("", "active")),
			apiField("session", apiRef("MinigameSessionDescriptor")),
			apiField("snapshot", apiRef("PitchSnapshot")),
		)},
		{Name: "MinigameCurrentNone", Schema: apiObject(
			apiField("kind", apiString("", "none")),
		)},
		{Name: "MinigameCurrentResponse", Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			apiRef("MinigameCurrentActive"), apiRef("MinigameCurrentNone"),
		}}},
		{Name: "MinigameEmptyRequest", Schema: apiObject()},
		{Name: "MinigameQualityChange", Schema: apiObject(
			apiField("new", apiRef("MinigameQualityState")),
			apiField("old", apiRef("MinigameQualityState")),
		)},
		{Name: "MinigameQualityState", Schema: apiObject(
			apiField("decay_remainder_ppm", integer(0, 999_999)),
			apiField("grade_ppm", integer(0, 1_000_000)),
			apiField("last_founder_attended_ms", integer(0, apiMaxExactInteger)),
		)},
		{Name: "MinigameRatingChange", Schema: apiObject(
			apiField("games_after", integer(0, apiMaxExactInteger)),
			apiField("games_before", integer(0, apiMaxExactInteger)),
			apiField("new_elo", integer(-apiMaxExactInteger, apiMaxExactInteger)),
			apiField("old_elo", integer(-apiMaxExactInteger, apiMaxExactInteger)),
			apiField("rated", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
			apiField("season_member", apiString("")),
		)},
		{Name: "MinigameResolutionReceipt", Schema: apiObject(
			apiField("cap_reason_key", apiString("")),
			apiField("certified_result_hash", apiString("sha256-prefixed")),
			apiField("company_revision", integer(1, apiMaxExactInteger)),
			apiField("configured_cap_forfeit_units", integer(0, apiMaxExactInteger)),
			apiField("credited_delta", apiString("canonical-decimal")),
			apiField("credited_resource_id", apiString("mechanical-id")),
			apiField("founder_revision", integer(1, apiMaxExactInteger)),
			apiField("intent_id", apiString("uuid-v7")),
			apiField("minigame_id", apiString("mechanical-id")),
			apiField("outcome", apiString("", "applied")),
			apiField("quality_change", apiRef("MinigameQualityChange")),
			apiField("rating_change", apiRef("MinigameRatingChange")),
			apiField("session_id", apiString("uuid-v7")),
		)},
		{Name: "MinigameSessionDescriptor", Schema: apiObject(
			apiField("constants_hash", apiString("sha256-prefixed")),
			apiField("engine_ref", apiString("mechanical-id")),
			apiField("engine_version", apiString("semver")),
			apiField("minigame_id", apiString("mechanical-id")),
			apiField("mode", apiString("", "async_snapshot", "solo")),
			apiField("revision", integer(1, apiMaxExactInteger)),
			apiField("session_id", apiString("uuid-v7")),
			apiField("status", apiString("", "active", "claimed", "resolved")),
		)},
		{Name: "MinigameSessionResponse", Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			apiRef("MinigameSessionResponseActive"), apiRef("MinigameSessionResponseTerminal"),
		}}},
		{Name: "MinigameSessionResponseActive", Schema: apiObject(
			apiField("constants_hash", apiString("sha256-prefixed")),
			apiField("engine_ref", apiString("mechanical-id")),
			apiField("engine_version", apiString("semver")),
			apiField("minigame_id", apiString("mechanical-id")),
			apiField("mode", apiString("", "async_snapshot", "solo")),
			apiField("revision", integer(1, apiMaxExactInteger)),
			apiField("session_id", apiString("uuid-v7")),
			apiField("snapshot", apiRef("PitchSnapshot")),
			apiField("status", apiString("", "active")),
		)},
		{Name: "MinigameSessionResponseTerminal", Schema: apiObject(
			apiField("constants_hash", apiString("sha256-prefixed")),
			apiField("engine_ref", apiString("mechanical-id")),
			apiField("engine_version", apiString("semver")),
			apiField("minigame_id", apiString("mechanical-id")),
			apiField("mode", apiString("", "async_snapshot", "solo")),
			apiField("resolution_receipt", apiRef("MinigameResolutionReceipt")),
			apiField("revision", integer(1, apiMaxExactInteger)),
			apiField("session_id", apiString("uuid-v7")),
			apiField("snapshot", apiRef("PitchSnapshot")),
			apiField("status", apiString("", "resolved")),
		)},
		{Name: "MinigameTenantCommand", Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			apiRef("PitchBuyHackCommand"), apiRef("PitchEndShopCommand"), apiRef("PitchPlayHandCommand"),
		}}},
		{Name: "PitchBuyHackCommand", Schema: apiObject(
			apiField("kind", apiString("", "buy_hack")),
			apiField("offer_id", apiString("")),
		)},
		{Name: "PitchEndShopCommand", Schema: apiObject(
			apiField("kind", apiString("", "end_shop")),
		)},
		{Name: "PitchPlayHandCommand", Schema: apiObject(
			apiField("card_ids", stringArray),
			apiField("kind", apiString("", "play_hand")),
		)},
		{Name: "PitchShopOffer", Schema: apiObject(
			apiField("hack_id", apiString("mechanical-id")),
			apiField("offer_id", apiString("")),
			apiField("price", integer(0, apiMaxExactInteger)),
		)},
		{Name: "PitchSnapshot", Schema: apiObject(
			apiField("deck_count", integer(0, apiMaxExactInteger)),
			apiField("funding_target", apiString("canonical-decimal")),
			apiField("hand", stringArray),
			apiField("hands_remaining", integer(0, apiMaxExactInteger)),
			apiField("phase", apiString("", "playing", "shop", "terminal")),
			apiField("pitch_content_hash", apiString("sha256-prefixed")),
			apiField("pitch_schema_version", integer(1, apiMaxExactInteger)),
			apiField("revision", integer(1, apiMaxExactInteger)),
			apiField("round", integer(1, apiMaxExactInteger)),
			apiField("round_best_valuation", apiString("canonical-decimal")),
			apiField("run_currency", integer(0, apiMaxExactInteger)),
			apiField("shop_offers", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("PitchShopOffer")}),
			apiField("slotted_hacks", stringArray),
		)},
	}
}

func minigameAPIResponses(success string) []publicapi.Response {
	return []publicapi.Response{
		{Kind: publicapi.ResponseSchema, Status: http.StatusOK, ContentType: publicapi.ContentJSON, SchemaRef: success},
		{Kind: publicapi.ResponseSchema, Status: http.StatusBadRequest, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
		{Kind: publicapi.ResponseSchema, Status: http.StatusUnauthorized, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
		{Kind: publicapi.ResponseSchema, Status: http.StatusNotFound, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
		{Kind: publicapi.ResponseSchema, Status: http.StatusConflict, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
		{Kind: publicapi.ResponseSchema, Status: http.StatusTooManyRequests, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
		{Kind: publicapi.ResponseSchema, Status: http.StatusInternalServerError, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
		{Kind: publicapi.ResponseSchema, Status: http.StatusServiceUnavailable, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
	}
}

func minigameAPIOperations() []publicapi.Operation {
	return []publicapi.Operation{
		{ID: "create_minigame_session", Method: http.MethodPost, Path: "/api/v1/minigames/{minigame_id}/sessions", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Parameters: []publicapi.Parameter{{Name: "minigame_id", Schema: apiString("mechanical-id")}},
			Request: "MinigameCreateRequest", Responses: minigameAPIResponses("MinigameSessionResponseActive")},
		{ID: "get_current_minigame_session", Method: http.MethodGet, Path: "/api/v1/minigames/sessions/current", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Responses: minigameAPIResponses("MinigameCurrentResponse")},
		{ID: "play_minigame_command", Method: http.MethodPost, Path: "/api/v1/minigames/sessions/{session_id}/commands", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Parameters: []publicapi.Parameter{{Name: "session_id", Schema: apiString("uuid-v7")}},
			Request: "MinigameCommandRequest", Responses: minigameAPIResponses("MinigameSessionResponse")},
		{ID: "resolve_minigame_session", Method: http.MethodPost, Path: "/api/v1/minigames/sessions/{session_id}/resolve", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Parameters: []publicapi.Parameter{{Name: "session_id", Schema: apiString("uuid-v7")}},
			Request: "MinigameEmptyRequest", Responses: minigameAPIResponses("MinigameSessionResponseTerminal")},
	}
}

func newPrivateAPIRegistry() (*publicapi.Registry, error) {
	schemas := append(minigameAPISchemas(), soulRecoveryAPISchemas()...)
	schemas = append(schemas, gameUIAPISchemas()...)
	operations := append(minigameAPIOperations(), soulRecoveryAPIOperations()...)
	operations = append(operations, gameUIAPIOperations()...)
	sort.Slice(schemas, func(left, right int) bool { return schemas[left].Name < schemas[right].Name })
	sort.Slice(operations, func(left, right int) bool { return operations[left].ID < operations[right].ID })
	return publicapi.NewRegistry(schemas, operations)
}

// PrivateAPIRegistry returns the immutable authority consumed by runtime
// mounting and generated API artifacts.
func PrivateAPIRegistry() (*publicapi.Registry, error) { return newPrivateAPIRegistry() }
