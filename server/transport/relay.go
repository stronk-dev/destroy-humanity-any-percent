package transport

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"cloud-clicker/server/save"
)

var ErrRelay = errors.New("transport receipt relay failed")

type ReceiptSource interface {
	ClaimReceiptOutbox(context.Context, int, time.Duration) ([]save.ReceiptOutboxItem, error)
	MarkReceiptPublished(context.Context, int64, string) error
	ReleaseReceiptClaim(context.Context, int64, string) error
	FailReceiptClaim(context.Context, int64, string, string, int) (bool, error)
}

type EnvelopePublisher interface {
	Publish(Envelope) error
}

type RelayInvariant struct {
	Kind         string
	OutboxID     int64
	FounderID    string
	IntentID     string
	AttemptCount int
	Detail       string
}

type RelayInvariantSink interface {
	ReportRelayInvariant(RelayInvariant)
}

type ReceiptRelay struct {
	source    ReceiptSource
	publisher EnvelopePublisher
	sink      RelayInvariantSink
	batchSize int
	lease     time.Duration
	mu        sync.Mutex
}

const receiptFailureLimit = 5

func NewReceiptRelay(source ReceiptSource, publisher EnvelopePublisher, sink RelayInvariantSink) (*ReceiptRelay, error) {
	if source == nil || publisher == nil || sink == nil {
		return nil, ErrRelay
	}
	return &ReceiptRelay{source: source, publisher: publisher, sink: sink, batchSize: 64, lease: 30 * time.Second}, nil
}

func (relay *ReceiptRelay) Flush(ctx context.Context) (int, error) {
	relay.mu.Lock()
	defer relay.mu.Unlock()
	items, err := relay.source.ClaimReceiptOutbox(ctx, relay.batchSize, relay.lease)
	if err != nil {
		return 0, err
	}
	published := 0
	for index, item := range items {
		envelope := Envelope{
			Version: WireVersion, Channel: "player:" + item.FounderID, Kind: "receipt", Revision: item.Revision,
			ConstantsHash: item.ConstantsHash, Timestamp: item.OccurredAt, Payload: item.Receipt,
		}
		if err := relay.publisher.Publish(envelope); err != nil {
			return published, relay.failBatch(ctx, items, index, err)
		}
		if err := relay.source.MarkReceiptPublished(ctx, item.ID, item.ClaimToken); err != nil {
			return published, relay.failBatch(ctx, items, index, err)
		}
		published++
	}
	return published, nil
}

func (relay *ReceiptRelay) failBatch(ctx context.Context, items []save.ReceiptOutboxItem, failedIndex int, cause error) error {
	failed := items[failedIndex]
	detail := failureDetail(cause)
	dead, failErr := relay.source.FailReceiptClaim(ctx, failed.ID, failed.ClaimToken, detail, receiptFailureLimit)
	if dead {
		relay.sink.ReportRelayInvariant(RelayInvariant{Kind: "receipt_dead_letter", OutboxID: failed.ID, FounderID: failed.FounderID,
			IntentID: failed.IntentID, AttemptCount: failed.AttemptCount + 1, Detail: detail})
	}
	var releaseErr error
	for _, item := range items[failedIndex+1:] {
		if err := relay.source.ReleaseReceiptClaim(ctx, item.ID, item.ClaimToken); err != nil {
			releaseErr = errors.Join(releaseErr, err)
		}
	}
	return errors.Join(cause, failErr, releaseErr)
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
