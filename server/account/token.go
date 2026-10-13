package account

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidToken = errors.New("invalid access token")

type Claims struct {
	Subject   string `json:"sub"`
	FounderID string `json:"fid"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	TokenID   string `json:"jti"`
}

type SigningKeys struct {
	CurrentID  string
	Current    []byte
	PreviousID string
	Previous   []byte
}

func (keys SigningKeys) validate() error {
	if keys.CurrentID == "" || len(keys.Current) < 32 || keys.PreviousID != "" && len(keys.Previous) < 32 ||
		keys.PreviousID != "" && keys.PreviousID == keys.CurrentID {
		return ErrInvalidToken
	}
	return nil
}

func signAccessToken(keys SigningKeys, claims Claims) (string, error) {
	if err := keys.validate(); err != nil {
		return "", err
	}
	header, _ := json.Marshal(struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
		KeyID     string `json:"kid"`
	}{Algorithm: "HS256", Type: "JWT", KeyID: keys.CurrentID})
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoding := base64.RawURLEncoding
	unsigned := encoding.EncodeToString(header) + "." + encoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, keys.Current)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + encoding.EncodeToString(mac.Sum(nil)), nil
}

func verifyAccessToken(keys SigningKeys, token string, now time.Time) (Claims, error) {
	if err := keys.validate(); err != nil {
		return Claims{}, err
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}
	encoding := base64.RawURLEncoding
	headerBytes, err := encoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
		KeyID     string `json:"kid"`
	}
	if err := decodeExactJSON(headerBytes, &header); err != nil || header.Algorithm != "HS256" || header.Type != "JWT" {
		return Claims{}, ErrInvalidToken
	}
	key := keys.Current
	if header.KeyID != keys.CurrentID {
		if header.KeyID == "" || header.KeyID != keys.PreviousID {
			return Claims{}, ErrInvalidToken
		}
		key = keys.Previous
	}
	signature, err := encoding.DecodeString(parts[2])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return Claims{}, ErrInvalidToken
	}
	payload, err := encoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var claims Claims
	if err := decodeExactJSON(payload, &claims); err != nil || claims.Subject == "" || claims.FounderID == "" || claims.TokenID == "" ||
		claims.IssuedAt <= 0 || claims.ExpiresAt <= claims.IssuedAt || now.UTC().Unix() >= claims.ExpiresAt {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}

func decodeExactJSON(data []byte, target any) error {
	// The private JWT DTO's JSON tags own the field set. Inspect original
	// members before struct decoding can discard duplicates or fold key case.
	shape := reflect.TypeOf(target)
	if !utf8.Valid(data) || shape == nil || shape.Kind() != reflect.Pointer || shape.Elem().Kind() != reflect.Struct {
		return ErrInvalidToken
	}
	shape = shape.Elem()
	fields := make(map[string]bool, shape.NumField())
	for index := 0; index < shape.NumField(); index++ {
		field := shape.Field(index)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if field.PkgPath != "" || name == "" || name == "-" {
			return ErrInvalidToken
		}
		if _, duplicate := fields[name]; duplicate {
			return ErrInvalidToken
		}
		fields[name] = false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if opening, err := decoder.Token(); err != nil || opening != json.Delim('{') {
		return ErrInvalidToken
	}
	for decoder.More() {
		token, err := decoder.Token()
		name, isName := token.(string)
		seen, declared := fields[name]
		if err != nil || !isName || !declared || seen {
			return ErrInvalidToken
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return ErrInvalidToken
		}
		fields[name] = true
	}
	if closing, err := decoder.Token(); err != nil || closing != json.Delim('}') {
		return ErrInvalidToken
	}
	for _, seen := range fields {
		if !seen {
			return ErrInvalidToken
		}
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return json.Unmarshal(data, target)
}
