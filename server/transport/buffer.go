package transport

import (
	"errors"
	"sync"
)

var (
	ErrQueueOverflow      = errors.New("transport receipt queue overflow")
	ErrInvalidReservation = errors.New("invalid transport queue reservation")
)

// ConnectionQueue is the application-owned discipline layered over
// Centrifuge's byte-bounded writer. It enforces the second independent player
// bound (message count) and marks all but the newest queued world revision
// stale at the transport-write hook.
type ConnectionQueue struct {
	mu                  sync.Mutex
	privatePending      int
	privateByRevision   map[int64]int
	maxPrivateMessages  int
	latestWorldRevision int64
}

func NewConnectionQueue(policy Policy) (*ConnectionQueue, error) {
	if !policy.valid() {
		return nil, ErrInvalidPolicy
	}
	return &ConnectionQueue{maxPrivateMessages: policy.PlayerQueueMessages, privateByRevision: map[int64]int{}}, nil
}

func (queue *ConnectionQueue) ReservePlayer(revision int64) error {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if revision < 1 {
		return ErrInvalidReservation
	}
	if queue.privatePending >= queue.maxPrivateMessages {
		return ErrQueueOverflow
	}
	queue.privatePending++
	queue.privateByRevision[revision]++
	return nil
}

func (queue *ConnectionQueue) FinishPlayer(revision int64) bool {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if queue.privateByRevision[revision] == 0 {
		return false
	}
	queue.privatePending--
	queue.privateByRevision[revision]--
	if queue.privateByRevision[revision] == 0 {
		delete(queue.privateByRevision, revision)
	}
	return true
}

func (queue *ConnectionQueue) ResetPlayer() {
	queue.mu.Lock()
	queue.privatePending = 0
	clear(queue.privateByRevision)
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
