package main

import (
	"context"
	"flag"
	"time"

	"github.com/mahadeva604/grpc-broker/pkg/broker"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func main() {
	var (
		server  string
		topic   string
		key     string
		message string
	)

	flag.StringVar(&server, "s", "", "server (localhost:8086)")
	flag.StringVar(&topic, "t", "", "topic name")
	flag.StringVar(&key, "k", "", "key")
	flag.StringVar(&message, "m", "", "message")
	flag.Parse()

	if topic == "" || key == "" || server == "" {
		flag.PrintDefaults()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prod := broker.NewProducer(server, grpc.WithInsecure(), grpc.WithBlock())

	prodOpts := broker.ProducerOpts{
		Ctx:   ctx,
		Topic: topic,
		Msg: broker.Message{
			Key:  key,
			Data: []byte(message),
		},
	}

	if err := prod.Publish(prodOpts); err != nil {
		log.Error().Err(err).Msg("exit")
	}
}
