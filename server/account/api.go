package account

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"cloud-clicker/server/httpapi"
	"cloud-clicker/server/production"
	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"

	"github.com/go-chi/chi/v5"
)

type IntentHandler interface {
	Handle(context.Context, string, production.EvaluationMode, time.Time, []byte) (production.HandleResult, error)
}

type GuildIntentHandler interface {
	HandleGuild(context.Context, string, []byte) (json.RawMessage, bool, error)
	IsInvalidGuildIntent(error) bool
}

type SoulRecoveryHandler interface {
	StartSoulRecovery(context.Context, production.StartSoulRecoveryRequest, time.Time) (production.HandleResult, error)
	ProgressSoulRecovery(context.Context, production.ProgressSoulRecoveryRequest, time.Time, save.ExitFaultInjector) (production.HandleResult, error)
	CancelSoulRecovery(context.Context, production.FinishSoulRecoveryRequest, time.Time, save.ExitFaultInjector) (production.HandleResult, error)
	ResolveSoulRecovery(context.Context, production.FinishSoulRecoveryRequest, time.Time, save.ExitFaultInjector) (production.HandleResult, error)
	SoulRecoveryBeatCeilingMS(context.Context, string, string) (int64, error)
}

type MinigameAPIHandler interface {
	CreateMinigameSession(context.Context, string, string, string, string, string, time.Time) (json.RawMessage, error)
	PlayMinigameCommand(context.Context, string, string, string, int64, json.RawMessage, time.Time) (json.RawMessage, error)
	CurrentMinigameSession(context.Context, string) (json.RawMessage, error)
	ResolveMinigameSession(context.Context, string, string) (json.RawMessage, error)
}

type GameUIHandler interface {
	GameUISnapshot(context.Context, string, time.Time) (json.RawMessage, error)
}

type APIConfig struct {
	UnauthenticatedBurst  int
	UnauthenticatedPerMin int
	AccountBurst          int
	AccountPerMin         int
	MaxBodyBytes          int64
	TrustedProxyHops      int
	LimiterMaxEntries     int
}

func Phase0APIConfig() APIConfig {
	return APIConfig{UnauthenticatedBurst: 10, UnauthenticatedPerMin: 30, AccountBurst: 60, AccountPerMin: 300,
		MaxBodyBytes: 64 << 10, LimiterMaxEntries: 65_536}
}

type API struct {
	repository       *Repository
	intents          IntentHandler
	guilds           GuildIntentHandler
	recoveries       SoulRecoveryHandler
	minigames        MinigameAPIHandler
	gameUI           GameUIHandler
	privateRegistry  *publicapi.Registry
	config           APIConfig
	unauth           *httpapi.TokenBuckets
	accounts         *httpapi.TokenBuckets
	recoveryProgress *recoveryBuckets
}

func (api *API) AttachGuildIntents(handler GuildIntentHandler) error {
	if api == nil || handler == nil || api.guilds != nil {
		return ErrInvalidRequest
	}
	api.guilds = handler
	return nil
}

func (api *API) AttachSoulRecoveries(handler SoulRecoveryHandler) error {
	if api == nil || handler == nil || api.recoveries != nil {
		return ErrInvalidRequest
	}
	api.recoveries = handler
	return nil
}

func (api *API) AttachMinigames(handler MinigameAPIHandler) error {
	if api == nil || handler == nil || api.minigames != nil {
		return ErrInvalidRequest
	}
	api.minigames = handler
	return nil
}

func (api *API) AttachGameUI(handler GameUIHandler) error {
	if api == nil || handler == nil || api.gameUI != nil {
		return ErrInvalidRequest
	}
	api.gameUI = handler
	return nil
}

type claimsContextKey struct{}

