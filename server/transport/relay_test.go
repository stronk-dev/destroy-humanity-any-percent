package transport

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

type memoryPlayerSource struct {
	items    []save.PlayerOutboxItem
	marked   []int64
	released []int64
	deferred []int64
	failed   []int64
	markErr  int64
}

func (source *memoryPlayerSource) ClaimPlayerOutbox(context.Context, int, time.Duration) ([]save.PlayerOutboxItem, error) {
	return append([]save.PlayerOutboxItem(nil), source.items...), nil
}
func (source *memoryPlayerSource) MarkPlayerPublished(_ context.Context, id int64, _ string) error {
	if source.markErr == id {
		return errors.New("mark unavailable")
	}
	source.marked = append(source.marked, id)
	return nil
}
func (source *memoryPlayerSource) ReleasePlayerClaim(_ context.Context, id int64, _ string) error {
	source.released = append(source.released, id)
	return nil
}
func (source *memoryPlayerSource) DeferPlayerClaim(_ context.Context, id int64, _ string, _ string, _ time.Duration) error {
	source.deferred = append(source.deferred, id)
	return nil
}
func (source *memoryPlayerSource) FailPlayerClaim(_ context.Context, id int64, _ string, _ string, maxAttempts int) (bool, error) {
	source.failed = append(source.failed, id)
	for index := range source.items {
		if source.items[index].ID == id {
			source.items[index].AttemptCount++
			return source.items[index].AttemptCount >= maxAttempts, nil
		}
	}
	return false, errors.New("missing item")
}

type memoryRelaySink struct{ reports []RelayInvariant }

func (sink *memoryRelaySink) ReportRelayInvariant(report RelayInvariant) {
	sink.reports = append(sink.reports, report)
}

type memoryPublisher struct {
	envelopes     []Envelope
	failAt        int
	deterministic bool
}

func (publisher *memoryPublisher) Publish(envelope Envelope) error {
	if publisher.failAt > 0 && len(publisher.envelopes)+1 == publisher.failAt {
		if publisher.deterministic {
			return ErrInvalidPolicy
		}
		return errors.New("publish unavailable")
	}
	publisher.envelopes = append(publisher.envelopes, envelope)
	return nil
}

func TestPlayerRelayMapsMixedCommittedOutboxExactly(t *testing.T) {
	when := time.Date(2026, 7, 29, 15, 0, 0, 0, time.UTC)
	receipt := json.RawMessage(`{"intent_id":"01985555-0010-7000-8000-000000000010","outcome":"applied","new_revision":9}`)
	event := json.RawMessage(`{"event_id":"01985555-0009-7000-8000-000000000009","kind":"gate_crossed","scope":"company","rev":8,"payload":{}}`)
	source := &memoryPlayerSource{items: []save.PlayerOutboxItem{{
		ID: 6, ClaimToken: "01985555-0011-7000-8000-000000000011", FounderID: "founder", MessageKind: "event", Scope: "company", Revision: 8,
		SourceID: "01985555-0009-7000-8000-000000000009", ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Payload: event, OccurredAt: when,
	}, {
		ID: 7, ClaimToken: "01985555-0011-7000-8000-000000000011", FounderID: "founder", Revision: 9,
		MessageKind: "receipt", Scope: "company", SourceID: "01985555-0010-7000-8000-000000000010",
		ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Payload: receipt, OccurredAt: when,
	}}}
	publisher := &memoryPublisher{}
	relay, err := NewPlayerRelay(source, publisher, &memoryRelaySink{})
	if err != nil {
		t.Fatal(err)
	}
	count, err := relay.Flush(context.Background())
	if err != nil || count != 2 || len(source.marked) != 2 || len(source.released) != 0 || len(publisher.envelopes) != 2 {
		t.Fatalf("count=%d marked=%v released=%v envelopes=%v err=%v", count, source.marked, source.released, publisher.envelopes, err)
	}
	if publisher.envelopes[0].Kind != "event" || string(publisher.envelopes[0].Payload) != string(event) {
		t.Fatalf("event envelope=%+v", publisher.envelopes[0])
	}
	envelope := publisher.envelopes[1]
	if envelope.Channel != "player:founder" || envelope.Kind != "receipt" || envelope.Revision != 9 ||
		!envelope.Timestamp.Equal(when) || string(envelope.Payload) != string(receipt) {
		t.Fatalf("envelope=%+v", envelope)
	}
}

