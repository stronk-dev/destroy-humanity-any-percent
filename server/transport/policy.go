package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/url"
)

var ErrInvalidPolicy = errors.New("invalid transport policy")

type Policy struct {
	SchemaVersion              int      `json:"schema_version"`
	WorldHz                    int      `json:"world_hz"`
	FeedHistorySize            int      `json:"feed_history_size"`
	PlayerHistorySize          int      `json:"player_history_size"`
	PlayerHistoryTTLMS         int64    `json:"player_history_ttl_ms"`
	PlayerQueueMessages        int      `json:"player_queue_messages"`
	PlayerQueueBytes           int      `json:"player_queue_bytes"`
	MessageBytes               int      `json:"message_bytes"`
	SubscriptionsPerConnection int      `json:"subscriptions_per_connection"`
	ConnectionsPerAccount      int      `json:"connections_per_account"`
	DrainTimeoutMS             int64    `json:"drain_timeout_ms"`
	AllowedOrigins             []string `json:"allowed_origins"`
}

func LoadPolicy(data []byte) (*Policy, error) {
	var policy Policy
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&policy); err != nil {
		return nil, ErrInvalidPolicy
	}
	if err := ensureEnd(decoder); err != nil || !policy.valid() {
		return nil, ErrInvalidPolicy
	}
	return &policy, nil
}

func (policy Policy) valid() bool {
	if policy.SchemaVersion != 1 || policy.WorldHz < 4 || policy.WorldHz > 10 || policy.FeedHistorySize < 1 || policy.FeedHistorySize > 512 ||
		policy.PlayerHistorySize < 256 || policy.PlayerHistorySize > 4096 || policy.PlayerHistoryTTLMS < 60_000 || policy.PlayerHistoryTTLMS > 3_600_000 ||
		policy.PlayerQueueMessages < 64 || policy.PlayerQueueMessages > 4096 || policy.PlayerQueueBytes < 65_536 || policy.PlayerQueueBytes > 16_777_216 ||
		policy.MessageBytes < 1_024 || policy.MessageBytes > 1_048_576 || policy.SubscriptionsPerConnection < 1 || policy.SubscriptionsPerConnection > 64 ||
		policy.ConnectionsPerAccount < 1 || policy.ConnectionsPerAccount > 10 || policy.DrainTimeoutMS < 1_000 || policy.DrainTimeoutMS > 60_000 || len(policy.AllowedOrigins) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, origin := range policy.AllowedOrigins {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Scheme != "http" && parsed.Scheme != "https" || seen[origin] {
			return false
		}
		seen[origin] = true
	}
	return true
}

func ensureEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrInvalidPolicy
	}
	return nil
}
