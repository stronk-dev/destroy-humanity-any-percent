package account

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cloud-clicker/server/internal/testhttp"
)

func jwtShapeFixture(t *testing.T) (SigningKeys, Claims, []byte, []byte, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	keys := SigningKeys{CurrentID: "shape-current", Current: bytes.Repeat([]byte{1}, 32),
		PreviousID: "shape-previous", Previous: bytes.Repeat([]byte{2}, 32)}
	claims := Claims{Subject: "01985555-1111-7111-8111-111111111111", FounderID: "01985555-2222-7222-8222-222222222222",
		IssuedAt: now.Unix(), ExpiresAt: now.Add(accessTTL).Unix(), TokenID: "01985555-3333-7333-8333-333333333333"}
	token, err := signAccessToken(keys, claims)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	return keys, claims, header, payload, now
}

// Keep literal duplicate/escaped spelling in the signed input; a map round trip
// would erase the very condition the decoder must reject.
func jwtMalformedObjects(t *testing.T, original []byte) map[string][]byte {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(original, &fields); err != nil {
		t.Fatal(err)
	}
	cases := map[string][]byte{
		"extra":    []byte(string(original[:len(original)-1]) + `,"extra":true}`),
		"array":    []byte(`[]`),
		"trailing": []byte(string(original) + ` {}`),
	}
	for key, value := range fields {
		member := fmt.Sprintf("%q:", key)
		cases["alias/"+key] = []byte(strings.Replace(string(original), member, fmt.Sprintf("%q:", strings.ToUpper(key)), 1))
		cases["duplicate/"+key] = []byte(string(original[:len(original)-1]) + fmt.Sprintf(",%q:%s}", key, value))
		escaped := fmt.Sprintf(`\u%04x%s`, key[0], key[1:])
		cases["escaped-duplicate/"+key] = []byte(string(original[:len(original)-1]) + fmt.Sprintf(",\"%s\":%s}", escaped, value))
		cases["null/"+key] = []byte(strings.Replace(string(original), member+string(value), member+"null", 1))
		without := make(map[string]json.RawMessage, len(fields)-1)
		for other, raw := range fields {
			if other != key {
				without[other] = raw
			}
		}
		cases["missing/"+key], _ = json.Marshal(without)
	}
	return cases
}

func TestJWTObjectShapesRejectAliasesDuplicatesAndMalformedUTF8(t *testing.T) {
	keys, _, header, payload, now := jwtShapeFixture(t)
	for _, part := range []struct {
		name string
		raw  []byte
	}{{"header", header}, {"claims", payload}} {
		cases := jwtMalformedObjects(t, part.raw)
		if part.name == "claims" {
			for _, key := range []string{"sub", "fid", "jti"} {
				cases["invalid-utf8/"+key] = []byte(strings.Replace(string(payload), fmt.Sprintf("%q:\"", key), fmt.Sprintf("%q:\"", key)+string([]byte{0xff}), 1))
			}
		}
		names := make([]string, 0, len(cases))
		for name := range cases {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			t.Run(part.name+"/"+name, func(t *testing.T) {
				candidateHeader, candidatePayload := header, payload
				if part.name == "header" {
					candidateHeader = cases[name]
				} else {
					candidatePayload = cases[name]
				}
				token := signRawJWT(t, keys.Current, base64.RawURLEncoding.EncodeToString(candidateHeader), candidatePayload)
				if _, err := verifyAccessToken(keys, token, now); !errors.Is(err, ErrInvalidToken) {
					t.Fatal("signed non-exact JWT object admitted (token withheld)")
				}
			})
		}
	}
}

