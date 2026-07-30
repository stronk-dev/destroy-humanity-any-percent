package transport

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"cloud-clicker/server/guild"
)

type GuildPresenceSource interface {
	ClaimPresence(context.Context, int, time.Duration) ([]guild.PresenceItem, error)
	MarkPresencePublished(context.Context, int64, string) error
	ReleasePresence(context.Context, int64, string) error
}

type GuildPresenceRelay struct {
	source        GuildPresenceSource
	publisher     EnvelopePublisher
	constantsHash string
	batchSize     int
	lease         time.Duration
}

func NewGuildPresenceRelay(source GuildPresenceSource, publisher EnvelopePublisher, constantsHash string) (*GuildPresenceRelay, error) {
	if source == nil || publisher == nil || !hashPattern.MatchString(constantsHash) {
		return nil, ErrRelay
	}
	return &GuildPresenceRelay{source: source, publisher: publisher, constantsHash: constantsHash, batchSize: 64, lease: 30 * time.Second}, nil
}

func (relay *GuildPresenceRelay) Flush(ctx context.Context) (int, error) {
	items, err := relay.source.ClaimPresence(ctx, relay.batchSize, relay.lease)
	if err != nil {
		return 0, err
	}
	published := 0
	for index, item := range items {
		joined, left := []string{}, []string{}
		switch item.Kind {
		case "joined":
			joined = append(joined, item.AccountID)
		case "left":
			left = append(left, item.AccountID)
		default:
			return published, relay.releaseTail(ctx, items, index, ErrInvalidPolicy)
		}
		payload, _ := json.Marshal(map[string]any{"joined": joined, "left": left, "count": item.ActiveCount})
		err := relay.publisher.Publish(Envelope{Version: WireVersion, Channel: "guild:" + item.GuildID,
			Kind: "presence", Revision: item.GuildRevision, ConstantsHash: relay.constantsHash,
			Timestamp: item.OccurredAt, Payload: payload})
		if err != nil {
			return published, relay.releaseTail(ctx, items, index, err)
		}
		if err := relay.source.MarkPresencePublished(ctx, item.ID, item.ClaimToken); err != nil {
			return published, relay.releaseTail(ctx, items, index+1, err)
		}
		published++
	}
	return published, nil
}

func (relay *GuildPresenceRelay) releaseTail(ctx context.Context, items []guild.PresenceItem, from int, cause error) error {
	combined := cause
	for _, item := range items[from:] {
		combined = errors.Join(combined, relay.source.ReleasePresence(ctx, item.ID, item.ClaimToken))
	}
	return combined
}
