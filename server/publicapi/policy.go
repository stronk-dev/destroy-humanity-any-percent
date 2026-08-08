package publicapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"cloud-clicker/server/httpapi"
)

var ErrInvalidPolicy = errors.New("invalid public API policy")

const requestIDPatternLiteral = `^[A-Za-z0-9-]{1,64}$`

type CachePolicy struct {
	Boards         int `json:"boards"`
	CatalogsEpochs int `json:"catalogs_epochs"`
	Registry       int `json:"registry"`
	Verification   int `json:"verification"`
}

type CursorKeyIDs struct {
	Current  string `json:"current"`
	Previous string `json:"previous"`
}

type LimiterPolicy struct {
	Burst           int `json:"burst"`
	MaxIPEntries    int `json:"max_ip_entries"`
	RefillPerMinute int `json:"refill_per_minute"`
}

type RequestIDPolicy struct {
	MaxBytes int    `json:"max_bytes"`
	Pattern  string `json:"pattern"`
}

type Policy struct {
	CacheMaxAgeSeconds CachePolicy     `json:"cache_max_age_seconds"`
	CursorKeyIDs       CursorKeyIDs    `json:"cursor_key_ids"`
	PublicLimiter      LimiterPolicy   `json:"public_limiter"`
	RequestID          RequestIDPolicy `json:"request_id"`
	SchemaVersion      int             `json:"schema_version"`
	TrustedProxyHops   int             `json:"trusted_proxy_hops"`
}

func LoadPolicy(data []byte) (Policy, error) {
	if len(data) == 0 || len(data) > 64<<10 || rejectDuplicateJSONKeys(data) != nil {
		return Policy{}, ErrInvalidPolicy
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var policy Policy
	if decoder.Decode(&policy) != nil {
		return Policy{}, ErrInvalidPolicy
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF || !validPolicy(policy) {
		return Policy{}, ErrInvalidPolicy
	}
	return policy, nil
}

func validPolicy(policy Policy) bool {
	return policy.SchemaVersion == 1 &&
		policy.CacheMaxAgeSeconds == (CachePolicy{Boards: 60, CatalogsEpochs: 3600, Registry: 300, Verification: 31_536_000}) &&
		policy.CursorKeyIDs == (CursorKeyIDs{Current: "k1", Previous: "k0"}) &&
		policy.PublicLimiter.Burst > 0 && policy.PublicLimiter.Burst <= 1_000_000 &&
		policy.PublicLimiter.RefillPerMinute > 0 && policy.PublicLimiter.RefillPerMinute <= 1_000_000 &&
		policy.PublicLimiter.MaxIPEntries > 0 && policy.PublicLimiter.MaxIPEntries <= 1_000_000 &&
		policy.RequestID == (RequestIDPolicy{MaxBytes: 64, Pattern: requestIDPatternLiteral}) &&
		policy.TrustedProxyHops >= 0 && policy.TrustedProxyHops <= 8
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := walkJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return ErrInvalidPolicy
	}
	return nil
}

func walkJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			key, ok := keyToken.(string)
			if err != nil || !ok || seen[key] {
				return ErrInvalidPolicy
			}
			seen[key] = true
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return ErrInvalidPolicy
		}
	case '[':
		for decoder.More() {
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return ErrInvalidPolicy
		}
	default:
		return ErrInvalidPolicy
	}
	return nil
}

type SecretResolver func(string) ([]byte, bool)

func ResolveCursorCodec(policy Policy, registry *Registry, resolve SecretResolver) (*CursorCodec, error) {
	if !validPolicy(policy) || registry == nil || resolve == nil {
		return nil, ErrInvalidPolicy
	}
	current, currentOK := resolve(policy.CursorKeyIDs.Current)
	previous, previousOK := resolve(policy.CursorKeyIDs.Previous)
	if !currentOK || !previousOK || len(current) < 32 || len(previous) < 32 {
		return nil, ErrInvalidPolicy
	}
	if bytes.Equal(current, previous) {
		previous = nil
	}
	codec, err := NewCursorCodec(current, previous, registry)
	if err != nil {
		return nil, ErrInvalidPolicy
	}
	return codec, nil
}

