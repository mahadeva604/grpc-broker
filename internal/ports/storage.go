package ports

type Storage interface {
	GreateTopic(topicName string, queueLength int) (Topic, error)
	GetTopic(topicName string) (Topic, error)
}
