package account

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"cloud-clicker/server/minigame"
	"cloud-clicker/server/production"

	"github.com/go-chi/chi/v5"
)

var (
	minigameOpaqueIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)
	apiMechanicalIDPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	apiUUIDPattern          = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	apiUUIDV7Pattern        = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// These handlers are mounted exclusively through the API Foundation registry.
func (api *API) createMinigameSession(response http.ResponseWriter, request *http.Request) {
	var body struct {
		IdempotencyKey string `json:"idempotency_key"`
	}
	minigameID := chi.URLParam(request, "minigame_id")
	if api.minigames == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "minigame_api")
		return
	}
	if decodeRequest(response, request, api.config.MaxBodyBytes, &body) != nil ||
		!minigameOpaqueIDPattern.MatchString(body.IdempotencyKey) || !apiMechanicalIDPattern.MatchString(minigameID) {
		writeError(response, http.StatusBadRequest, "invalid", "minigame_create")
		return
	}
	now := api.repository.clock()
	sessionID, err := newUUIDv7(now, api.repository.random)
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal_invariant", "session_id")
		return
	}
	commandID, err := newUUIDv7(now, api.repository.random)
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal_invariant", "session_id")
		return
	}
	receipt, err := api.minigames.CreateMinigameSession(request.Context(), requestClaims(request).Subject,
		minigameID, sessionID, commandID, body.IdempotencyKey, now)
	api.writeMinigameResult(response, receipt, err, "create")
}

func (api *API) playMinigameCommand(response http.ResponseWriter, request *http.Request) {
	var body struct {
		CommandID        string          `json:"command_id"`
		ExpectedRevision int64           `json:"expected_revision"`
		Command          json.RawMessage `json:"command"`
	}
	if api.minigames == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "minigame_api")
		return
	}
	if decodeRequest(response, request, api.config.MaxBodyBytes, &body) != nil ||
		!minigameOpaqueIDPattern.MatchString(body.CommandID) || body.ExpectedRevision < 1 || body.ExpectedRevision > apiMaxExactInteger ||
		!apiUUIDV7Pattern.MatchString(chi.URLParam(request, "session_id")) ||
		len(body.Command) == 0 || body.Command[0] != '{' {
		writeError(response, http.StatusBadRequest, "invalid", "minigame_command")
		return
	}
	receipt, err := api.minigames.PlayMinigameCommand(request.Context(), requestClaims(request).Subject,
		chi.URLParam(request, "session_id"), body.CommandID, body.ExpectedRevision, body.Command, api.repository.clock())
	api.writeMinigameResult(response, receipt, err, "command")
}

func (api *API) getCurrentMinigameSession(response http.ResponseWriter, request *http.Request) {
	if api.minigames == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "minigame_api")
		return
	}
	if decodeNoRequestBody(response, request, api.config.MaxBodyBytes) != nil {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	receipt, err := api.minigames.CurrentMinigameSession(request.Context(), requestClaims(request).Subject)
	api.writeMinigameResult(response, receipt, err, "current")
}

func (api *API) resolveMinigameSession(response http.ResponseWriter, request *http.Request) {
	if api.minigames == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "minigame_api")
		return
	}
	var body struct{}
	if decodeRequest(response, request, api.config.MaxBodyBytes, &body) != nil ||
		!apiUUIDV7Pattern.MatchString(chi.URLParam(request, "session_id")) {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	receipt, err := api.minigames.ResolveMinigameSession(request.Context(), requestClaims(request).Subject,
		chi.URLParam(request, "session_id"))
	api.writeMinigameResult(response, receipt, err, "resolve")
}

func decodeNoRequestBody(response http.ResponseWriter, request *http.Request, limit int64) error {
	body, err := io.ReadAll(http.MaxBytesReader(response, request.Body, limit))
	if err != nil || strings.TrimSpace(string(body)) != "" {
		return ErrInvalidRequest
	}
	return nil
}

func (api *API) writeMinigameResult(response http.ResponseWriter, receipt json.RawMessage, err error, action string) {
	if err == nil {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(receipt)
		return
	}
	var rejection *minigame.Rejection
	switch {
	case errors.As(err, &rejection):
		writeError(response, http.StatusConflict, "not_eligible", rejection.Code)
	case errors.Is(err, production.ErrMinigameFiscalUnlockRequired):
		writeError(response, http.StatusConflict, "not_eligible", "fiscal_unlock_required")
	case errors.Is(err, production.ErrMinigameHumanContentLocked):
		writeError(response, http.StatusConflict, "not_eligible", "human_content_locked")
	case errors.Is(err, minigame.ErrExclusiveActivity):
		writeError(response, http.StatusConflict, "not_eligible", "exclusive_activity")
	case errors.Is(err, minigame.ErrAPIIdempotency):
		detail := "minigame_command"
		if action == "create" {
			detail = "minigame_session"
		}
		writeError(response, http.StatusConflict, "idempotency_conflict", detail)
	case errors.Is(err, minigame.ErrSessionGone):
		writeError(response, http.StatusNotFound, "unknown_id", "minigame_session")
	case errors.Is(err, ErrFounderNotFound):
		writeError(response, http.StatusNotFound, "unknown_id", "founder")
	case errors.Is(err, minigame.ErrSessionBusy), errors.Is(err, minigame.ErrClaimLost):
		writeError(response, http.StatusConflict, "conflict", "minigame_session")
	case errors.Is(err, minigame.ErrSessionRevision):
		writeError(response, http.StatusConflict, "conflict", "minigame_revision")
	case errors.Is(err, minigame.ErrInvalidSession), errors.Is(err, production.ErrInvalidIntent):
		writeError(response, http.StatusBadRequest, "invalid", "minigame_"+action)
	case errors.Is(err, minigame.ErrTenantVersion):
		writeError(response, http.StatusNotFound, "unknown_id", "minigame_tenant")
	default:
		writeError(response, http.StatusInternalServerError, "internal_invariant", "minigame_api")
	}
}
