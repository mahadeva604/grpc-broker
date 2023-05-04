package memdb

import (
	"sync"
	"testing"

	"github.com/mahadeva604/grpc-broker/internal/domain/models"
	"github.com/mahadeva604/grpc-broker/internal/ports"
	"github.com/stretchr/testify/require"
)

func TestTopic(t *testing.T) {
	totalMessages := 1000 // even number
	totalProducer := 4

	storage := NewStorage()
	_, err := storage.GreateTopic("new_topic", 100)
	require.NoError(t, err)

	topic, err := storage.GetTopic("new_topic")
	require.NoError(t, err)

	messages := genMessages(totalMessages)
	indexRanges := genIndexRange(totalMessages, totalProducer)

	wgProducer := sync.WaitGroup{}

	for i := 0; i < totalProducer; i++ {
		wgProducer.Add(1)
		go func(producerID int) {
			defer wgProducer.Done()

			indexRange := indexRanges[producerID]

			for _, msg := range messages[indexRange.start:indexRange.end] {
				topic.Publish(msg)
			}
		}(i)
	}

	wgSubscriber := &sync.WaitGroup{}

	var gotMessagesTotal []models.Message

	wgSubscriber.Add(1)
	go func() {
		defer wgSubscriber.Done()

		gotMessages, err := runSubscriber(topic, totalMessages/2)
		require.NoError(t, err)

		gotMessagesTotal = append(gotMessagesTotal, gotMessages...)
	}()

	wgSubscriber.Wait()

	wgSubscriber.Add(1)
	go func() {
		defer wgSubscriber.Done()

		gotMessages, err := runSubscriber(topic, totalMessages/2)
		require.NoError(t, err)

		gotMessagesTotal = append(gotMessagesTotal, gotMessages...)
	}()

	wgProducer.Wait()

	wgSubscriber.Wait()

	require.ElementsMatch(t, messages, gotMessagesTotal)
}

func runSubscriber(topic ports.Topic, stopAfterMsg int) ([]models.Message, error) {
	var gotMessages []models.Message

	done := make(chan struct{})
	ack := make(chan struct{})
	defer close(ack)

	messages := topic.Subscribe(ack, done)

	i := 0
	for msg := range messages {
		gotMessages = append(gotMessages, *msg)

		// process msg

		select {
		case ack <- struct{}{}:
		case <-done:
			// remove last message (not ack), to prevent double it test
			gotMessages = gotMessages[0 : len(gotMessages)-1]
		}

		i++
		if i == stopAfterMsg {
			close(done)
		}
	}

	return gotMessages, nil
}
