package gameserver

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/transport"
)

type fixedWorldSource struct {
	sample WorldSample
	err    error
}

func (source fixedWorldSource) SampleWorld(context.Context) (WorldSample, error) {
	return source.sample, source.err
}

type recordingWorldPublisher struct {
	mu        sync.Mutex
	envelopes []transport.Envelope
	err       error
}

func (publisher *recordingWorldPublisher) Publish(envelope transport.Envelope) error {
	publisher.mu.Lock()
	defer publisher.mu.Unlock()
	if publisher.err != nil {
		return publisher.err
	}
	publisher.envelopes = append(publisher.envelopes, envelope)
	return nil
}

func TestWorldAggregatorMonotonicClosedAndZeroHonest(t *testing.T) {
	publisher := &recordingWorldPublisher{}
	now := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	aggregator, err := NewWorldAggregator(fixedWorldSource{sample: WorldSample{
		Commons:    transport.WorldCommons{ServerHealthPPM: 750_000, ActiveFounders: 12, CompactMembers: 7},
		Population: transport.WorldPopulation{Online: 3, FoundersTotal: 20},
		Epoch:      transport.WorldEpoch{EpochID: 5, Name: "Phase 0"},
	}}, publisher, "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 4, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	for index := int64(1); index <= 1_000; index++ {
		if err := aggregator.PublishOnce(context.Background()); err != nil {
			t.Fatal(err)
		}
		envelope := publisher.envelopes[len(publisher.envelopes)-1]
		if envelope.Revision != index {
			t.Fatalf("revision=%d want=%d", envelope.Revision, index)
		}
		var payload struct {
			Scope string                  `json:"scope"`
			Rev   int64                   `json:"rev"`
			State transport.WorldSnapshot `json:"state"`
		}
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil || payload.Scope != "world" || payload.Rev != index || payload.State.WorldRev != index {
			t.Fatalf("payload=%+v err=%v", payload, err)
		}
		if payload.State.Planet != (transport.WorldPlanet{}) || payload.State.Milestones.ActiveID != nil || payload.State.Milestones.ProgressPPM != 0 {
			t.Fatalf("unshipped systems not zero-honest: %+v", payload.State)
		}
		if _, err := transport.Encode(envelope, 65_536); err != nil {
			t.Fatalf("wire rejected revision %d: %v", index, err)
		}
	}
}

func TestWorldAggregatorDoesNotAdvanceOnFailedPublish(t *testing.T) {
	injected := errors.New("publisher unavailable")
	publisher := &recordingWorldPublisher{err: injected}
	source := fixedWorldSource{sample: WorldSample{Epoch: transport.WorldEpoch{EpochID: 5, Name: "Phase 0"}}}
	aggregator, err := NewWorldAggregator(source, publisher, "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 4, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if err := aggregator.PublishOnce(context.Background()); !errors.Is(err, injected) {
		t.Fatalf("publish error=%v", err)
	}
	publisher.err = nil
	if err := aggregator.PublishOnce(context.Background()); err != nil || len(publisher.envelopes) != 1 || publisher.envelopes[0].Revision != 1 {
		t.Fatalf("envelopes=%+v err=%v", publisher.envelopes, err)
	}
}