func TestJWTObjectShapesPreserveLegalEncodingAndRotation(t *testing.T) {
	keys, claims, header, payload, now := jwtShapeFixture(t)
	for _, previous := range []bool{false, true} {
		key, candidateHeader := keys.Current, header
		if previous {
			key = keys.Previous
			candidateHeader = []byte(strings.Replace(string(header), keys.CurrentID, keys.PreviousID, 1))
		}
		for _, format := range []string{"issued", "reordered-whitespace", "escaped-names"} {
			t.Run(fmt.Sprintf("previous=%t/%s", previous, format), func(t *testing.T) {
				h, p := candidateHeader, payload
				if format == "reordered-whitespace" {
					var object any
					_ = json.Unmarshal(h, &object)
					h, _ = json.MarshalIndent(object, "", "  ")
					_ = json.Unmarshal(p, &object)
					p, _ = json.MarshalIndent(object, "", "  ")
				} else if format == "escaped-names" {
					h = []byte(strings.Replace(string(h), `"alg"`, `"\u0061lg"`, 1))
					p = []byte(strings.Replace(string(p), `"sub"`, `"\u0073ub"`, 1))
				}
				token := signRawJWT(t, key, base64.RawURLEncoding.EncodeToString(h), p)
				got, err := verifyAccessToken(keys, token, now)
				if err != nil || got != claims {
					t.Fatal("legal signed JWT encoding/rotation changed (token withheld)")
				}
				wrongKey := bytes.Repeat([]byte{0x99}, 32)
				badMAC := signRawJWT(t, wrongKey, base64.RawURLEncoding.EncodeToString(h), p)
				if _, err := verifyAccessToken(keys, badMAC, now); !errors.Is(err, ErrInvalidToken) {
					t.Fatal("well-shaped JWT with an invalid signature verified (token withheld)")
				}
			})
		}
	}
	t.Run("valid-replacement-character-key-id", func(t *testing.T) {
		// Valid UTF-8 U+FFFD is distinct from a malformed byte repaired by JSON.
		keys.CurrentID = "shape-\ufffd-current"
		token, err := signAccessToken(keys, claims)
		if err != nil {
			t.Fatal(err)
		}
		got, err := verifyAccessToken(keys, token, now)
		if err != nil || got != claims {
			t.Fatal("valid UTF-8 key identifier was rejected (token withheld)")
		}
	})
}

// Controlled DB refusal is a tripwire, not Postgres authorization evidence.
// A canonical JWT must reach it; malformed JWTs must fail before any connection.
type jwtDatabaseTripwire struct{ calls atomic.Int64 }

func (tripwire *jwtDatabaseTripwire) Connect(context.Context) (driver.Conn, error) {
	tripwire.calls.Add(1)
	return nil, errors.New("controlled JWT database boundary")
}
func (tripwire *jwtDatabaseTripwire) Driver() driver.Driver { return tripwire }
func (tripwire *jwtDatabaseTripwire) Open(string) (driver.Conn, error) {
	return tripwire.Connect(context.Background())
}

func TestJWTObjectRefusalThroughMountedHTTPPrecedesDatabase(t *testing.T) {
	keys, claims, header, payload, now := jwtShapeFixture(t)
	for _, part := range []string{"header", "claims"} {
		for _, kind := range []string{"alias", "duplicate", "escaped-duplicate"} {
			t.Run(part+"/"+kind, func(t *testing.T) {
				tripwire := &jwtDatabaseTripwire{}
				db := sql.OpenDB(tripwire)
				defer db.Close()
				api, err := NewAPI(&Repository{keys: keys, db: db, clock: func() time.Time { return now }},
					refreshCensusUnreachableIntents{t}, Phase0APIConfig(testBootstrapReceiptKeys()))
				if err != nil {
					t.Fatal(err)
				}
				server := testhttp.New(api.Router())
				defer server.Close()
				request := func(token string) {
					req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/founder", nil)
					if err != nil {
						t.Fatal(err)
					}
					req.Header.Set("Authorization", "Bearer "+token)
					response, err := server.Client.Do(req)
					if err != nil {
						t.Fatal("mounted authentication request failed (credentials withheld)")
					}
					body, readErr := io.ReadAll(response.Body)
					_ = response.Body.Close()
					if readErr != nil || response.StatusCode != http.StatusUnauthorized || string(body) != "{\"category\":\"unauthorized\",\"detail\":\"access_token\"}\n" {
						t.Fatal("authentication refusal lost its exact status/body (credentials withheld)")
					}
				}
				h, p := header, payload
				if part == "header" {
					h = jwtMalformedObjects(t, header)[kind+"/alg"]
				} else {
					p = jwtMalformedObjects(t, payload)[kind+"/sub"]
				}
				request(signRawJWT(t, keys.Current, base64.RawURLEncoding.EncodeToString(h), p))
				if tripwire.calls.Load() != 0 {
					t.Fatal("non-exact JWT reached the database boundary")
				}
				canonical, err := signAccessToken(keys, claims)
				if err != nil {
					t.Fatal(err)
				}
				request(canonical)
				if tripwire.calls.Load() != 1 {
					t.Fatal("canonical JWT did not reach exactly one database check")
				}
			})
		}
	}
}
