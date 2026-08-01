package transport

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"cloud-clicker/server/save"
)

var ErrRelay = errors.New("transport player relay failed")

type PlayerSource interface {
	ClaimPlayerOutbox(context.Context, int, time.Duration) ([]save.PlayerOutboxItem, error)
	MarkPlayerPublished(context.Context, int64, string) error
	ReleasePlayerClaim(context.Context, int64, string) error
	DeferPlayerClaim(context.Context, int64, string, string, time.Duration) error
	FailPlayerClaim(context.Context, int64, string, string, int) (bool, error)
}

type EnvelopePublisher interface {
	Publish(Envelope) error
}

type RelayInvariant struct {
	Kind         string
	OutboxID     int64
	FounderID    string
	SourceID     string
	AttemptCount int
	Detail       string
}

type RelayInvariantSink interface {
	ReportRelayInvariant(RelayInvariant)
}

type PlayerRelay struct {
	source    PlayerSource
	publisher EnvelopePublisher
	sink      RelayInvariantSink
	batchSize int
	lease     time.Duration
	backoff   time.Duration
	mu        sync.Mutex
}

const playerFailureLimit = 5

func NewPlayerRelay(source PlayerSource, publisher EnvelopePublisher, sink RelayInvariantSink) (*PlayerRelay, error) {
	if source == nil || publisher == nil || sink == nil {
		return nil, ErrRelay
	}
	return &PlayerRelay{source: source, publisher: publisher, sink: sink, batchSize: 64, lease: 30 * time.Second, backoff: time.Second}, nil
}

func (relay *PlayerRelay) Flush(ctx context.Context) (int, error) {
	relay.mu.Lock()
	defer relay.mu.Unlock()
	items, err := relay.source.ClaimPlayerOutbox(ctx, relay.batchSize, relay.lease)
	if err != nil {
		return 0, err
	}
	published := 0
	for index, item := range items {
		envelope := Envelope{
			Version: WireVersion, Channel: "player:" + item.FounderID, Kind: item.MessageKind, Revision: item.Revision,
			ConstantsHash: item.ConstantsHash, Timestamp: item.OccurredAt, Payload: item.Payload,
		}
		if err := relay.publisher.Publish(envelope); err != nil {
			return published, relay.failBatch(ctx, items, index, err, errors.Is(err, ErrInvalidPolicy))
		}
		if err := relay.source.MarkPlayerPublished(ctx, item.ID, item.ClaimToken); err != nil {
			return published, relay.failBatch(ctx, items, index, err, false)
		}
		published++
	}
	return published, nil
}

func (relay *PlayerRelay) failBatch(ctx context.Context, items []save.PlayerOutboxItem, failedIndex int, cause error, deterministic bool) error {
	failed := items[failedIndex]
	detail := failureDetail(cause)
	var dead bool
	var failErr error
	var resyncErr error
	if deterministic {
		dead, failErr = relay.source.FailPlayerClaim(ctx, failed.ID, failed.ClaimToken, detail, playerFailureLimit)
		if dead {
			relay.sink.ReportRelayInvariant(RelayInvariant{Kind: "player_message_dead_letter", OutboxID: failed.ID, FounderID: failed.FounderID,
				SourceID: failed.SourceID, AttemptCount: failed.AttemptCount + 1, Detail: detail})
			if failed.MessageKind == "event" {
				payload, _ := json.Marshal(map[string]string{"code": "resync_required"})
				resyncErr = relay.publisher.Publish(Envelope{Version: WireVersion, Channel: "player:" + failed.FounderID, Kind: "system",
					Revision: failed.Revision, ConstantsHash: failed.ConstantsHash, Timestamp: time.Now().UTC(), Payload: payload})
				if resyncErr != nil {
					relay.sink.ReportRelayInvariant(RelayInvariant{Kind: "player_resync_signal_failed", OutboxID: failed.ID, FounderID: failed.FounderID,
						SourceID: failed.SourceID, AttemptCount: failed.AttemptCount + 1, Detail: failureDetail(resyncErr)})
				}
			}
		}
	} else {
		failErr = relay.source.DeferPlayerClaim(ctx, failed.ID, failed.ClaimToken, detail, relay.backoff)
	}
	var releaseErr error
	for _, item := range items[failedIndex+1:] {
		if err := relay.source.ReleasePlayerClaim(ctx, item.ID, item.ClaimToken); err != nil {
			releaseErr = errors.Join(releaseErr, err)
		}
	}
	return errors.Join(cause, failErr, resyncErr, releaseErr)
}

func failureDetail(err error) string {
	detail := strings.TrimSpace(err.Error())
	if detail == "" {
		detail = ErrRelay.Error()
	}
	runes := []rune(detail)
	if len(runes) > 512 {
		detail = string(runes[:512])
	}
	return detail
}
