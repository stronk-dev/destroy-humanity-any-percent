package account

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/httpapi"

	"golang.org/x/crypto/argon2"
)

func testBootstrapReceiptKeys() BootstrapReceiptKeys {
	return BootstrapReceiptKeys{CurrentID: "bootstrap-test", Current: bytes.Repeat([]byte{0x7b}, 32)}
}

func TestRecoveryCredentialRoundTripAndEncodedParameters(t *testing.T) {
	random := bytes.NewReader(bytes.Repeat([]byte{0x5a}, 64))
	code, err := newRecoveryCode(random)
	if err != nil || !validRecoveryCode(code) {
		t.Fatalf("code=%q err=%v", code, err)
	}
	encoded, err := hashRecoveryCode(code, random)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(encoded, "$m=19456,t=2,p=1$") || !verifyRecoveryCode(encoded, "  "+strings.ToUpper(code)+"  ") || verifyRecoveryCode(encoded, strings.Repeat("a", 26)) {
		t.Fatalf("credential verification/parameters failed: %s", encoded)
	}
	stronger := recoveryHashForTest(code, argonMemoryKiB+1024, argonIterations+1, argonParallelism)
	if valid, upgrade := verifyRecoveryCodeForUpgrade(stronger, code); !valid || !upgrade {
		t.Fatalf("stored stronger credential valid=%v upgrade=%v", valid, upgrade)
	}
	belowFloor := recoveryHashForTest(code, argonMemoryKiB-1, argonIterations, argonParallelism)
	if valid, _ := verifyRecoveryCodeForUpgrade(belowFloor, code); valid {
		t.Fatal("credential below the Argon2 memory floor verified")
	}
	aboveCeiling := strings.Replace(dummyRecoveryHash, "m=19456", fmt.Sprintf("m=%d", argonMemoryMaxKiB+1), 1)
	if valid, _ := verifyRecoveryCodeForUpgrade(aboveCeiling, code); valid {
		t.Fatal("credential above the Argon2 work ceiling verified")
	}
	if verifyRecoveryCode(dummyRecoveryHash, code) {
		t.Fatal("dummy recovery hash authenticated a real code")
	}
}

func TestJWTExactClaimsExpirySignatureAndPreviousKey(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	keys := SigningKeys{CurrentID: "current", Current: bytes.Repeat([]byte{1}, 32), PreviousID: "previous", Previous: bytes.Repeat([]byte{2}, 32)}
	claims := Claims{Subject: "01985555-1111-7111-8111-111111111111", FounderID: "01985555-2222-7222-8222-222222222222", IssuedAt: now.Unix(), ExpiresAt: now.Add(accessTTL).Unix(), TokenID: "01985555-3333-7333-8333-333333333333"}
	token, err := signAccessToken(keys, claims)
	if err != nil {
		t.Fatal(err)
	}
	verified, err := verifyAccessToken(keys, token, now)
	if err != nil || verified != claims {
		t.Fatalf("claims=%+v err=%v", verified, err)
	}
	if _, err := verifyAccessToken(keys, token, now.Add(accessTTL)); err == nil {
		t.Fatal("expired token verified")
	}

	parts := strings.Split(token, ".")
	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var exact map[string]json.RawMessage
	if json.Unmarshal(payload, &exact) != nil || len(exact) != 5 {
		t.Fatalf("JWT claims=%s", payload)
	}
	extraPayload := []byte(`{"sub":"01985555-1111-7111-8111-111111111111","fid":"01985555-2222-7222-8222-222222222222","exp":1785327300,"iat":1785326400,"jti":"01985555-3333-7333-8333-333333333333","role":"admin"}`)
	extra := signRawJWT(t, keys.Current, parts[0], extraPayload)
	if _, err := verifyAccessToken(keys, extra, now); err == nil {
		t.Fatal("token with an extra claim verified")
	}

	previousKeys := SigningKeys{CurrentID: "previous", Current: keys.Previous}
	previousToken, err := signAccessToken(previousKeys, claims)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifyAccessToken(keys, previousToken, now); err != nil {
		t.Fatalf("previous key token failed: %v", err)
	}
}

func TestUUIDv7AndTokenBucketClockRegression(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 123_000_000, time.UTC)
	id, err := newUUIDv7(now, bytes.NewReader(bytes.Repeat([]byte{0xaa}, 16)))
	if err != nil || len(id) != 36 || id[14] != '7' || !strings.Contains("89ab", string(id[19])) {
		t.Fatalf("uuid=%q err=%v", id, err)
	}
	buckets, err := httpapi.NewTokenBuckets(1, 60, 2)
	if err != nil || !buckets.Allow("client", now) || buckets.Allow("client", now.Add(-time.Hour)) || !buckets.Allow("client", now.Add(time.Second)) {
		t.Fatal("token bucket did not fail closed on clock regression/refill")
	}
	if !buckets.Allow("second", now.Add(2*time.Second)) || !buckets.Allow("third", now.Add(2*time.Second)) || buckets.Len() != 2 {
		t.Fatalf("bounded LRU buckets=%d", buckets.Len())
	}
	if buckets.Contains("client") {
		t.Fatal("least-recently-used bucket was not evicted")
	}
}

func TestTrustedProxyAddressAndFailedAuthenticationLimit(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/founder", nil)
	request.RemoteAddr = "10.0.0.5:443"
	request.Header.Set("X-Forwarded-For", "203.0.113.8, 10.0.0.4")
	api := &API{config: APIConfig{TrustedProxyHops: 2}}
	if got := api.clientIP(request); got != "203.0.113.8" {
		t.Fatalf("trusted proxy client=%q", got)
	}
	api.config.TrustedProxyHops = 0
	if got := api.clientIP(request); got != "10.0.0.5" {
		t.Fatalf("untrusted forwarded chain changed client=%q", got)
	}

	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	unauth, _ := httpapi.NewTokenBuckets(1, 1, 10)
	api = &API{repository: &Repository{clock: func() time.Time { return now }}, config: APIConfig{}, unauth: unauth}
	handler := api.authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("unauthenticated request reached handler")
	}))
	for index, want := range []int{http.StatusUnauthorized, http.StatusTooManyRequests} {
		response := httptest.NewRecorder()
		attempt := httptest.NewRequest(http.MethodGet, "/api/v1/founder", nil)
		attempt.RemoteAddr = "192.0.2.10:1234"
		handler.ServeHTTP(response, attempt)
		if response.Code != want {
			t.Fatalf("attempt %d status=%d want=%d body=%s", index, response.Code, want, response.Body.String())
		}
	}
}

func recoveryHashForTest(code string, memory, iterations uint32, parallelism uint8) string {
	salt := bytes.Repeat([]byte{0x33}, argonSaltBytes)
	hash := argon2.IDKey([]byte(code), salt, iterations, memory, parallelism, argonKeyBytes)
	encoding := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, iterations, parallelism,
		encoding.EncodeToString(salt), encoding.EncodeToString(hash))
}

func signRawJWT(t *testing.T, key []byte, encodedHeader string, payload []byte) string {
	t.Helper()
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	unsigned := encodedHeader + "." + encodedPayload
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
