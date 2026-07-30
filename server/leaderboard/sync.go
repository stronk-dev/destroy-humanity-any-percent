package leaderboard

import (
	"context"
	"errors"
	"time"

	"cloud-clicker/server/epochseed"
)

type SeedSynchronizer struct {
	repository *Repository
	bundle     epochseed.Bundle
	clock      func() time.Time
}

func NewSeedSynchronizer(repository *Repository, bundle epochseed.Bundle, clock func() time.Time) (*SeedSynchronizer, error) {
	if repository == nil || bundle.Hash == "" || clock == nil {
		return nil, ErrInvalidEpoch
	}
	return &SeedSynchronizer{repository: repository, bundle: bundle, clock: clock}, nil
}

func (synchronizer *SeedSynchronizer) Sync(ctx context.Context) (string, error) {
	if synchronizer == nil || synchronizer.repository == nil || synchronizer.clock == nil {
		return "", ErrInvalidEpoch
	}
	if err := synchronizer.repository.ReconcileSeed(ctx, synchronizer.bundle, synchronizer.clock()); err != nil {
		return "", errors.Join(ErrInvalidEpoch, err)
	}
	return synchronizer.bundle.Hash, nil
}