func NewAPI(repository *Repository, intents IntentHandler, config APIConfig) (*API, error) {
	if config.LimiterMaxEntries == 0 {
		config.LimiterMaxEntries = Phase0APIConfig().LimiterMaxEntries
	}
	if repository == nil || intents == nil || config.UnauthenticatedBurst < 1 || config.UnauthenticatedPerMin < 1 ||
		config.AccountBurst < 1 || config.AccountPerMin < 1 || config.MaxBodyBytes < 1024 ||
		config.TrustedProxyHops < 0 || config.TrustedProxyHops > 8 || config.LimiterMaxEntries < 1 {
		return nil, ErrInvalidRequest
	}
	unauth, _ := httpapi.NewTokenBuckets(config.UnauthenticatedBurst, config.UnauthenticatedPerMin, config.LimiterMaxEntries)
	accounts, _ := httpapi.NewTokenBuckets(config.AccountBurst, config.AccountPerMin, config.LimiterMaxEntries)
	privateRegistry, err := newPrivateAPIRegistry()
	if err != nil {
		return nil, ErrInvalidRequest
	}
	return &API{repository: repository, intents: intents, config: config,
		unauth:           unauth,
		accounts:         accounts,
		privateRegistry:  privateRegistry,
		recoveryProgress: newRecoveryBuckets(config.LimiterMaxEntries)}, nil
}

func (api *API) Router() http.Handler {
	router := chi.NewRouter()
	router.NotFound(func(response http.ResponseWriter, _ *http.Request) {
		writeError(response, http.StatusNotFound, "unknown_id", "route")
	})
	router.MethodNotAllowed(func(response http.ResponseWriter, _ *http.Request) {
		writeError(response, http.StatusMethodNotAllowed, "invalid", "method")
	})
	router.Route("/api/v1", func(v1 chi.Router) {
		v1.With(api.limitUnauthenticated).Post("/account", api.createAccount)
		v1.With(api.limitUnauthenticated).Post("/session", api.createSession)
		v1.With(api.limitUnauthenticated).Post("/session/refresh", api.refreshSession)
		v1.Group(func(authenticated chi.Router) {
			authenticated.Use(api.authenticate)
			authenticated.Use(api.limitAccount)
			authenticated.Post("/account/email", api.attachEmail)
			authenticated.Delete("/account", api.deleteAccount)
			authenticated.Delete("/session", api.deleteSession)
			authenticated.Post("/founder", api.newFounder)
			authenticated.Get("/founder", api.getFounder)
			authenticated.Post("/founder/import", api.importFounder)
			authenticated.Post("/intents", api.submitIntent)
			authenticated.Post("/guild/intents", api.submitGuildIntent)
		})
	})
	bindings := []publicapi.Binding{
		{OperationID: "cancel_soul_recovery", Handler: http.HandlerFunc(api.cancelSoulRecovery)},
		{OperationID: "create_minigame_session", Handler: http.HandlerFunc(api.createMinigameSession)},
		{OperationID: "get_current_minigame_session", Handler: http.HandlerFunc(api.getCurrentMinigameSession)},
		{OperationID: "get_game_ui_snapshot", Handler: http.HandlerFunc(api.getGameUISnapshot)},
		{OperationID: "play_minigame_command", Handler: http.HandlerFunc(api.playMinigameCommand)},
		{OperationID: "progress_soul_recovery", Handler: http.HandlerFunc(api.progressSoulRecovery)},
		{OperationID: "resolve_minigame_session", Handler: http.HandlerFunc(api.resolveMinigameSession)},
		{OperationID: "resolve_soul_recovery", Handler: http.HandlerFunc(api.resolveSoulRecovery)},
		{OperationID: "start_soul_recovery", Handler: http.HandlerFunc(api.startSoulRecovery)},
	}
	if err := api.privateRegistry.Mount(router, bindings, map[publicapi.AuthMode]publicapi.Middleware{
		publicapi.AuthAccessToken: func(next http.Handler) http.Handler { return api.authenticate(api.limitAccount(next)) },
	}); err != nil {
		panic(err)
	}
	return router
}

func (api *API) getGameUISnapshot(response http.ResponseWriter, request *http.Request) {
	if api.gameUI == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "game_ui_snapshot")
		return
	}
	state, err := api.repository.ActiveCompanyState(request.Context(), requestClaims(request).Subject)
	if err != nil {
		writeError(response, http.StatusNotFound, "unknown_id", "founder_state")
		return
	}
	encoded, err := api.gameUI.GameUISnapshot(request.Context(), state.StreamID, api.repository.clock())
	if err != nil || api.privateRegistry.ValidateResponse("get_game_ui_snapshot", http.StatusOK, encoded) != nil {
		writeError(response, http.StatusInternalServerError, "internal_invariant", "game_ui_snapshot")
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(encoded)
}

