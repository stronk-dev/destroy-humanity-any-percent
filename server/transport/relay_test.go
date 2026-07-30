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
}

func (source *memoryReceiptSource) ClaimReceiptOutbox(context.Context, int, time.Duration) ([]save.ReceiptOutboxItem, error) {
	return append([]save.ReceiptOutboxItem(nil), source.items...), nil
}
func (source *memoryReceiptSource) MarkReceiptPublished(_ context.Context, id int64, _ string) error {
	source.marked = append(source.marked, id)
	return nil
}
func (source *memoryReceiptSource) ReleaseReceiptClaim(_ context.Context, id int64, _ string) error {
	source.released = append(source.released, id)
	return nil
}

type memoryPublisher struct {
	envelopes []Envelope
	failAt    int
}

func (publisher *memoryPublisher) Publish(envelope Envelope) error {
	if publisher.failAt > 0 && len(publisher.envelopes)+1 == publisher.failAt {
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
	relay, err := NewReceiptRelay(source, publisher)
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
	relay, _ := NewReceiptRelay(source, &memoryPublisher{failAt: 1})
	if count, err := relay.Flush(context.Background()); err == nil || count != 0 || len(source.marked) != 0 || len(source.released) != 1 {
		t.Fatalf("count=%d marked=%v released=%v err=%v", count, source.marked, source.released, err)
	}
}
