package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mahadeva604/grpc-broker/pkg/broker"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// ProcessMsg handler for incoming messages
func ProcessMsg(messages <-chan broker.Message, ack chan<- struct{}, done <-chan struct{}) {
	defer close(ack)

	for msg := range messages {
		fmt.Printf("Key: %s, Data: %s\n", msg.Key, msg.Data)

		select {
		case ack <- struct{}{}:
		case <-done:
			return
		}
	}
}

func main() {
	var (
		server string
		topic  string
	)

	flag.StringVar(&server, "s", "", "server (localhost:8086)")
	flag.StringVar(&topic, "t", "", "topic name")
	flag.Parse()

	if topic == "" || server == "" {
		flag.PrintDefaults()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	cons := broker.NewSubscriber(server, grpc.WithInsecure(), grpc.WithBlock())

	subscOpts := broker.SubscribeOpts{
		Ctx:     ctx,
		Timeout: 5 * time.Second,
		Topic:   topic,
		Handler: ProcessMsg,
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := cons.Subscribe(subscOpts); err != nil {
			log.Error().Err(err).Msg("exit")
			cancel()
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		cancel()
	case <-ctx.Done():
	}

	wg.Wait()
}
