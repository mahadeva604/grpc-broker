package memdb

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mahadeva604/grpc-broker/internal/domain/models"
	"github.com/stretchr/testify/require"
)

type IndexRange struct {
	start int
	end   int
}

func TestQueue(t *testing.T) {
	totalMessages := 1000
	totalProducer := 4
	queue := newMessageQueue(100)

	messages := genMessages(totalMessages)
	indexRanges := genIndexRange(totalMessages, totalProducer)

	wgEnqueue := &sync.WaitGroup{}

	for i := 0; i < totalProducer; i++ {
		wgEnqueue.Add(1)
		go func(producerID int) {
			defer wgEnqueue.Done()

			indexRange := indexRanges[producerID]

			for _, msg := range messages[indexRange.start:indexRange.end] {
				queue.enqueue(msg)
			}
		}(i)
	}

	wgDequeue := &sync.WaitGroup{}

	var gotMessages []models.Message

	// get half of messages
	wgDequeue.Add(1)
	go func() {
		defer wgDequeue.Done()

		queue.resetStopDequeue()

		for {
			msgWitchAck := queue.dequeue()
			if msgWitchAck.message == nil {
				return
			}

			gotMessages = append(gotMessages, *msgWitchAck.message)
			msgWitchAck.ackFunc()

			// stop after half of expected messages
			if len(gotMessages) == totalMessages/2 {
				queue.stopDequeue()
			}
		}
	}()

	wgDequeue.Wait()

	// resume getting messages
	wgDequeue.Add(1)
	go func() {
		defer wgDequeue.Done()

		queue.resetStopDequeue()

		for {
			msgWitchAck := queue.dequeue()
			if msgWitchAck.message == nil {
				return
			}

			gotMessages = append(gotMessages, *msgWitchAck.message)
			msgWitchAck.ackFunc()

			// after expected messages wait and stop
			if len(gotMessages) == totalMessages {
				go func() {
					time.Sleep(3 * time.Second)
					queue.stopDequeue()
				}()
			}
		}
	}()

	wgEnqueue.Wait()
	wgDequeue.Wait()

	require.ElementsMatch(t, messages, gotMessages)
}

func genMessages(msgLen int) []models.Message {
	messages := make([]models.Message, msgLen)

	for i := 0; i < msgLen; i++ {
		key := fmt.Sprintf("key_%d", i)

		messages[i] = models.Message{
			Key:  key,
			Data: nil,
		}
	}

	return messages
}

func genIndexRange(len int, part int) []IndexRange {
	if part == 0 {
		return nil
	}

	if part == 1 {
		return []IndexRange{{0, len}}
	}

	var indexRange []IndexRange

	step := len / part
	mod := len % part

	for start := 0; start < len; {
		idx := IndexRange{start, start + step}
		if mod > 0 {
			idx.end++
			mod--
		}

		start = idx.end

		indexRange = append(indexRange, idx)
	}

	return indexRange
}
