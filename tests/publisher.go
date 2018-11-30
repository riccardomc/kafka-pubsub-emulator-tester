package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
)

//Publisher is a Pub/Sub publisher
type Publisher struct {
	Name   string
	Client *pubsub.Client
}

//NewPublisher gives you a fresh Publisher
func NewPublisher(ctx context.Context, project, name string) (*Publisher, error) {
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}
	return &Publisher{name, client}, nil
}

//Publish publishes n messages to the topic and returns the number of errors
func (p *Publisher) Publish(ctx context.Context, topic string, n int) int64 {
	var wg sync.WaitGroup
	var totalErrors int64
	t := p.Client.Topic(topic)

	for i := 0; i < n; i++ {
		result := t.Publish(ctx, &pubsub.Message{
			Data: []byte("Message-" + p.Name + "-" + strconv.Itoa(i)),
		})

		wg.Add(1)
		go func(i int, res *pubsub.PublishResult) {
			defer wg.Done()
			// The Get method blocks until a server-generated ID or
			// an error is returned for the published message.
			id, err := res.Get(ctx)
			if err != nil {
				// Error handling code can be added here.
				log.Output(1, fmt.Sprintf("Failed to publish: %v", err))
				atomic.AddInt64(&totalErrors, 1)
				return
			}
			log.Printf("Published message %d; msg ID: %v\n", i, id)
		}(i, result)
	}

	wg.Wait()

	return totalErrors
}