type CacheClass string

const (
	CacheBoards       CacheClass = "boards"
	CacheCatalogEpoch CacheClass = "catalogs_epochs"
	CacheRegistry     CacheClass = "registry"
	CacheVerification CacheClass = "verification"
)

type Runtime struct {
	policy     Policy
	limiter    *httpapi.TokenBuckets
	requestIDs *httpapi.RequestIDs
	clock      func() time.Time
}

func NewRuntime(policy Policy, clock func() time.Time, requestIDs *httpapi.RequestIDs) (*Runtime, error) {
	if !validPolicy(policy) || clock == nil || requestIDs == nil {
		return nil, ErrInvalidPolicy
	}
	limiter, err := httpapi.NewTokenBuckets(policy.PublicLimiter.Burst, policy.PublicLimiter.RefillPerMinute, policy.PublicLimiter.MaxIPEntries)
	if err != nil {
		return nil, ErrInvalidPolicy
	}
	return &Runtime{policy: policy, limiter: limiter, requestIDs: requestIDs, clock: clock}, nil
}

type requestIDContextKey struct{}

func (runtime *Runtime) WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if runtime == nil || next == nil || request == nil {
			http.Error(response, "internal invariant", http.StatusInternalServerError)
			return
		}
		requestID, err := runtime.requestIDs.Resolve(request.Header.Get("X-Request-ID"), runtime.clock())
		if err != nil {
			http.Error(response, "internal invariant", http.StatusInternalServerError)
			return
		}
		response.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(response, request.WithContext(context.WithValue(request.Context(), requestIDContextKey{}, requestID)))
	})
}

func RequestID(request *http.Request) string {
	if request == nil {
		return ""
	}
	value, _ := request.Context().Value(requestIDContextKey{}).(string)
	return value
}

func (runtime *Runtime) WriteCached(response http.ResponseWriter, request *http.Request, class CacheClass, contentType string, body []byte) error {
	if runtime == nil || response == nil || request == nil || RequestID(request) == "" || (contentType != ContentJSON && contentType != ContentGzip) {
		return ErrInvalidPolicy
	}
	cacheControl, ok := runtime.cacheControl(class)
	if !ok {
		return ErrInvalidPolicy
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(body))
	etag := `"` + digest + `"`
	response.Header().Set("Cache-Control", cacheControl)
	response.Header().Set("ETag", etag)
	if request.Header.Get("If-None-Match") == etag {
		response.WriteHeader(http.StatusNotModified)
		return nil
	}
	clientIP := httpapi.ClientIP(request, runtime.policy.TrustedProxyHops)
	if !runtime.limiter.Allow(clientIP, runtime.clock()) {
		response.Header().Del("Cache-Control")
		response.Header().Del("ETag")
		response.Header().Set("Content-Type", ContentJSON)
		response.WriteHeader(http.StatusTooManyRequests)
		_, _ = response.Write([]byte(`{"category":"rate_limited","detail":"ip"}` + "\n"))
		return nil
	}
	response.Header().Set("Content-Type", contentType)
	response.Header().Set("Content-Length", strconv.Itoa(len(body)))
	response.WriteHeader(http.StatusOK)
	_, err := response.Write(body)
	return err
}

func (runtime *Runtime) cacheControl(class CacheClass) (string, bool) {
	seconds := 0
	immutable := false
	switch class {
	case CacheBoards:
		seconds = runtime.policy.CacheMaxAgeSeconds.Boards
	case CacheCatalogEpoch:
		seconds = runtime.policy.CacheMaxAgeSeconds.CatalogsEpochs
	case CacheRegistry:
		seconds = runtime.policy.CacheMaxAgeSeconds.Registry
	case CacheVerification:
		seconds, immutable = runtime.policy.CacheMaxAgeSeconds.Verification, true
	default:
		return "", false
	}
	value := "public,max-age=" + strconv.Itoa(seconds)
	if immutable {
		value += ",immutable"
	}
	return value, true
}
