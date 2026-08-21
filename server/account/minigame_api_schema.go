package account

import (
	"bytes"
	"net/http"
	"sort"
	"strconv"

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
			apiField("detail", apiString("", "access_token", "account", "body", "bootstrap", "bootstrap_expired", "duplicate_card", "exclusive_activity", "fiscal_unlock_required", "founder", "founder_state", "game_ui_snapshot", "hack_slots_full", "hand_too_large", "human_content_locked", "illegal_phase", "insufficient_currency", "ip", "minigame_api", "minigame_command", "minigame_create", "minigame_revision", "minigame_session", "minigame_tenant", "recovery_progress", "recovery_session", "recovery_token", "session_id", "soul_recovery_cancel", "soul_recovery_not_ready", "soul_recovery_progress", "soul_recovery_resolve", "soul_recovery_start", "unknown_card", "unknown_offer")),
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

type apiErrorPair struct {
	category string
	detail   string
}

func exactAPIErrorJSON(pairs ...apiErrorPair) [][]byte {
	result := make([][]byte, len(pairs))
	for index, pair := range pairs {
		result[index] = []byte(`{"category":` + strconv.Quote(pair.category) + `,"detail":` + strconv.Quote(pair.detail) + "}\n")
	}
	sort.Slice(result, func(left, right int) bool { return bytes.Compare(result[left], result[right]) < 0 })
	return result
}

func minigameErrorJSON(action string, status int) [][]byte {
	if action == "" {
		return nil
	}
	switch status {
	case http.StatusBadRequest:
		detail := "body"
		if action == "create" || action == "command" {
			detail = "minigame_" + action
		}
		return exactAPIErrorJSON(apiErrorPair{"invalid", detail})
	case http.StatusUnauthorized:
		return exactAPIErrorJSON(apiErrorPair{"unauthorized", "access_token"})
	case http.StatusNotFound:
		return exactAPIErrorJSON(
			apiErrorPair{"unknown_id", "founder"},
			apiErrorPair{"unknown_id", "minigame_session"},
			apiErrorPair{"unknown_id", "minigame_tenant"},
		)
	case http.StatusConflict:
		idempotencyDetail := "minigame_command"
		if action == "create" {
			idempotencyDetail = "minigame_session"
		}
		return exactAPIErrorJSON(
			apiErrorPair{"conflict", "minigame_revision"},
			apiErrorPair{"conflict", "minigame_session"},
			apiErrorPair{"idempotency_conflict", idempotencyDetail},
			apiErrorPair{"not_eligible", "duplicate_card"},
			apiErrorPair{"not_eligible", "exclusive_activity"},
			apiErrorPair{"not_eligible", "fiscal_unlock_required"},
			apiErrorPair{"not_eligible", "hack_slots_full"},
			apiErrorPair{"not_eligible", "hand_too_large"},
			apiErrorPair{"not_eligible", "human_content_locked"},
			apiErrorPair{"not_eligible", "illegal_phase"},
			apiErrorPair{"not_eligible", "insufficient_currency"},
			apiErrorPair{"not_eligible", "unknown_card"},
			apiErrorPair{"not_eligible", "unknown_offer"},
		)
	case http.StatusTooManyRequests:
		return exactAPIErrorJSON(
			apiErrorPair{"rate_limited", "account"},
			apiErrorPair{"rate_limited", "ip"},
		)
	case http.StatusInternalServerError:
		pairs := []apiErrorPair{{"internal_invariant", "minigame_api"}}
		if action == "create" {
			pairs = append(pairs, apiErrorPair{"internal_invariant", "session_id"})
		}
		return exactAPIErrorJSON(pairs...)
	case http.StatusServiceUnavailable:
		return exactAPIErrorJSON(apiErrorPair{"not_configured", "minigame_api"})
	default:
		return nil
	}
}

func minigameAPIResponses(success, action string) []publicapi.Response {
	return []publicapi.Response{
		{Kind: publicapi.ResponseSchema, Status: http.StatusOK, ContentType: publicapi.ContentJSON, SchemaRef: success},
		{Kind: publicapi.ResponseSchema, Status: http.StatusBadRequest, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: minigameErrorJSON(action, http.StatusBadRequest)},
		{Kind: publicapi.ResponseSchema, Status: http.StatusUnauthorized, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: minigameErrorJSON(action, http.StatusUnauthorized)},
		{Kind: publicapi.ResponseSchema, Status: http.StatusNotFound, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: minigameErrorJSON(action, http.StatusNotFound)},
		{Kind: publicapi.ResponseSchema, Status: http.StatusConflict, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: minigameErrorJSON(action, http.StatusConflict)},
		{Kind: publicapi.ResponseSchema, Status: http.StatusTooManyRequests, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: minigameErrorJSON(action, http.StatusTooManyRequests)},
		{Kind: publicapi.ResponseSchema, Status: http.StatusInternalServerError, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: minigameErrorJSON(action, http.StatusInternalServerError)},
		{Kind: publicapi.ResponseSchema, Status: http.StatusServiceUnavailable, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: minigameErrorJSON(action, http.StatusServiceUnavailable)},
	}
}

func minigameAPIOperations() []publicapi.Operation {
	return []publicapi.Operation{
		{ID: "create_minigame_session", Method: http.MethodPost, Path: "/api/v1/minigames/{minigame_id}/sessions", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Parameters: []publicapi.Parameter{{Name: "minigame_id", Schema: apiString("mechanical-id")}},
			Request: "MinigameCreateRequest", Responses: minigameAPIResponses("MinigameSessionResponseActive", "create")},
		{ID: "get_current_minigame_session", Method: http.MethodGet, Path: "/api/v1/minigames/sessions/current", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Responses: minigameAPIResponses("MinigameCurrentResponse", "current")},
		{ID: "play_minigame_command", Method: http.MethodPost, Path: "/api/v1/minigames/sessions/{session_id}/commands", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Parameters: []publicapi.Parameter{{Name: "session_id", Schema: apiString("uuid-v7")}},
			Request: "MinigameCommandRequest", Responses: minigameAPIResponses("MinigameSessionResponse", "command")},
		{ID: "resolve_minigame_session", Method: http.MethodPost, Path: "/api/v1/minigames/sessions/{session_id}/resolve", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Parameters: []publicapi.Parameter{{Name: "session_id", Schema: apiString("uuid-v7")}},
			Request: "MinigameEmptyRequest", Responses: minigameAPIResponses("MinigameSessionResponseTerminal", "resolve")},
	}
}

func newPrivateAPIRegistry() (*publicapi.Registry, error) {
	schemas := append(minigameAPISchemas(), soulRecoveryAPISchemas()...)
	schemas = append(schemas, gameUIAPISchemas()...)
	schemas = append(schemas, bootstrapAPISchemas()...)
	operations := append(minigameAPIOperations(), soulRecoveryAPIOperations()...)
	operations = append(operations, gameUIAPIOperations()...)
	operations = append(operations, bootstrapAPIOperations()...)
	sort.Slice(schemas, func(left, right int) bool { return schemas[left].Name < schemas[right].Name })
	sort.Slice(operations, func(left, right int) bool { return operations[left].ID < operations[right].ID })
	return publicapi.NewRegistry(schemas, operations)
}

// PrivateAPIRegistry returns the immutable authority consumed by runtime
// mounting and generated API artifacts.
func PrivateAPIRegistry() (*publicapi.Registry, error) { return newPrivateAPIRegistry() }
