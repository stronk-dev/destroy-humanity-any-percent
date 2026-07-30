package transport

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

type memoryReceiptSource struct {
	items    []save.ReceiptOutboxItem
	marked   []int64
	released []int64
	deferred []int64
	failed   []int64
	markErr  int64
}

func (source *memoryReceiptSource) ClaimReceiptOutbox(context.Context, int, time.Duration) ([]save.ReceiptOutboxItem, error) {
	return append([]save.ReceiptOutboxItem(nil), source.items...), nil
}
func (source *memoryReceiptSource) MarkReceiptPublished(_ context.Context, id int64, _ string) error {
	if source.markErr == id {
		return errors.New("mark unavailable")
	}
	source.marked = append(source.marked, id)
	return nil
}
func (source *memoryReceiptSource) ReleaseReceiptClaim(_ context.Context, id int64, _ string) error {
	source.released = append(source.released, id)
	return nil
}
func (source *memoryReceiptSource) DeferReceiptClaim(_ context.Context, id int64, _ string, _ string, _ time.Duration) error {
	source.deferred = append(source.deferred, id)
	return nil
}
func (source *memoryReceiptSource) FailReceiptClaim(_ context.Context, id int64, _ string, _ string, maxAttempts int) (bool, error) {
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

func TestReceiptRelayMapsCommittedOutboxExactly(t *testing.T) {
	when := time.Date(2026, 7, 29, 15, 0, 0, 0, time.UTC)
	receipt := json.RawMessage(`{"intent_id":"01985555-0010-7000-8000-000000000010","outcome":"applied","new_revision":9}`)
	source := &memoryReceiptSource{items: []save.ReceiptOutboxItem{{
		ID: 7, ClaimToken: "01985555-0011-7000-8000-000000000011", FounderID: "founder", Revision: 9,
		ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Receipt: receipt, OccurredAt: when,
	}}}
	publisher := &memoryPublisher{}
	relay, err := NewReceiptRelay(source, publisher, &memoryRelaySink{})
	if err != nil {
		t.Fatal(err)
	}
	count, err := relay.Flush(context.Background())
	if err != nil || count != 1 || len(source.marked) != 1 || len(source.released) != 0 || len(publisher.envelopes) != 1 {
		t.Fatalf("count=%d marked=%v released=%v envelopes=%v err=%v", count, source.marked, source.released, publisher.envelopes, err)
	}
	envelope := publisher.envelopes[0]
	if envelope.Channel != "player:founder" || envelope.Kind != "receipt" || envelope.Revision != 9 ||
		!envelope.Timestamp.Equal(when) || string(envelope.Payload) != string(receipt) {
		t.Fatalf("envelope=%+v", envelope)
	}
}

func TestReceiptRelayReleasesFailedPublication(t *testing.T) {
	source := &memoryReceiptSource{items: []save.ReceiptOutboxItem{{
		ID: 8, ClaimToken: "01985555-0012-7000-8000-000000000012", FounderID: "founder", Revision: 1,
		ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Receipt: json.RawMessage(`{"outcome":"rejected"}`), OccurredAt: time.Now().UTC(),
	}}}
	relay, _ := NewReceiptRelay(source, &memoryPublisher{failAt: 1}, &memoryRelaySink{})
	if count, err := relay.Flush(context.Background()); err == nil || count != 0 || len(source.marked) != 0 || len(source.failed) != 0 || len(source.deferred) != 1 {
		t.Fatalf("count=%d marked=%v failed=%v deferred=%v err=%v", count, source.marked, source.failed, source.deferred, err)
	}
}

func TestReceiptRelayReleasesWholeRemainderAndPreservesOrder(t *testing.T) {
	items := make([]save.ReceiptOutboxItem, 3)
	for index := range items {
		items[index] = save.ReceiptOutboxItem{ID: int64(index + 1), ClaimToken: "01985555-0012-7000-8000-000000000012",
			FounderID: "founder", IntentID: "01985555-0010-7000-8000-000000000010", Revision: int64(index + 1),
			ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Receipt:       json.RawMessage(`{"outcome":"applied"}`), OccurredAt: time.Now().UTC()}
	}
	source := &memoryReceiptSource{items: items}
	relay, _ := NewReceiptRelay(source, &memoryPublisher{failAt: 2}, &memoryRelaySink{})
	count, err := relay.Flush(context.Background())
	if err == nil || count != 1 || len(source.marked) != 1 || source.marked[0] != 1 ||
		len(source.deferred) != 1 || source.deferred[0] != 2 || len(source.released) != 1 || source.released[0] != 3 {
		t.Fatalf("count=%d marked=%v deferred=%v released=%v err=%v", count, source.marked, source.deferred, source.released, err)
	}
}

func TestReceiptRelayMarkFailureAlsoReleasesRemainder(t *testing.T) {
	items := []save.ReceiptOutboxItem{
		{ID: 1, ClaimToken: "01985555-0012-7000-8000-000000000012", FounderID: "founder", IntentID: "01985555-0010-7000-8000-000000000010", Revision: 1, ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Receipt: json.RawMessage(`{"outcome":"applied"}`), OccurredAt: time.Now().UTC()},
		{ID: 2, ClaimToken: "01985555-0013-7000-8000-000000000013", FounderID: "other", IntentID: "01985555-0011-7000-8000-000000000011", Revision: 1, ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Receipt: json.RawMessage(`{"outcome":"applied"}`), OccurredAt: time.Now().UTC()},
	}
	source := &memoryReceiptSource{items: items, markErr: 1}
	relay, _ := NewReceiptRelay(source, &memoryPublisher{}, &memoryRelaySink{})
	count, err := relay.Flush(context.Background())
	if err == nil || count != 0 || len(source.deferred) != 1 || source.deferred[0] != 1 || len(source.released) != 1 || source.released[0] != 2 {
		t.Fatalf("count=%d deferred=%v released=%v err=%v", count, source.deferred, source.released, err)
	}
}

func TestReceiptRelayDeadLettersAndReportsBoundedPoison(t *testing.T) {
	source := &memoryReceiptSource{items: []save.ReceiptOutboxItem{{
		ID: 9, ClaimToken: "01985555-0012-7000-8000-000000000012", FounderID: "founder",
		IntentID: "01985555-0010-7000-8000-000000000010", Revision: 1,
		ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Receipt:       json.RawMessage(`{"outcome":"rejected"}`), OccurredAt: time.Now().UTC(),
	}}}
	sink := &memoryRelaySink{}
	relay, _ := NewReceiptRelay(source, &memoryPublisher{failAt: 1, deterministic: true}, sink)
	for attempt := 1; attempt <= receiptFailureLimit; attempt++ {
		if _, err := relay.Flush(context.Background()); err == nil {
			t.Fatalf("attempt %d succeeded", attempt)
		}
	}
	if len(sink.reports) != 1 || sink.reports[0].Kind != "receipt_dead_letter" || sink.reports[0].AttemptCount != receiptFailureLimit {
		t.Fatalf("reports=%+v", sink.reports)
	}
}
