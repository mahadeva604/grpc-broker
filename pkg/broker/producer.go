package broker

import (
	"context"

	pb "github.com/mahadeva604/grpc-broker/pkg/api/broker"
	"google.golang.org/grpc"
)

type Producer struct {
	target   string
	grpcOpts []grpc.DialOption
}

type ProducerOpts struct {
	Ctx   context.Context
	Topic string
	Msg   Message
}

func NewProducer(taget string, grpcOpts ...grpc.DialOption) *Producer {
	return &Producer{
		target:   taget,
		grpcOpts: grpcOpts,
	}
}

func (p *Producer) Publish(opts ProducerOpts) error {
	cc, err := grpc.DialContext(opts.Ctx, p.target, p.grpcOpts...)
	if err != nil {
		return err
	}

	publisherCLient := pb.NewPublisherClient(cc)

	_, err = publisherCLient.Publish(opts.Ctx, &pb.PublishRequest{
		Topic: opts.Topic,
		Msg: &pb.Message{
			Key:  opts.Msg.Key,
			Data: opts.Msg.Data,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
