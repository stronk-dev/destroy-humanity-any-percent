package publicapi

import (
	"bytes"
	"net/http"
	"sort"
	"time"

	"cloud-clicker/server/httpapi"
)

// One authority for this API-wide pre-handler failure: operation declarations
// and actual middleware consume the same literal, not independently written
// tables. No internal error text or invented correlation ID reaches the wire.
const requestIDFailureJSON = "{\"category\":\"internal_invariant\",\"detail\":\"request_id\"}\n"

func RequestIDFailureResponse() Response {
	return Response{Kind: ResponseSchema, Status: http.StatusInternalServerError, ContentType: ContentJSON,
		SchemaRef: "APIError", ExactJSON: [][]byte{[]byte(requestIDFailureJSON)}}
}

func writeRequestIDFailure(response http.ResponseWriter, _ *http.Request) {
	declaration := RequestIDFailureResponse()
	for _, header := range []string{"X-Request-ID", "ETag", "X-Content-SHA256", "Content-Length", "Content-Encoding"} {
		response.Header().Del(header)
	}
	response.Header().Set("Content-Type", declaration.ContentType)
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(declaration.Status)
	_, _ = response.Write(declaration.ExactJSON[0])
}

// WithRequestID binds shared HTTP mechanics to the API's declared refusal.
// It owns no rate bucket, authentication or application work.
func WithRequestID(next http.Handler, ids *httpapi.RequestIDs, clock func() time.Time) http.Handler {
	return ids.WithRequestIDFailure(next, clock, http.HandlerFunc(writeRequestIDFailure))
}

// BindRequestIDFailure adds the pre-handler failure to each operation, retaining
// existing broad or exact error domains. Return a new immutable registry; never
// narrow an existing broad500 or mutate the caller's declarations.
func BindRequestIDFailure(registry *Registry) (*Registry, error) {
	if registry == nil {
		return nil, ErrInvalidOperation
	}
	operations := registry.Operations()
	declaration := RequestIDFailureResponse()
	for index := range operations {
		responses := operations[index].Responses
		found := false
		for responseIndex := range responses {
			response := &responses[responseIndex]
			// C19 permits gzip and JSON at the same status. Only the JSON
			// representation can overlap this middleware-owned declaration.
			if response.Status != declaration.Status || response.ContentType != declaration.ContentType {
				continue
			}
			if response.Kind != ResponseSchema || response.ContentType != ContentJSON || response.SchemaRef != declaration.SchemaRef {
				return nil, ErrInvalidOperation
			}
			found = true
			if len(response.ExactJSON) != 0 {
				hasLiteral := false
				for _, body := range response.ExactJSON {
					hasLiteral = hasLiteral || bytes.Equal(body, declaration.ExactJSON[0])
				}
				if !hasLiteral {
					response.ExactJSON = append(response.ExactJSON, bytes.Clone(declaration.ExactJSON[0]))
					sort.Slice(response.ExactJSON, func(i, j int) bool { return bytes.Compare(response.ExactJSON[i], response.ExactJSON[j]) < 0 })
				}
			}
		}
		if !found {
			responses = append(responses, RequestIDFailureResponse())
			sort.Slice(responses, func(i, j int) bool {
				if responses[i].Status != responses[j].Status {
					return responses[i].Status < responses[j].Status
				}
				return responses[i].ContentType < responses[j].ContentType
			})
		}
		operations[index].Responses = responses
	}
	return NewRegistry(registry.Schemas(), operations)
}

// APIErrorSchema is the single API-wide typed-rejection descriptor (A4/A5).
// Public and private registries share it without making public readers import
// the private account repository. Exact operation/status pairs remain owned by
// each operation; the two enums alone intentionally are not a pair oracle.
func APIErrorSchema() NamedSchema {
	return NamedSchema{Name: "APIError", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
		{Name: "category", Required: true, Schema: &Schema{Kind: SchemaString, Enum: []string{
			"conflict", "idempotency_conflict", "internal_invariant", "invalid", "not_configured", "not_eligible", "rate_limited", "refresh_reused", "unauthorized", "unknown_id",
		}}},
		{Name: "detail", Required: true, Schema: &Schema{Kind: SchemaString, Enum: []string{
			"access_token", "account", "account_create", "body", "bootstrap", "bootstrap_expired", "category", "company_stream", "credential", "curriculum_exit_required", "cursor", "duplicate_card", "epoch", "exclusive_activity", "fiscal_unlock_required", "founder", "founder_create", "founder_state", "game_ui_snapshot", "garden", "hack_slots_full", "hand_too_large", "human_content_locked", "illegal_phase", "insufficient_currency", "invalid_assist_level", "invalid_text", "ip", "limit", "line_too_long", "mandate", "minigame_api", "minigame_command", "minigame_create", "minigame_revision", "minigame_session", "minigame_tenant", "public_api", "recovery_progress", "recovery_session", "recovery_token", "refresh_token", "request_id", "run", "session_family_revoked", "session_id", "soul_recovery", "soul_recovery_cancel", "soul_recovery_not_ready", "soul_recovery_progress", "soul_recovery_resolve", "soul_recovery_start", "tier_required", "unknown_card", "unknown_offer", "variables",
		}}},
	}}}
}
