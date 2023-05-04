package broker

import (
	"context"
	"errors"

	"github.com/mahadeva604/grpc-broker/internal/domain/derrors"
	"github.com/mahadeva604/grpc-broker/internal/domain/models"
	"github.com/mahadeva604/grpc-broker/internal/ports"
	pb "github.com/mahadeva604/grpc-broker/pkg/api/broker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Publisher object
type Publisher struct {
	store ports.Storage
	pb.UnimplementedPublisherServer
}

// NewPublisher create new publisher
func NewPublisher(store ports.Storage) *Publisher {
	return &Publisher{
		store: store,
	}
}

// Publish publish message to topic
func (p *Publisher) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	topicName := req.GetTopic()
	if topicName == "" {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "topic must be not empty")
	}

	topic, err := p.store.GetTopic(topicName)
	if err != nil && !errors.Is(err, derrors.ErrTopicNotFound) {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, err.Error())
	}

	if errors.Is(err, derrors.ErrTopicNotFound) {
		var err error

		topic, err = p.store.GreateTopic(topicName, 1000)
		if err != nil {
			return &emptypb.Empty{}, status.Error(codes.Internal, "can't create topic")
		}
	}

	message := models.Message{
		Key:  req.GetMsg().Key,
		Data: req.GetMsg().Data,
	}

	topic.Publish(message)

	return &emptypb.Empty{}, nil
}
