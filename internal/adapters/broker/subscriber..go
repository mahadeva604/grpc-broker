package broker

import (
	"errors"

	"github.com/mahadeva604/grpc-broker/internal/domain/derrors"
	"github.com/mahadeva604/grpc-broker/internal/ports"
	pb "github.com/mahadeva604/grpc-broker/pkg/api/broker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Subscriber struct {
	store ports.Storage
	pb.UnimplementedSubscriberServer
}

func NewSubscriber(store ports.Storage) *Subscriber {
	return &Subscriber{
		store: store,
	}
}

func (s Subscriber) Subscribe(stream pb.Subscriber_SubscribeServer) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	initReq := req.GetInitialRequest()
	if initReq == nil {
		return status.Error(codes.InvalidArgument, "init request not found")
	}

	topicName := initReq.Topic
	if topicName == "" {
		return status.Error(codes.InvalidArgument, "topic must be not empty")
	}

	topic, err := s.store.GetTopic(topicName)
	if err != nil && !errors.Is(err, derrors.ErrTopicNotFound) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if errors.Is(err, derrors.ErrTopicNotFound) {
		var err error

		topic, err = s.store.GreateTopic(topicName, 1000)
		if err != nil {
			return status.Error(codes.Internal, "can't create topic")
		}
	}

	ack := make(chan struct{})
	done := stream.Context().Done()
	defer close(ack)

	messages, err := topic.TrySubscribe(ack, done)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	for message := range messages {
		pbMessage := pb.Message{
			Key:  message.Key,
			Data: message.Data,
		}

		if err := stream.Send(&pbMessage); err != nil {
			return err
		}

		ackReply, err := stream.Recv()
		if err != nil {
			return err
		}

		if ackReply.GetAckReply() == nil {
			return status.Error(codes.InvalidArgument, "unexpected reply")
		}

		select {
		case ack <- struct{}{}:
		case <-done:
			return stream.Context().Err()
		}
	}

	return nil
}