func (api *API) startSoulRecovery(response http.ResponseWriter, request *http.Request) {
	var body struct {
		ActivityID string `json:"activity_id"`
	}
	if api.recoveries == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "soul_recovery")
		return
	}
	if decodeRequest(response, request, api.config.MaxBodyBytes, &body) != nil || !apiMechanicalIDPattern.MatchString(body.ActivityID) {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	state, err := api.repository.ActiveCompanyState(request.Context(), requestClaims(request).Subject)
	if err != nil {
		writeError(response, http.StatusNotFound, "unknown_id", "company_stream")
		return
	}
	now := api.repository.clock()
	sessionID, err := newUUIDv7(now, api.repository.random)
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal_invariant", "session_id")
		return
	}
	result, err := api.recoveries.StartSoulRecovery(request.Context(), production.StartSoulRecoveryRequest{
		SessionID: sessionID, FounderID: state.FounderID, CompanyStreamID: state.StreamID, ActivityID: body.ActivityID}, now)
	api.writeSoulRecoveryResult(response, result, err, "start")
}

func (api *API) progressSoulRecovery(response http.ResponseWriter, request *http.Request) {
	var body struct {
		SessionID     string `json:"session_id"`
		ProgressToken string `json:"progress_token"`
	}
	if api.recoveries == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "soul_recovery")
		return
	}
	if decodeRequest(response, request, api.config.MaxBodyBytes, &body) != nil || !apiUUIDV7Pattern.MatchString(body.SessionID) || !apiUUIDPattern.MatchString(body.ProgressToken) {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	founder, err := api.repository.ActiveFounder(request.Context(), requestClaims(request).Subject)
	if err != nil {
		writeError(response, http.StatusNotFound, "unknown_id", "founder")
		return
	}
	ceiling, err := api.recoveries.SoulRecoveryBeatCeilingMS(request.Context(), founder.ID, body.SessionID)
	if err != nil {
		api.writeSoulRecoveryResult(response, production.HandleResult{}, err, "progress")
		return
	}
	now := api.repository.clock()
	if !api.recoveryProgress.allow(body.SessionID, now, maxInt64(1, ceiling/6)) {
		writeError(response, http.StatusTooManyRequests, "rate_limited", "recovery_progress")
		return
	}
	result, err := api.recoveries.ProgressSoulRecovery(request.Context(), production.ProgressSoulRecoveryRequest{
		SessionID: body.SessionID, FounderID: founder.ID, ProgressToken: body.ProgressToken}, now, nil)
	api.writeSoulRecoveryResult(response, result, err, "progress")
}

func (api *API) cancelSoulRecovery(response http.ResponseWriter, request *http.Request) {
	api.finishSoulRecovery(response, request, false)
}

func (api *API) resolveSoulRecovery(response http.ResponseWriter, request *http.Request) {
	api.finishSoulRecovery(response, request, true)
}

func (api *API) finishSoulRecovery(response http.ResponseWriter, request *http.Request, resolve bool) {
	var body struct {
		SessionID string `json:"session_id"`
	}
	if api.recoveries == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "soul_recovery")
		return
	}
	if decodeRequest(response, request, api.config.MaxBodyBytes, &body) != nil || !apiUUIDV7Pattern.MatchString(body.SessionID) {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	founder, err := api.repository.ActiveFounder(request.Context(), requestClaims(request).Subject)
	if err != nil {
		writeError(response, http.StatusNotFound, "unknown_id", "founder")
		return
	}
	finish := production.FinishSoulRecoveryRequest{SessionID: body.SessionID, FounderID: founder.ID}
	var result production.HandleResult
	if resolve {
		result, err = api.recoveries.ResolveSoulRecovery(request.Context(), finish, api.repository.clock(), nil)
	} else {
		result, err = api.recoveries.CancelSoulRecovery(request.Context(), finish, api.repository.clock(), nil)
	}
	if err == nil {
		api.recoveryProgress.remove(body.SessionID)
	}
	api.writeSoulRecoveryResult(response, result, err, map[bool]string{true: "resolve", false: "cancel"}[resolve])
}

