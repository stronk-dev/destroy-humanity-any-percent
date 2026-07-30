package transport

import (
	"errors"
	"sync"
)

var ErrQueueOverflow = errors.New("transport receipt queue overflow")

// ConnectionQueue is the application-owned discipline layered over
// Centrifuge's byte-bounded writer. It enforces the second independent player
// bound (message count) and marks all but the newest queued world revision
// stale at the transport-write hook.
type ConnectionQueue struct {
	mu                  sync.Mutex
	privatePending      int
	maxPrivateMessages  int
	latestWorldRevision int64
}

func NewConnectionQueue(policy Policy) (*ConnectionQueue, error) {
	if !policy.valid() {
		return nil, ErrInvalidPolicy
	}
	return &ConnectionQueue{maxPrivateMessages: policy.PlayerQueueMessages}, nil
}

func (queue *ConnectionQueue) ReservePlayer() error {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if queue.privatePending >= queue.maxPrivateMessages {
		return ErrQueueOverflow
	}
	queue.privatePending++
	return nil
}

func (queue *ConnectionQueue) FinishPlayer() {
	queue.mu.Lock()
	if queue.privatePending > 0 {
		queue.privatePending--
	}
	queue.mu.Unlock()
}

func (queue *ConnectionQueue) ResetPlayer() {
	queue.mu.Lock()
	queue.privatePending = 0
	queue.mu.Unlock()
}

func (queue *ConnectionQueue) ReserveWorld(revision int64) int64 {
	queue.mu.Lock()
	previous := queue.latestWorldRevision
	if revision >= previous {
		queue.latestWorldRevision = revision
	}
	queue.mu.Unlock()
	return previous
}

func (queue *ConnectionQueue) RollbackWorld(revision, previous int64) {
	queue.mu.Lock()
	if queue.latestWorldRevision == revision {
		queue.latestWorldRevision = previous
	}
	queue.mu.Unlock()
}

func (queue *ConnectionQueue) AllowWorldWrite(revision int64) bool {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return revision >= queue.latestWorldRevision
}
