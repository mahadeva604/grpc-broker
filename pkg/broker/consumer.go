package broker

import (
	"context"
	"time"

	pb "github.com/mahadeva604/grpc-broker/pkg/api/broker"
	"google.golang.org/grpc"
)

type Consumer struct {
	target   string
	grpcOpts []grpc.DialOption
}

type Message struct {
	Key  string
	Data []byte
}

type SubscribeOpts struct {
	Ctx     context.Context
	Timeout time.Duration
	Topic   string
	Handler func(<-chan Message, chan<- struct{}, <-chan struct{})
}

func NewSubscriber(taget string, grpcOpts ...grpc.DialOption) *Consumer {
	return &Consumer{
		target:   taget,
		grpcOpts: grpcOpts,
	}
}

func (p *Consumer) Subscribe(opts SubscribeOpts) error {
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	cc, err := grpc.DialContext(ctx, p.target, p.grpcOpts...)
	if err != nil {
		return err
	}

	subscriberCLient := pb.NewSubscriberClient(cc)

	subscriberStream, err := subscriberCLient.Subscribe(opts.Ctx)
	if err != nil {
		return err
	}

	err = subscriberStream.Send(&pb.SubscribeRequest{
		SubscribeRequestType: &pb.SubscribeRequest_InitialRequest{
			InitialRequest: &pb.InitialRequest{
				Topic: opts.Topic,
			},
		},
	})
	if err != nil {
		return err
	}

	messages := make(chan Message)
	defer close(messages)

	ack := make(chan struct{})

	go opts.Handler(messages, ack, opts.Ctx.Done())

	for {
		message, err := subscriberStream.Recv()
		if err != nil {
			return err
		}

		select {
		case messages <- Message{
			Key:  message.Key,
			Data: message.Data,
		}:
		case <-opts.Ctx.Done():
			return opts.Ctx.Err()
		}

		select {
		case <-ack:
		case <-opts.Ctx.Done():
			return opts.Ctx.Err()
		}

		err = subscriberStream.Send(&pb.SubscribeRequest{
			SubscribeRequestType: &pb.SubscribeRequest_AckReply{
				AckReply: &pb.AckReply{
					Key: message.Key,
				},
			},
		})
		if err != nil {
			return err
		}
	}
}