func (api *API) writeSoulRecoveryResult(response http.ResponseWriter, result production.HandleResult, err error, action string) {
	if err == nil {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write(result.Receipt)
		return
	}
	switch {
	case errors.Is(err, soul.ErrRecoveryGone):
		writeError(response, http.StatusNotFound, "unknown_id", "recovery_session")
	case errors.Is(err, soul.ErrRecoveryActive):
		writeError(response, http.StatusConflict, "not_eligible", "exclusive_activity")
	case errors.Is(err, soul.ErrRecoveryBusy), errors.Is(err, soul.ErrRecoveryClaimLost):
		writeError(response, http.StatusConflict, "conflict", "recovery_session")
	case errors.Is(err, save.ErrIdempotencyConflict), errors.Is(err, soul.ErrRecoveryIdempotency):
		writeError(response, http.StatusConflict, "idempotency_conflict", "recovery_session")
	case errors.Is(err, production.ErrSoulRecoveryToken):
		writeError(response, http.StatusBadRequest, "not_eligible", "recovery_token")
	case errors.Is(err, production.ErrSoulRecoveryNotReady):
		writeError(response, http.StatusBadRequest, "not_eligible", "soul_recovery_not_ready")
	case errors.Is(err, production.ErrInvalidIntent), errors.Is(err, soul.ErrInvalidRecovery):
		writeError(response, http.StatusBadRequest, "invalid", "soul_recovery_"+action)
	default:
		writeError(response, http.StatusInternalServerError, "internal_invariant", "soul_recovery")
	}
}

