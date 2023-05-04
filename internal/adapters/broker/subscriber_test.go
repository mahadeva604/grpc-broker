package broker

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mahadeva604/grpc-broker/internal/domain/models"
	"github.com/mahadeva604/grpc-broker/internal/ports"
	mock_ports "github.com/mahadeva604/grpc-broker/internal/ports/mock"
	pb "github.com/mahadeva604/grpc-broker/pkg/api/broker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestSubscriber(t *testing.T) {
	ctx := context.Background()

	messages := []*models.Message{
		{
			Key:  "key 1",
			Data: []byte("data 1"),
		},
		{
			Key:  "key 2",
			Data: []byte("data 2"),
		},
		{
			Key:  "key 3",
			Data: []byte("data 3"),
		},
	}

	messagesChan := make(chan *models.Message, len(messages))

	for _, message := range messages {
		messagesChan <- message
	}

	close(messagesChan)

	ctrl := gomock.NewController(t)
	mockTopic := mock_ports.NewMockTopic(ctrl)
	mockTopic.EXPECT().
		TrySubscribe(gomock.Any(), gomock.Any()).
		Do(
			func(ack <-chan struct{}, done <-chan struct{}) {
				go func() {
					for range ack {
					}
				}()
			},
		).Return(messagesChan, nil)

	mockStrorage := mock_ports.NewMockStorage(ctrl)
	mockStrorage.EXPECT().GetTopic("test_topic").Return(mockTopic, nil)

	client, closer := subscriberServer(ctx, mockStrorage)
	defer closer()

	stream, err := client.Subscribe(ctx)
	require.NoError(t, err)

	stream.Send(&pb.SubscribeRequest{
		SubscribeRequestType: &pb.SubscribeRequest_InitialRequest{
			InitialRequest: &pb.InitialRequest{
				Topic: "test_topic",
			},
		},
	})

	for _, expectMessage := range messages {
		gotMessage, err := stream.Recv()
		require.NoError(t, err)
		require.Equal(t, expectMessage.Key, gotMessage.Key)
		require.Equal(t, expectMessage.Data, gotMessage.Data)

		stream.Send(&pb.SubscribeRequest{
			SubscribeRequestType: &pb.SubscribeRequest_AckReply{
				AckReply: &pb.AckReply{
					Key: gotMessage.Key,
				},
			},
		})
	}
}

func subscriberServer(ctx context.Context, storage ports.Storage) (pb.SubscriberClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterSubscriberServer(baseServer, NewSubscriber(storage))

	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := pb.NewSubscriberClient(conn)

	return client, closer
}
