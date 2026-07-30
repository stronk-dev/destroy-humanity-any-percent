package account

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"cloud-clicker/server/production"

	"github.com/go-chi/chi/v5"
)

type IntentHandler interface {
	Handle(context.Context, string, production.EvaluationMode, time.Time, []byte) (production.HandleResult, error)
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
	repository *Repository
	intents    IntentHandler
	config     APIConfig
	unauth     *tokenBuckets
	accounts   *tokenBuckets
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
	return &API{repository: repository, intents: intents, config: config,
		unauth:   newTokenBuckets(config.UnauthenticatedBurst, config.UnauthenticatedPerMin, config.LimiterMaxEntries),
		accounts: newTokenBuckets(config.AccountBurst, config.AccountPerMin, config.LimiterMaxEntries)}, nil
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
			authenticated.Get("/founder/state", api.getFounderState)
			authenticated.Post("/intents", api.submitIntent)
		})
	})
	return router
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

func (api *API) getFounderState(response http.ResponseWriter, request *http.Request) {
	state, err := api.repository.ActiveCompanyState(request.Context(), requestClaims(request).Subject)
	if err != nil {
		writeError(response, http.StatusNotFound, "unknown_id", "founder_state")
		return
	}
	writeJSON(response, http.StatusOK, state)
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
		if !api.unauth.allow(api.clientIP(request), api.repository.clock()) {
			writeError(response, http.StatusTooManyRequests, "rate_limited", "ip")
			return
		}
		next.ServeHTTP(response, request)
	})
}

func (api *API) writeAuthenticationFailure(response http.ResponseWriter, request *http.Request) {
	if !api.unauth.allow(api.clientIP(request), api.repository.clock()) {
		writeError(response, http.StatusTooManyRequests, "rate_limited", "ip")
		return
	}
	writeError(response, http.StatusUnauthorized, "unauthorized", "access_token")
}

func (api *API) clientIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		host = request.RemoteAddr
	}
	if parsed := net.ParseIP(strings.TrimSpace(host)); parsed != nil {
		host = parsed.String()
	}
	if api.config.TrustedProxyHops == 0 {
		return host
	}
	var forwarded []string
	for _, value := range request.Header.Values("X-Forwarded-For") {
		for _, entry := range strings.Split(value, ",") {
			forwarded = append(forwarded, strings.TrimSpace(entry))
		}
	}
	if len(forwarded) < api.config.TrustedProxyHops {
		return host
	}
	candidate := forwarded[len(forwarded)-api.config.TrustedProxyHops]
	if parsed := net.ParseIP(candidate); parsed != nil {
		return parsed.String()
	}
	return host
}

func (api *API) limitAccount(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !api.accounts.allow(requestClaims(request).Subject, api.repository.clock()) {
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

type bucket struct {
	tokens   float64
	last     time.Time
	lastSeen time.Time
	element  *list.Element
}

type tokenBuckets struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	recency  *list.List
	capacity float64
	perMS    float64
	idleTTL  time.Duration
	max      int
}

func newTokenBuckets(capacity, perMinute, maxEntries int) *tokenBuckets {
	refillMinutes := (capacity + perMinute - 1) / perMinute
	if refillMinutes < 1 {
		refillMinutes = 1
	}
	return &tokenBuckets{buckets: make(map[string]*bucket), recency: list.New(), capacity: float64(capacity),
		perMS: float64(perMinute) / float64(time.Minute/time.Millisecond), idleTTL: time.Duration(refillMinutes) * time.Minute, max: maxEntries}
}

func (buckets *tokenBuckets) allow(key string, now time.Time) bool {
	buckets.mu.Lock()
	defer buckets.mu.Unlock()
	buckets.evictExpired(now)
	current, ok := buckets.buckets[key]
	if !ok {
		if len(buckets.buckets) >= buckets.max {
			buckets.removeOldest()
		}
		current = &bucket{tokens: buckets.capacity, last: now, lastSeen: now}
		current.element = buckets.recency.PushFront(key)
		buckets.buckets[key] = current
	} else if !now.Before(current.lastSeen) {
		current.lastSeen = now
		buckets.recency.MoveToFront(current.element)
	}
	elapsed := now.Sub(current.last).Milliseconds()
	if elapsed > 0 {
		current.tokens += float64(elapsed) * buckets.perMS
		if current.tokens > buckets.capacity {
			current.tokens = buckets.capacity
		}
		current.last = now
	}
	if current.tokens < 1 {
		return false
	}
	current.tokens--
	return true
}

func (buckets *tokenBuckets) evictExpired(now time.Time) {
	for element := buckets.recency.Back(); element != nil; element = buckets.recency.Back() {
		key := element.Value.(string)
		current := buckets.buckets[key]
		if now.Before(current.lastSeen) || now.Sub(current.lastSeen) < buckets.idleTTL {
			return
		}
		buckets.recency.Remove(element)
		delete(buckets.buckets, key)
	}
}

func (buckets *tokenBuckets) removeOldest() {
	element := buckets.recency.Back()
	if element == nil {
		return
	}
	delete(buckets.buckets, element.Value.(string))
	buckets.recency.Remove(element)
}