func (api *API) submitGuildIntent(response http.ResponseWriter, request *http.Request) {
	if api.guilds == nil {
		writeError(response, http.StatusServiceUnavailable, "not_configured", "guild_intents")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(response, request.Body, api.config.MaxBodyBytes))
	if err != nil || len(body) == 0 || !json.Valid(body) {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	receipt, _, err := api.guilds.HandleGuild(request.Context(), requestClaims(request).Subject, body)
	if err != nil && api.guilds.IsInvalidGuildIntent(err) {
		writeError(response, http.StatusBadRequest, "invalid", "guild_intent")
		return
	}
	if err != nil {
		writeError(response, http.StatusConflict, "conflict", "guild_intent")
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(receipt)
}

func (api *API) createAccount(response http.ResponseWriter, request *http.Request) {
	if err := decodeEmptyRequest(response, request, api.config.MaxBodyBytes); err != nil {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	created, err := api.repository.CreateAccount(request.Context())
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal_invariant", "account_create")
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusCreated, created)
}

func (api *API) createSession(response http.ResponseWriter, request *http.Request) {
	var body struct {
		AccountID    string `json:"account_id"`
		RecoveryCode string `json:"recovery_code"`
	}
	if err := decodeRequest(response, request, api.config.MaxBodyBytes, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	pair, err := api.repository.CreateSession(request.Context(), body.AccountID, body.RecoveryCode)
	if err != nil {
		writeError(response, http.StatusUnauthorized, "unauthorized", "credential")
		return
	}
	writeJSON(response, http.StatusOK, pair)
}

func (api *API) refreshSession(response http.ResponseWriter, request *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeRequest(response, request, api.config.MaxBodyBytes, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	pair, err := api.repository.RefreshSession(request.Context(), body.RefreshToken)
	if errors.Is(err, ErrRefreshReuse) {
		writeError(response, http.StatusUnauthorized, "refresh_reused", "session_family_revoked")
		return
	}
	if err != nil {
		writeError(response, http.StatusUnauthorized, "unauthorized", "refresh_token")
		return
	}
	writeJSON(response, http.StatusOK, pair)
}

func (api *API) attachEmail(response http.ResponseWriter, _ *http.Request) {
	writeError(response, http.StatusNotImplemented, "not_configured", "email_provider")
}

func (api *API) deleteSession(response http.ResponseWriter, request *http.Request) {
	if err := decodeEmptyRequest(response, request, api.config.MaxBodyBytes); err != nil {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	if err := api.repository.RevokeSession(request.Context(), requestClaims(request)); err != nil {
		writeError(response, http.StatusUnauthorized, "unauthorized", "session")
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (api *API) deleteAccount(response http.ResponseWriter, request *http.Request) {
	if err := decodeEmptyRequest(response, request, api.config.MaxBodyBytes); err != nil {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	if err := api.repository.DeleteAccount(request.Context(), requestClaims(request).Subject); err != nil {
		writeError(response, http.StatusNotFound, "unknown_id", "account")
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (api *API) newFounder(response http.ResponseWriter, request *http.Request) {
	if err := decodeEmptyRequest(response, request, api.config.MaxBodyBytes); err != nil {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	founder, err := api.repository.NewFounder(request.Context(), requestClaims(request).Subject)
	if errors.Is(err, ErrAccountNotFound) {
		writeError(response, http.StatusNotFound, "unknown_id", "account")
		return
	}
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal_invariant", "founder_create")
		return
	}
	writeJSON(response, http.StatusCreated, founder)
}

func (api *API) getFounder(response http.ResponseWriter, request *http.Request) {
	founder, err := api.repository.ActiveFounder(request.Context(), requestClaims(request).Subject)
	if err != nil {
		writeError(response, http.StatusNotFound, "unknown_id", "founder")
		return
	}
	writeJSON(response, http.StatusOK, struct {
		ID        string         `json:"id"`
		CreatedAt time.Time      `json:"created_at"`
		Display   map[string]any `json:"display"`
	}{ID: founder.ID, CreatedAt: founder.CreatedAt, Display: map[string]any{}})
}

func (api *API) importFounder(response http.ResponseWriter, request *http.Request) {
	var body struct {
		Version       int             `json:"version"`
		ConstantsHash string          `json:"constants_hash"`
		State         json.RawMessage `json:"state"`
	}
	if err := decodeRequest(response, request, api.config.MaxBodyBytes, &body); err != nil {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	founder, err := api.repository.ImportFounder(request.Context(), requestClaims(request).Subject, body.ConstantsHash, body.Version, body.State)
	if errors.Is(err, ErrImportUnavailable) {
		writeError(response, http.StatusConflict, "already_unlocked", "founder_import")
		return
	}
	if err != nil {
		writeError(response, http.StatusBadRequest, "invalid", "save")
		return
	}
	writeJSON(response, http.StatusOK, founder)
}

func (api *API) submitIntent(response http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(response, request.Body, api.config.MaxBodyBytes))
	if err != nil || len(body) == 0 || !json.Valid(body) {
		writeError(response, http.StatusBadRequest, "invalid", "body")
		return
	}
	state, err := api.repository.ActiveCompanyState(request.Context(), requestClaims(request).Subject)
	if err != nil {
		writeError(response, http.StatusNotFound, "unknown_id", "company_stream")
		return
	}
	result, err := api.intents.Handle(request.Context(), state.StreamID, production.ModeOnline, api.repository.clock(), body)
	if errors.Is(err, production.ErrInvalidIntent) {
		writeError(response, http.StatusBadRequest, "invalid", "intent")
		return
	}
	if errors.Is(err, production.ErrInvalidEngineState) {
		writeError(response, http.StatusInternalServerError, "internal_invariant", "intent")
		return
	}
	if err != nil {
		writeError(response, http.StatusConflict, "conflict", "intent")
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(result.Receipt)
}

func (api *API) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") || strings.Contains(strings.TrimPrefix(authorization, "Bearer "), " ") {
			api.writeAuthenticationFailure(response, request)
			return
		}
		claims, err := api.repository.Authenticate(request.Context(), strings.TrimPrefix(authorization, "Bearer "))
		if err != nil {
			api.writeAuthenticationFailure(response, request)
			return
		}
		next.ServeHTTP(response, request.WithContext(context.WithValue(request.Context(), claimsContextKey{}, claims)))
	})
}

func (api *API) limitUnauthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !api.unauth.Allow(api.clientIP(request), api.repository.clock()) {
			writeError(response, http.StatusTooManyRequests, "rate_limited", "ip")
			return
		}
		next.ServeHTTP(response, request)
	})
}

func (api *API) writeAuthenticationFailure(response http.ResponseWriter, request *http.Request) {
	if !api.unauth.Allow(api.clientIP(request), api.repository.clock()) {
		writeError(response, http.StatusTooManyRequests, "rate_limited", "ip")
		return
	}
	writeError(response, http.StatusUnauthorized, "unauthorized", "access_token")
}

func (api *API) clientIP(request *http.Request) string {
	return httpapi.ClientIP(request, api.config.TrustedProxyHops)
}

func (api *API) limitAccount(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !api.accounts.Allow(requestClaims(request).Subject, api.repository.clock()) {
			writeError(response, http.StatusTooManyRequests, "rate_limited", "account")
			return
		}
		next.ServeHTTP(response, request)
	})
}

func requestClaims(request *http.Request) Claims {
	claims, _ := request.Context().Value(claimsContextKey{}).(Claims)
	return claims
}

func decodeRequest(response http.ResponseWriter, request *http.Request, limit int64, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, limit))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return ErrInvalidRequest
		}
		return err
	}
	return nil
}

