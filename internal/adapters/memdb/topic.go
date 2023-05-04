package memdb

import (
	"errors"
	"sync"

	"github.com/mahadeva604/grpc-broker/internal/domain/models"
)

// Topic object
type Topic struct {
	mu    sync.Mutex
	queue *messageQueue
}

// NewTopic create new topic object
func NewTopic(queueLength int) *Topic {
	return &Topic{
		mu:    sync.Mutex{},
		queue: newMessageQueue(queueLength),
	}
}

// Publish publish message to topic
func (t *Topic) Publish(msg models.Message) {
	t.queue.enqueue(msg)
}

// TrySubscribe subscribe to topic, return err if topic busy
func (t *Topic) TrySubscribe(ack <-chan struct{}, done <-chan struct{}) (<-chan *models.Message, error) {
	if !t.mu.TryLock() {
		return nil, errors.New("topic busy")
	}

	return t.subscribe(ack, done), nil
}

// Subscribe subscribe to topic, block if topic busy
func (t *Topic) Subscribe(ack <-chan struct{}, done <-chan struct{}) <-chan *models.Message {
	t.mu.Lock()

	return t.subscribe(ack, done)
}

func (t *Topic) subscribe(ack <-chan struct{}, done <-chan struct{}) <-chan *models.Message {
	messages := make(chan *models.Message)
	waitStop := make(chan struct{})

	go func() {
		<-done
		t.queue.stopDequeue()
		close(waitStop)
	}()

	go func() {
		defer t.mu.Unlock()
		defer close(messages)
		defer t.queue.resetStopDequeue()

	L:
		for {
			msgWitchAck := t.queue.dequeue()
			if msgWitchAck.message == nil {
				break L
			}

			select {
			case messages <- msgWitchAck.message:
			case <-done:
				break L
			}

			select {
			case <-ack:
				msgWitchAck.ackFunc()
			case <-done:
				break L
			}
		}

		<-waitStop
	}()

	return messages
}
