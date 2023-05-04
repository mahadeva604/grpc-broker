package memdb

import (
	"sync"
	"sync/atomic"

	"github.com/mahadeva604/grpc-broker/internal/domain/models"
)

type messageWithAck struct {
	message *models.Message
	ackFunc func()
}

type messageQueue struct {
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

func newMessageQueue(buffLength int) *messageQueue {
	mu := &sync.Mutex{}

	msgQueue := messageQueue{
		mu:         mu,
		ringBuffer: make([]*models.Message, buffLength),
		buffLength: buffLength,
		notEmpty:   sync.NewCond(mu),
		notFull:    sync.NewCond(mu),
	}

	return &msgQueue
}

func (q *messageQueue) enqueue(msg models.Message) {
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

func (q *messageQueue) dequeue() messageWithAck {
	q.mu.Lock()
	defer q.mu.Unlock()

	for q.currLength == 0 && !q.isStopDequeue.Load() {
		q.notEmpty.Wait()
	}

	if q.isStopDequeue.Load() {
		return messageWithAck{}
	}

	return messageWithAck{q.ringBuffer[q.nextRead], q.ack}
}

func (q *messageQueue) stopDequeue() {
	q.isStopDequeue.Store(true)
	q.notEmpty.Signal()
}

func (q *messageQueue) resetStopDequeue() {
	q.isStopDequeue.Store(false)
}

func (q *messageQueue) ack() {
	q.mu.Lock()

	q.nextRead++
	if q.nextRead == q.buffLength {
		q.nextRead = 0
	}
	q.currLength--

	q.notFull.Signal()

	q.mu.Unlock()
}