func decodeEmptyRequest(response http.ResponseWriter, request *http.Request, limit int64) error {
	body, err := io.ReadAll(http.MaxBytesReader(response, request.Body, limit))
	if err != nil {
		return err
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed != "" && trimmed != "{}" {
		return ErrInvalidRequest
	}
	return nil
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func writeError(response http.ResponseWriter, status int, category, detail string) {
	writeJSON(response, status, struct {
		Category string `json:"category"`
		Detail   string `json:"detail"`
	}{Category: category, Detail: detail})
}

type recoveryBucket struct {
	tokens       int64
	lastRefillMS int64
	lastSeen     time.Time
	element      *list.Element
}

type recoveryBuckets struct {
	mu      sync.Mutex
	values  map[string]*recoveryBucket
	recency *list.List
	max     int
}

func newRecoveryBuckets(maxEntries int) *recoveryBuckets {
	return &recoveryBuckets{values: map[string]*recoveryBucket{}, recency: list.New(), max: maxEntries}
}

func (buckets *recoveryBuckets) allow(key string, now time.Time, refillMS int64) bool {
	if buckets == nil || key == "" || now.IsZero() || refillMS < 1 {
		return false
	}
	buckets.mu.Lock()
	defer buckets.mu.Unlock()
	buckets.evict(now)
	nowMS := now.UnixMilli()
	current := buckets.values[key]
	if current == nil {
		if len(buckets.values) >= buckets.max {
			buckets.removeOldest()
		}
		current = &recoveryBucket{tokens: 6, lastRefillMS: nowMS, lastSeen: now}
		current.element = buckets.recency.PushFront(key)
		buckets.values[key] = current
	} else {
		if nowMS > current.lastRefillMS {
			refills := (nowMS - current.lastRefillMS) / refillMS
			if refills > 0 {
				current.tokens = minInt64(6, current.tokens+refills)
				current.lastRefillMS += refills * refillMS
			}
		}
		if !now.Before(current.lastSeen) {
			current.lastSeen = now
			buckets.recency.MoveToFront(current.element)
		}
	}
	if current.tokens == 0 {
		return false
	}
	current.tokens--
	return true
}

func (buckets *recoveryBuckets) remove(key string) {
	if buckets == nil {
		return
	}
	buckets.mu.Lock()
	defer buckets.mu.Unlock()
	if current := buckets.values[key]; current != nil {
		buckets.recency.Remove(current.element)
		delete(buckets.values, key)
	}
}

func (buckets *recoveryBuckets) evict(now time.Time) {
	for element := buckets.recency.Back(); element != nil; element = buckets.recency.Back() {
		key := element.Value.(string)
		current := buckets.values[key]
		if now.Before(current.lastSeen) || now.Sub(current.lastSeen) < 15*time.Minute {
			return
		}
		buckets.recency.Remove(element)
		delete(buckets.values, key)
	}
}

func (buckets *recoveryBuckets) removeOldest() {
	element := buckets.recency.Back()
	if element == nil {
		return
	}
	delete(buckets.values, element.Value.(string))
	buckets.recency.Remove(element)
}

func minInt64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
