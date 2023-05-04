package memdb

import (
	"sync"
	"sync/atomic"

	"github.com/mahadeva604/grpc-broker/internal/domain/models"
)

type MessageWithAck struct {
	message *models.Message
	ackFunc func()
}

type MessageQueue struct {
	mu            *sync.Mutex
	isStopDequeue atomic.Bool
	notFull       *sync.Cond
	notEmpty      *sync.Cond
	nextRead      int
	nextWrite     int
	currLength    int
	buffLength    int
	ringBuffer    []*models.Message
}

func NewMessageQueue(buffLength int) *MessageQueue {
	mu := &sync.Mutex{}

	msgQueue := MessageQueue{
		mu:         mu,
		ringBuffer: make([]*models.Message, buffLength),
		buffLength: buffLength,
		notEmpty:   sync.NewCond(mu),
		notFull:    sync.NewCond(mu),
	}

	return &msgQueue
}

func (q *MessageQueue) Enqueue(msg models.Message) {
	q.mu.Lock()

	for q.currLength == q.buffLength {
		q.notFull.Wait()
	}

	q.ringBuffer[q.nextWrite] = &msg

	q.currLength++
	q.nextWrite++

	if q.nextWrite == q.buffLength {
		q.nextWrite = 0
	}

	q.notEmpty.Signal()

	q.mu.Unlock()
}

func (q *MessageQueue) Dequeue() MessageWithAck {
	q.mu.Lock()
	defer q.mu.Unlock()

	for q.currLength == 0 && !q.isStopDequeue.Load() {
		q.notEmpty.Wait()
	}

	if q.isStopDequeue.Load() {
		return MessageWithAck{}
	}

	return MessageWithAck{q.ringBuffer[q.nextRead], q.ack}
}

func (q *MessageQueue) StopDequeue() {
	q.isStopDequeue.Store(true)
	q.notEmpty.Signal()
}

func (q *MessageQueue) ResetStopDequeue() {
	q.isStopDequeue.Store(false)
}

func (q *MessageQueue) ack() {
	q.mu.Lock()

	q.nextRead++
	if q.nextRead == q.buffLength {
		q.nextRead = 0
	}
	q.currLength--

	q.notFull.Signal()

	q.mu.Unlock()
}
