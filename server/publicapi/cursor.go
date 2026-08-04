package publicapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

const (
	cursorMACBytes       = sha256.Size
	maximumCursorEncoded = 2048
)

var ErrInvalidCursor = errors.New("invalid cursor")

type CursorCodec struct {
	current  []byte
	previous []byte
	registry *Registry
}

type cursorFields struct {
	FilterSHA256 string          `json:"filter_sha256"`
	Key          json.RawMessage `json:"key"`
	Operation    string          `json:"op"`
	Version      int             `json:"v"`
}

var filterHashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func NewCursorCodec(current, previous []byte, registry *Registry) (*CursorCodec, error) {
	if len(current) < 32 || len(previous) != 0 && len(previous) < 32 || len(previous) != 0 && bytes.Equal(current, previous) || registry == nil {
		return nil, ErrInvalidCursor
	}
	return &CursorCodec{current: append([]byte(nil), current...), previous: append([]byte(nil), previous...), registry: registry}, nil
}

func (codec *CursorCodec) Encode(operation, filterSHA256 string, key any) (string, error) {
	keySchema, known := codec.cursorKeySchema(operation)
	if codec == nil || !known || !filterHashPattern.MatchString(filterSHA256) || key == nil {
		return "", ErrInvalidCursor
	}
	keyBytes, err := json.Marshal(key)
	if err != nil || !canonicalJSON(keyBytes) || ValidateJSON(keySchema, keyBytes, codec.registry.schemas) != nil {
		return "", ErrInvalidCursor
	}
	payload, err := json.Marshal(cursorFields{Version: 1, Operation: operation, FilterSHA256: filterSHA256, Key: keyBytes})
	if err != nil || len(payload)+cursorMACBytes > base64.RawURLEncoding.DecodedLen(maximumCursorEncoded) {
		return "", ErrInvalidCursor
	}
	mac := hmac.New(sha256.New, codec.current)
	_, _ = mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(append(payload, mac.Sum(nil)...)), nil
}

func (codec *CursorCodec) Decode(token, operation, filterSHA256 string, keyTarget any) error {
	keySchema, known := codec.cursorKeySchema(operation)
	if codec == nil || !known || len(token) == 0 || len(token) > maximumCursorEncoded || !filterHashPattern.MatchString(filterSHA256) || keyTarget == nil {
		return ErrInvalidCursor
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) <= cursorMACBytes {
		return ErrInvalidCursor
	}
	payload, signature := raw[:len(raw)-cursorMACBytes], raw[len(raw)-cursorMACBytes:]
	if !codec.validMAC(payload, signature) {
		return ErrInvalidCursor
	}
	var fields cursorFields
	if !canonicalJSON(payload) {
		return ErrInvalidCursor
	}
	if err := decodeCursorStrict(payload, &fields); err != nil || fields.Version != 1 || fields.Operation != operation || fields.FilterSHA256 != filterSHA256 || !canonicalJSON(fields.Key) || ValidateJSON(keySchema, fields.Key, codec.registry.schemas) != nil {
		return ErrInvalidCursor
	}
	if err := decodeCursorStrict(fields.Key, keyTarget); err != nil {
		return ErrInvalidCursor
	}
	reencoded, err := json.Marshal(keyTarget)
	if err != nil || !bytes.Equal(reencoded, fields.Key) {
		return ErrInvalidCursor
	}
	return nil
}

func (codec *CursorCodec) cursorKeySchema(operation string) (string, bool) {
	if codec == nil || !operationIDPattern.MatchString(operation) {
		return "", false
	}
	return codec.registry.cursorSchema(operation)
}

func (codec *CursorCodec) validMAC(payload, signature []byte) bool {
	for _, key := range [][]byte{codec.current, codec.previous} {
		if len(key) == 0 {
			continue
		}
		mac := hmac.New(sha256.New, key)
		_, _ = mac.Write(payload)
		if hmac.Equal(signature, mac.Sum(nil)) {
			return true
		}
	}
	return false
}

func canonicalJSON(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil {
		return false
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return false
	}
	encoded, err := json.Marshal(value)
	return err == nil && bytes.Equal(encoded, data)
}

func decodeCursorStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ErrInvalidCursor
	}
	return nil
}
