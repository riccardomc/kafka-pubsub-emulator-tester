package tests

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strconv"

	"cloud.google.com/go/pubsub"
	. "github.com/smartystreets/goconvey/convey" // noqa
)

func createRandomTopic(ctx context.Context, projectID string) (*pubsub.Topic, error) {
	topic := "randomtopic-" + strconv.Itoa(rand.Int())
	log.Println("New Topic: " + topic)
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client.CreateTopic(ctx, topic)
}

func createRandomSubscription(ctx context.Context, topic string, projectID string) (*pubsub.Subscription, error) {
	subscription := "subscription-" + strconv.Itoa(rand.Int())
	log.Println("New Subscription: " + subscription)
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	config := pubsub.SubscriptionConfig{
		Topic: client.Topic(topic),
	}
	return client.CreateSubscription(ctx, subscription, config)
}

func shouldEqualPubsubMessage(actual interface{}, expected ...interface{}) string {
	a := actual.(*pubsub.Message)
	e := expected[0].(*pubsub.Message)
	if str := ShouldResemble(a.Data, e.Data); str != "" {
		return str
	}
	if str := ShouldResemble(a.Attributes, e.Attributes); str != "" {
		return str
	}
	return ""
}

func shouldContainPubSubMessage(actual interface{}, expected ...interface{}) string {
	messages := actual.([]*pubsub.Message)
	message := expected[0].(*pubsub.Message)

	for _, m := range messages {
		if reflect.DeepEqual(message.Attributes, m.Attributes) &&
			reflect.DeepEqual(message.Data, m.Data) {
			return ""
		}
	}

	return fmt.Sprintf("Message not found: %v", messages)
}
