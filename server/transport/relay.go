package transport

import (
	"context"
	"errors"
	"sync"
	"time"

	"cloud-clicker/server/save"
)

var ErrRelay = errors.New("transport receipt relay failed")

type ReceiptSource interface {
	ClaimReceiptOutbox(context.Context, int, time.Duration) ([]save.ReceiptOutboxItem, error)
	MarkReceiptPublished(context.Context, int64, string) error
	ReleaseReceiptClaim(context.Context, int64, string) error
}

type EnvelopePublisher interface {
	Publish(Envelope) error
}

type ReceiptRelay struct {
	source    ReceiptSource
	publisher EnvelopePublisher
	batchSize int
	lease     time.Duration
	mu        sync.Mutex
}

func NewReceiptRelay(source ReceiptSource, publisher EnvelopePublisher) (*ReceiptRelay, error) {
	if source == nil || publisher == nil {
		return nil, ErrRelay
	}
	return &ReceiptRelay{source: source, publisher: publisher, batchSize: 64, lease: 30 * time.Second}, nil
}

func (relay *ReceiptRelay) Flush(ctx context.Context) (int, error) {
	relay.mu.Lock()
	defer relay.mu.Unlock()
	items, err := relay.source.ClaimReceiptOutbox(ctx, relay.batchSize, relay.lease)
	if err != nil {
		return 0, err
	}
	published := 0
	for _, item := range items {
		envelope := Envelope{
			Version: WireVersion, Channel: "player:" + item.FounderID, Kind: "receipt", Revision: item.Revision,
			ConstantsHash: item.ConstantsHash, Timestamp: item.OccurredAt, Payload: item.Receipt,
		}
		if err := relay.publisher.Publish(envelope); err != nil {
			_ = relay.source.ReleaseReceiptClaim(ctx, item.ID, item.ClaimToken)
			return published, err
		}
		if err := relay.source.MarkReceiptPublished(ctx, item.ID, item.ClaimToken); err != nil {
			return published, err
		}
		published++
	}
	return published, nil
}
