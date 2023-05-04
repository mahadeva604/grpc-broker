package ports

import "github.com/mahadeva604/grpc-broker/internal/domain/models"

// Topic representing topic interface
type Topic interface {
	Publish(msg models.Message)
	Subscribe(ack <-chan struct{}, done <-chan struct{}) <-chan *models.Message
	TrySubscribe(ack <-chan struct{}, done <-chan struct{}) (<-chan *models.Message, error)
}
