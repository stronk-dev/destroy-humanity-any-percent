package gameserver

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidJob = errors.New("invalid gameserver job")

type PeriodicJob struct {
	interval time.Duration
	run      func(context.Context) error
}

func NewPeriodicJob(interval time.Duration, run func(context.Context) error) (*PeriodicJob, error) {
	if interval <= 0 || run == nil {
		return nil, ErrInvalidJob
	}
	return &PeriodicJob{interval: interval, run: run}, nil
}

func (job *PeriodicJob) Run(ctx context.Context) error {
	if job == nil || job.run == nil || job.interval <= 0 {
		return ErrInvalidJob
	}
	ticker := time.NewTicker(job.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := job.run(ctx); err != nil {
				return err
			}
		}
	}
}

func (job *PeriodicJob) Prime(ctx context.Context) error {
	if job == nil || job.run == nil || job.interval <= 0 {
		return ErrInvalidJob
	}
	return job.run(ctx)
}
