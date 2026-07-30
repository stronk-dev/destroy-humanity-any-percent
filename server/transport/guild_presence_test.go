package transport

import (
	"context"
	"errors"
	"testing"
	"time"

	"cloud-clicker/server/guild"
)

type guildPresenceSource struct {
	items     []guild.PresenceItem
	published []int64
	released  []int64
}

func (source *guildPresenceSource) ClaimPresence(context.Context, int, time.Duration) ([]guild.PresenceItem, error) {
	return source.items, nil
}
func (source *guildPresenceSource) MarkPresencePublished(_ context.Context, id int64, _ string) error {
	source.published = append(source.published, id)
	return nil
}
func (source *guildPresenceSource) ReleasePresence(_ context.Context, id int64, _ string) error {
	source.released = append(source.released, id)
	return nil
}

type envelopeCollector struct {
	envelopes []Envelope
	failAt    int
}

func (collector *envelopeCollector) Publish(envelope Envelope) error {
	if collector.failAt > 0 && len(collector.envelopes)+1 == collector.failAt {
		return errors.New("publish")
	}
	collector.envelopes = append(collector.envelopes, envelope)
	return nil
}

func TestGuildPresenceRelayPublishesStrictMembershipEnvelopes(t *testing.T) {
	now := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	source := &guildPresenceSource{items: []guild.PresenceItem{
		{ID: 1, ClaimToken: "018f0000-0000-4000-8000-000000000001", GuildID: "018f0000-0000-7000-8000-000000000010", AccountID: "018f0000-0000-4000-8000-000000000011", Kind: "joined", GuildRevision: 2, ActiveCount: 2, OccurredAt: now},
		{ID: 2, ClaimToken: "018f0000-0000-4000-8000-000000000002", GuildID: "018f0000-0000-7000-8000-000000000010", AccountID: "018f0000-0000-4000-8000-000000000011", Kind: "left", GuildRevision: 3, ActiveCount: 1, OccurredAt: now},
	}}
	collector := &envelopeCollector{}
	relay, err := NewGuildPresenceRelay(source, collector, "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	count, err := relay.Flush(context.Background())
	if err != nil || count != 2 || len(collector.envelopes) != 2 || collector.envelopes[0].Channel != "guild:"+source.items[0].GuildID || collector.envelopes[0].Kind != "presence" {
		t.Fatalf("count=%d envelopes=%+v err=%v", count, collector.envelopes, err)
	}
	for _, envelope := range collector.envelopes {
		if _, err := Encode(envelope, 64<<10); err != nil {
			t.Fatal(err)
		}
	}
}
