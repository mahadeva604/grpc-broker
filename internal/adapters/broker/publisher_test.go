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

func TestPublisher(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	mockTopic := mock_ports.NewMockTopic(ctrl)
	mockTopic.EXPECT().Publish(models.Message{Key: "test_key", Data: []byte("test data")})

	mockStrorage := mock_ports.NewMockStorage(ctrl)
	mockStrorage.EXPECT().GetTopic("test_topic").Return(mockTopic, nil)

	client, closer := publisherServer(ctx, mockStrorage)
	defer closer()

	_, err := client.Publish(ctx, &pb.PublishRequest{
		Topic: "test_topic",
		Msg: &pb.Message{
			Key:  "test_key",
			Data: []byte("test data"),
		},
	})

	require.NoError(t, err)
}

func publisherServer(ctx context.Context, storage ports.Storage) (pb.PublisherClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	pb.RegisterPublisherServer(baseServer, NewPublisher(storage))

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

	client := pb.NewPublisherClient(conn)

	return client, closer
}
