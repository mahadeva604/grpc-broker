package memdb

import (
	"fmt"
	"sync"

	"github.com/mahadeva604/grpc-broker/internal/domain/derrors"
	"github.com/mahadeva604/grpc-broker/internal/ports"
)

type Storage struct {
	mu     sync.Mutex
	topics map[string]*Topic
}

func NewStorage() *Storage {
	return &Storage{
		mu:     sync.Mutex{},
		topics: map[string]*Topic{},
	}
}

func (s *Storage) GreateTopic(topicName string, queueLength int) (ports.Topic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.topics[topicName]; ok {
		return nil, fmt.Errorf("topic %s exists", topicName)
	}

	newTopic := NewTopic(queueLength)

	s.topics[topicName] = newTopic

	return newTopic, nil
}

func (s *Storage) GetTopic(topicName string) (ports.Topic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic, ok := s.topics[topicName]
	if !ok {
		return nil, derrors.ErrTopicNotFound
	}

	return topic, nil
}
