package utils

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
)

//Subscriber is a Pub/Sub subscriber
type Subscriber struct {
	Name   string
	Client *pubsub.Client
}

//NewSubscriber gives you a fresh Subscriber
func NewSubscriber(ctx context.Context, project, name string) (*Subscriber, error) {
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}
	return &Subscriber{name, client}, nil
}

//GetOrCreateSubscription returns an existing Subscription or creates a new one
func (s *Subscriber) GetOrCreateSubscription(ctx context.Context, topic, subscription string) (*pubsub.Subscription, error) {
	sub := s.Client.Subscription(subscription)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		config := pubsub.SubscriptionConfig{
			Topic: s.Client.Topic(topic),
		}
		sub, err = s.Client.CreateSubscription(ctx, subscription, config)
		if err != nil {
			return nil, err
		}
	}
	return sub, err
}

//Receive all messages and return after timeout if there aren't any
// See these for implementation details:
// https://cloud.google.com/pubsub/docs/pull
// https://github.com/GoogleCloudPlatform/google-cloud-go/issues/881#issuecomment-362703005
func (s *Subscriber) Receive(ctx context.Context, topic, subscription string, timeout time.Duration) ([]*pubsub.Message, error) {

	sub, err := s.GetOrCreateSubscription(ctx, topic, subscription)

	if err != nil {
		return nil, err
	}
	messages := make([]*pubsub.Message, 0)
	var mutex sync.Mutex

	seen := make(chan int, 1)
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		for {
			select {
			case <-seen:
			case <-time.After(timeout):
				cancel()
				return
			}
		}
	}()

	err = sub.Receive(cctx, func(ctx1 context.Context, message *pubsub.Message) {
		select {
		case seen <- 1:
		default:
		}
		message.Ack()
		mutex.Lock()
		defer mutex.Unlock()
		messages = append(messages, message)
	})
	return messages, err
}
