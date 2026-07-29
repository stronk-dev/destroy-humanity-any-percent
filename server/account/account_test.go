package account

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

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
	if !strings.Contains(encoded, "$m=19456,t=2,p=1$") || !verifyRecoveryCode(encoded, code) || verifyRecoveryCode(encoded, strings.Repeat("a", 26)) {
		t.Fatalf("credential verification/parameters failed: %s", encoded)
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
	buckets := newTokenBuckets(1, 60)
	if !buckets.allow("client", now) || buckets.allow("client", now.Add(-time.Hour)) || !buckets.allow("client", now.Add(time.Second)) {
		t.Fatal("token bucket did not fail closed on clock regression/refill")
	}
}

func signRawJWT(t *testing.T, key []byte, encodedHeader string, payload []byte) string {
	t.Helper()
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	unsigned := encodedHeader + "." + encodedPayload
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
