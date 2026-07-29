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
	"strings"
	"time"
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
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}
