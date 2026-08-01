package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"time"
)

const WireVersion = 2

const (
	CloseQueueOverflow = 4000
	CloseAuthExpired   = 4001
	CloseReplaced      = 4002
	CloseServerDrain   = 4003
	CloseInvalidFrame  = 4004
)

var hashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var eventKindPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)

type Envelope struct {
	Version       int             `json:"v"`
	Channel       string          `json:"ch"`
	Kind          string          `json:"kind"`
	Revision      int64           `json:"rev"`
	ConstantsHash string          `json:"constants_hash"`
	Timestamp     time.Time       `json:"ts"`
	Payload       json.RawMessage `json:"payload"`
}

func Encode(envelope Envelope, messageCap int) ([]byte, error) {
	if envelope.Version != WireVersion || envelope.Channel == "" || !knownKind(envelope.Kind) || envelope.Revision < 0 ||
		!hashPattern.MatchString(envelope.ConstantsHash) || envelope.Timestamp.IsZero() || len(envelope.Payload) == 0 || messageCap < 1 ||
		!channelAllowsKind(envelope.Channel, envelope.Kind) || validatePayload(envelope) != nil {
		return nil, ErrInvalidPolicy
	}
	data, err := json.Marshal(envelope)
	if err != nil || len(data) > messageCap {
		return nil, ErrInvalidPolicy
	}
	return data, nil
}

func knownKind(kind string) bool {
	return kind == "receipt" || kind == "snapshot" || kind == "event" || kind == "presence" || kind == "system"
}

func channelAllowsKind(channel, kind string) bool {
	switch {
	case isPlayerChannel(channel):
		id := strings.TrimPrefix(channel, "player:")
		return id != "" && !strings.Contains(id, ":")
	case channel == "world":
		return kind == "snapshot" || kind == "presence" || kind == "system"
	case channel == "feed":
		return kind == "event" || kind == "presence" || kind == "system"
	case strings.HasPrefix(channel, "guild:"), strings.HasPrefix(channel, "cohort:"), strings.HasPrefix(channel, "match:"):
		_, id, ok := strings.Cut(channel, ":")
		return ok && id != "" && !strings.Contains(id, ":") && kind != "receipt"
	default:
		return false
	}
}

func validatePayload(envelope Envelope) error {
	switch envelope.Kind {
	case "receipt":
		// Production C1 owns the exact receipt schema. Transport preserves it
		// byte-for-byte and only asserts that the owned payload is an object.
		if !isJSONObject(envelope.Payload) {
			return ErrInvalidPolicy
		}
	case "snapshot":
		var payload struct {
			Scope string          `json:"scope"`
			Rev   int64           `json:"rev"`
			State json.RawMessage `json:"state"`
		}
		if decodeExactPayload(envelope.Payload, &payload) != nil || payload.Rev < 0 || payload.Rev != envelope.Revision ||
			!isJSONObject(payload.State) || !scopeMatchesChannel(payload.Scope, envelope.Channel) {
			return ErrInvalidPolicy
		}
	case "event":
		var payload struct {
			EventID      string          `json:"event_id"`
			Kind         string          `json:"kind"`
			Scope        string          `json:"scope"`
			Rev          int64           `json:"rev"`
			CursorEffect string          `json:"cursor_effect"`
			Payload      json.RawMessage `json:"payload"`
		}
		if decodeExactPayload(envelope.Payload, &payload) != nil || payload.EventID == "" || !eventKindPattern.MatchString(payload.Kind) ||
			(payload.Scope != "company" && payload.Scope != "founder") || payload.Rev < 1 ||
			payload.Rev != envelope.Revision || !isJSONObject(payload.Payload) ||
			(payload.Kind == "compensation") != (payload.CursorEffect == "historical") ||
			(payload.CursorEffect != "advance" && payload.CursorEffect != "historical") {
			return ErrInvalidPolicy
		}
	case "presence":
		var payload struct {
			Joined []string `json:"joined"`
			Left   []string `json:"left"`
			Count  int64    `json:"count"`
		}
		if decodeExactPayload(envelope.Payload, &payload) != nil || payload.Joined == nil || payload.Left == nil || payload.Count < 0 ||
			!nonemptyStrings(payload.Joined) || !nonemptyStrings(payload.Left) {
			return ErrInvalidPolicy
		}
	case "system":
		var payload struct {
			Code          string `json:"code"`
			ResumeAfterMS *int64 `json:"resume_after_ms,omitempty"`
		}
		if decodeExactPayload(envelope.Payload, &payload) != nil {
			return ErrInvalidPolicy
		}
		switch payload.Code {
		case "server_restarting":
			if payload.ResumeAfterMS == nil || *payload.ResumeAfterMS < 0 {
				return ErrInvalidPolicy
			}
		case "history_expired", "resync_required":
			if payload.ResumeAfterMS != nil {
				return ErrInvalidPolicy
			}
		default:
			return ErrInvalidPolicy
		}
	default:
		return ErrInvalidPolicy
	}
	return nil
}

func scopeMatchesChannel(scope, channel string) bool {
	switch scope {
	case "company":
		return isPlayerChannel(channel)
	case "world":
		return channel == "world"
	case "guild":
		return strings.HasPrefix(channel, "guild:")
	case "cohort":
		return strings.HasPrefix(channel, "cohort:")
	default:
		return false
	}
}

func decodeExactPayload(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrInvalidPolicy
	}
	return nil
}

func isJSONObject(data []byte) bool {
	var object map[string]json.RawMessage
	return decodeExactPayload(data, &object) == nil && object != nil
}

func nonemptyStrings(values []string) bool {
	for _, value := range values {
		if value == "" {
			return false
		}
	}
	return true
}
