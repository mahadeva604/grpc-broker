package ports

// Storage representing storage interface
type Storage interface {
	GreateTopic(topicName string, queueLength int) (Topic, error)
	GetTopic(topicName string) (Topic, error)
}
