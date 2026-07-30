package transport

import (
	"sync"
	"time"
)

type Position struct {
	Offset uint64 `json:"offset"`
}

type Publication struct {
	Position Position
	Data     []byte
	At       time.Time
}

type History struct {
	mu       sync.Mutex
	policy   Policy
	channels map[string][]Publication
	offsets  map[string]uint64
}

func NewHistory(policy Policy) (*History, error) {
	if !policy.valid() {
		return nil, ErrInvalidPolicy
	}
	return &History{policy: policy, channels: map[string][]Publication{}, offsets: map[string]uint64{}}, nil
}

func (history *History) Publish(channel string, data []byte, now time.Time) (Position, error) {
	if channel == "" || len(data) == 0 || now.IsZero() {
		return Position{}, ErrInvalidPolicy
	}
	history.mu.Lock()
	defer history.mu.Unlock()
	history.offsets[channel]++
	position := Position{Offset: history.offsets[channel]}
	entry := Publication{Position: position, Data: append([]byte(nil), data...), At: now.UTC()}
	if channel == "world" {
		history.channels[channel] = []Publication{entry}
		return position, nil
	}
	items := append(history.channels[channel], entry)
	limit := history.policy.FeedHistorySize
	ttl := time.Duration(history.policy.PlayerHistoryTTLMS) * time.Millisecond
	if isPlayerChannel(channel) {
		limit = history.policy.PlayerHistorySize
		cutoff := now.Add(-ttl)
		first := 0
		for first < len(items) && items[first].At.Before(cutoff) {
			first++
		}
		items = items[first:]
	}
	if len(items) > limit {
		items = items[len(items)-limit:]
	}
	history.channels[channel] = items
	return position, nil
}

func (history *History) Recover(channel string, after Position, now time.Time) ([]Publication, bool) {
	history.mu.Lock()
	defer history.mu.Unlock()
	items := history.channels[channel]
	if channel == "world" {
		if len(items) == 0 {
			return nil, true
		}
		return clonePublications(items[len(items)-1:]), true
	}
	if isPlayerChannel(channel) && len(items) > 0 {
		cutoff := now.Add(-time.Duration(history.policy.PlayerHistoryTTLMS) * time.Millisecond)
		if items[0].At.Before(cutoff) || after.Offset+1 < items[0].Position.Offset {
			return nil, false
		}
	}
	var result []Publication
	for _, item := range items {
		if item.Position.Offset > after.Offset {
			result = append(result, item)
		}
	}
	return clonePublications(result), true
}

func clonePublications(source []Publication) []Publication {
	result := make([]Publication, len(source))
	for index, item := range source {
		result[index] = item
		result[index].Data = append([]byte(nil), item.Data...)
	}
	return result
}
