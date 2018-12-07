package main

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
)

//Publisher is a Pub/Sub publisher
type Publisher struct {
	Name   string
	Client *pubsub.Client
}

//PublishResult represent a sent Message
type PublishResult struct {
	Message  *pubsub.Message
	ServerID string
	Err      error
}

//NewPublisher gives you a fresh Publisher
func NewPublisher(ctx context.Context, project, name string) (*Publisher, error) {
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}
	return &Publisher{name, client}, nil
}

//Publish publishes n messages to the topic and returns a list PublishResult
func (p *Publisher) Publish(ctx context.Context, topic string, n int) []*PublishResult {
	var wg sync.WaitGroup
	var mutex sync.Mutex

	t := p.Client.Topic(topic)
	messages := p.GenerateMessages(topic, n)
	results := []*PublishResult{}

	for _, message := range messages {
		result := t.Publish(ctx, message)

		wg.Add(1)
		go func(message *pubsub.Message, result *pubsub.PublishResult) {
			defer mutex.Unlock()
			defer wg.Done()

			// The Get method blocks until a server-generated ID or
			// an error is returned for the published message.
			id, err := result.Get(ctx)

			mutex.Lock()
			results = append(results, &PublishResult{message, id, err})
		}(message, result)
	}

	wg.Wait()

	return results
}

func (p *Publisher) GenerateMessages(topic string, n int) []*pubsub.Message {
	messages := []*pubsub.Message{}
	now := time.Now()
	for i := 0; i < n; i++ {
		messages = append(messages, &pubsub.Message{
			Attributes: map[string]string{
				"random":    strconv.Itoa(rand.Int()),
				"time":      now.String(),
				"sequence":  strconv.Itoa(i),
				"topic":     topic,
				"publisher": p.Name,
			},
			Data: []byte(p.Name + "-" + topic + "-" + strconv.Itoa(i)),
		})
	}

	return messages
}
