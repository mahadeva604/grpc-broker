package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mahadeva604/grpc-broker/internal/adapters/broker"
	"github.com/mahadeva604/grpc-broker/internal/adapters/memdb"
	pb "github.com/mahadeva604/grpc-broker/pkg/api/broker"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8086")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	storage := memdb.NewStorage()

	subscriber := broker.NewSubscriber(storage)
	publisher := broker.NewPublisher(storage)

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterSubscriberServer(grpcServer, subscriber)
	pb.RegisterPublisherServer(grpcServer, publisher)

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			fmt.Println()
			//log.Fatal("error occurred while running grpc server", logger.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	grpcServer.Stop()
}