func TestPlayerRelayReleasesFailedPublication(t *testing.T) {
	source := &memoryPlayerSource{items: []save.PlayerOutboxItem{{
		ID: 8, ClaimToken: "01985555-0012-7000-8000-000000000012", FounderID: "founder", Revision: 1,
		MessageKind: "receipt", Scope: "company", SourceID: "01985555-0010-7000-8000-000000000010", ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Payload: json.RawMessage(`{"outcome":"rejected"}`), OccurredAt: time.Now().UTC(),
	}}}
	relay, _ := NewPlayerRelay(source, &memoryPublisher{failAt: 1}, &memoryRelaySink{})
	if count, err := relay.Flush(context.Background()); err == nil || count != 0 || len(source.marked) != 0 || len(source.failed) != 0 || len(source.deferred) != 1 {
		t.Fatalf("count=%d marked=%v failed=%v deferred=%v err=%v", count, source.marked, source.failed, source.deferred, err)
	}
}

func TestPlayerRelayReleasesWholeRemainderAndPreservesOrder(t *testing.T) {
	items := make([]save.PlayerOutboxItem, 3)
	for index := range items {
		items[index] = save.PlayerOutboxItem{ID: int64(index + 1), ClaimToken: "01985555-0012-7000-8000-000000000012",
			FounderID: "founder", MessageKind: "receipt", Scope: "company", SourceID: "01985555-0010-7000-8000-000000000010", Revision: int64(index + 1),
			ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Payload:       json.RawMessage(`{"outcome":"applied"}`), OccurredAt: time.Now().UTC()}
	}
	source := &memoryPlayerSource{items: items}
	relay, _ := NewPlayerRelay(source, &memoryPublisher{failAt: 2}, &memoryRelaySink{})
	count, err := relay.Flush(context.Background())
	if err == nil || count != 1 || len(source.marked) != 1 || source.marked[0] != 1 ||
		len(source.deferred) != 1 || source.deferred[0] != 2 || len(source.released) != 1 || source.released[0] != 3 {
		t.Fatalf("count=%d marked=%v deferred=%v released=%v err=%v", count, source.marked, source.deferred, source.released, err)
	}
}

func TestPlayerRelayMarkFailureAlsoReleasesRemainder(t *testing.T) {
	items := []save.PlayerOutboxItem{
		{ID: 1, ClaimToken: "01985555-0012-7000-8000-000000000012", FounderID: "founder", MessageKind: "receipt", Scope: "company", SourceID: "01985555-0010-7000-8000-000000000010", Revision: 1, ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Payload: json.RawMessage(`{"outcome":"applied"}`), OccurredAt: time.Now().UTC()},
		{ID: 2, ClaimToken: "01985555-0013-7000-8000-000000000013", FounderID: "other", MessageKind: "receipt", Scope: "company", SourceID: "01985555-0011-7000-8000-000000000011", Revision: 1, ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Payload: json.RawMessage(`{"outcome":"applied"}`), OccurredAt: time.Now().UTC()},
	}
	source := &memoryPlayerSource{items: items, markErr: 1}
	relay, _ := NewPlayerRelay(source, &memoryPublisher{}, &memoryRelaySink{})
	count, err := relay.Flush(context.Background())
	if err == nil || count != 0 || len(source.deferred) != 1 || source.deferred[0] != 1 || len(source.released) != 1 || source.released[0] != 2 {
		t.Fatalf("count=%d deferred=%v released=%v err=%v", count, source.deferred, source.released, err)
	}
}

func TestPlayerRelayDeadLettersAndReportsBoundedPoison(t *testing.T) {
	source := &memoryPlayerSource{items: []save.PlayerOutboxItem{{
		ID: 9, ClaimToken: "01985555-0012-7000-8000-000000000012", FounderID: "founder",
		MessageKind: "receipt", Scope: "company", SourceID: "01985555-0010-7000-8000-000000000010", Revision: 1,
		ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Payload:       json.RawMessage(`{"outcome":"rejected"}`), OccurredAt: time.Now().UTC(),
	}}}
	sink := &memoryRelaySink{}
	relay, _ := NewPlayerRelay(source, &memoryPublisher{failAt: 1, deterministic: true}, sink)
	for attempt := 1; attempt <= playerFailureLimit; attempt++ {
		if _, err := relay.Flush(context.Background()); err == nil {
			t.Fatalf("attempt %d succeeded", attempt)
		}
	}
	if len(sink.reports) != 1 || sink.reports[0].Kind != "player_message_dead_letter" || sink.reports[0].AttemptCount != playerFailureLimit {
		t.Fatalf("reports=%+v", sink.reports)
	}
}
