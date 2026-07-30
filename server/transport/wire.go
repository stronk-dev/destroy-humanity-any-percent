package transport

import (
	"encoding/json"
	"regexp"
	"time"
)

const WireVersion = 1

const (
	CloseQueueOverflow = 4000
	CloseAuthExpired   = 4001
	CloseReplaced      = 4002
	CloseServerDrain   = 4003
)

var hashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

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
		!hashPattern.MatchString(envelope.ConstantsHash) || envelope.Timestamp.IsZero() || len(envelope.Payload) == 0 || messageCap < 1 {
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
