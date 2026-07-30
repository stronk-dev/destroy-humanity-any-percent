package transport

import (
	"errors"
	"sync"
)

var ErrQueueOverflow = errors.New("transport receipt queue overflow")

type QueuedMessage struct {
	Channel string
	Kind    string
	Data    []byte
}

type ConnectionQueue struct {
	mu       sync.Mutex
	messages []QueuedMessage
	bytes    int
	maxCount int
	maxBytes int
}

func NewConnectionQueue(policy Policy) (*ConnectionQueue, error) {
	if !policy.valid() {
		return nil, ErrInvalidPolicy
	}
	return &ConnectionQueue{maxCount: policy.PlayerQueueMessages, maxBytes: policy.PlayerQueueBytes}, nil
}

func (queue *ConnectionQueue) Enqueue(message QueuedMessage) error {
	if message.Channel == "" || !knownKind(message.Kind) || len(message.Data) == 0 {
		return ErrInvalidPolicy
	}
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if message.Channel == "world" {
		for index := len(queue.messages) - 1; index >= 0; index-- {
			if queue.messages[index].Channel == "world" {
				queue.bytes -= len(queue.messages[index].Data)
				queue.messages[index] = message
				queue.bytes += len(message.Data)
				return nil
			}
		}
	}
	if len(queue.messages)+1 > queue.maxCount || queue.bytes+len(message.Data) > queue.maxBytes {
		return ErrQueueOverflow
	}
	queue.messages = append(queue.messages, QueuedMessage{Channel: message.Channel, Kind: message.Kind, Data: append([]byte(nil), message.Data...)})
	queue.bytes += len(message.Data)
	return nil
}

func (queue *ConnectionQueue) Drain() []QueuedMessage {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	result := append([]QueuedMessage(nil), queue.messages...)
	queue.messages, queue.bytes = nil, 0
	return result
}
