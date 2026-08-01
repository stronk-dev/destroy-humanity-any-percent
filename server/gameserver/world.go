package gameserver

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"cloud-clicker/server/transport"
)

var ErrWorldAggregator = errors.New("invalid world aggregator")

type WorldSample struct {
	Planet     transport.WorldPlanet
	Commons    transport.WorldCommons
	Population transport.WorldPopulation
	Milestones transport.WorldMilestones
	Epoch      transport.WorldEpoch
}

type WorldSource interface {
	SampleWorld(context.Context) (WorldSample, error)
}

type WorldPublisher interface {
	Publish(transport.Envelope) error
}

type WorldAggregator struct {
	source        WorldSource
	publisher     WorldPublisher
	constantsHash string
	cadence       time.Duration
	clock         func() time.Time
	mu            sync.Mutex
	revision      int64
}

func NewWorldAggregator(source WorldSource, publisher WorldPublisher, constantsHash string, worldHz int, clock func() time.Time) (*WorldAggregator, error) {
	if source == nil || publisher == nil || !validConstantsHash(constantsHash) || worldHz < 4 || worldHz > 10 {
		return nil, ErrWorldAggregator
	}
	if clock == nil {
		clock = time.Now
	}
	return &WorldAggregator{source: source, publisher: publisher, constantsHash: constantsHash,
		cadence: time.Second / time.Duration(worldHz), clock: clock}, nil
}

func (aggregator *WorldAggregator) Run(ctx context.Context) error {
	if aggregator == nil {
		return ErrWorldAggregator
	}
	ticker := time.NewTicker(aggregator.cadence)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := aggregator.PublishOnce(ctx); err != nil {
				return err
			}
		}
	}
}

func (aggregator *WorldAggregator) PublishOnce(ctx context.Context) error {
	if aggregator == nil {
		return ErrWorldAggregator
	}
	aggregator.mu.Lock()
	defer aggregator.mu.Unlock()
	sample, err := aggregator.source.SampleWorld(ctx)
	if err != nil {
		return err
	}
	revision := aggregator.revision + 1
	snapshot := transport.WorldSnapshot{Version: 1, WorldRev: revision, Planet: sample.Planet,
		Commons: sample.Commons, Population: sample.Population, Milestones: sample.Milestones, Epoch: sample.Epoch}
	if err := transport.ValidateWorldSnapshot(snapshot); err != nil {
		return errors.Join(ErrWorldAggregator, err)
	}
	state, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		Scope string          `json:"scope"`
		Rev   int64           `json:"rev"`
		State json.RawMessage `json:"state"`
	}{Scope: "world", Rev: revision, State: state})
	if err != nil {
		return err
	}
	if err := aggregator.publisher.Publish(transport.Envelope{Version: transport.WireVersion, Channel: "world", Kind: "snapshot",
		Revision: revision, ConstantsHash: aggregator.constantsHash, Timestamp: aggregator.clock().UTC(), Payload: payload}); err != nil {
		return err
	}
	aggregator.revision = revision
	return nil
}
