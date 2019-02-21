package tests

import (
	"context"
	"log"
	"math/rand"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/riccardomc/kafka-pubsub-emulator-tester/utils"

	. "github.com/smartystreets/goconvey/convey" // noqa
)

func TestSinglePublisherSingleSubscriber(t *testing.T) {
	var publishResults []*utils.PublishResult
	var subscription string
	var topic string

	// Setup a new topic and subscription and delete them when done
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()
	newTopic, err := createRandomTopic(ctx, projectID)
	if newTopic != nil && err == nil {
		topic = newTopic.ID()
		defer newTopic.Delete(ctx)
	} else {
		log.Fatalf("Failed to init tests: %v", err)
	}
	newSubscription, err := createRandomSubscription(ctx, newTopic.ID(), projectID)
	if newSubscription != nil && err == nil {
		subscription = newSubscription.ID()
		defer newSubscription.Delete(ctx)
	} else {
		log.Fatalf("Failed to init tests: %v", err)
	}

	Convey("Given a Publisher", t, func() {
		ctx := context.Background()
		publisher, err := utils.NewPublisher(ctx, projectID, "pub1")
		if err != nil {
			t.Fatalf("Unable to create client: %v", err)
		}
		Convey("When the publisher publishes to a new topic", func() {
			publishResults = publisher.Publish(ctx, topic, 10)
			Convey("Messages are published without errors", func() {
				So(publishResults, ShouldNotBeEmpty)
				for _, r := range publishResults {
					So(r.Err, ShouldBeNil)
				}
			})
		})
	})

	Convey("Given a Subscriber", t, func() {
		ctx := context.Background()
		subscriber, err := utils.NewSubscriber(ctx, projectID, "sub1")
		if err != nil {
			t.Fatalf("Unable to create client: %v", err)
		}
		Convey("When the subscriber receives on a new subscription", func() {
			receivedMessages, err := subscriber.Receive(ctx, topic, subscription, 10*time.Second)
			Convey("Then it receives the same amount of published messages without errors", func() {
				So(err, ShouldBeNil)
				So(len(receivedMessages), ShouldEqual, len(publishResults))
				Convey("And the messages received are the same as the published ones", func() {
					publishedMessages := []*pubsub.Message{}
					for _, p := range publishResults {
						publishedMessages = append(publishedMessages, p.Message)
					}
					for _, message := range receivedMessages {
						So(publishedMessages, shouldContainPubSubMessage, message)
					}
				})
			})
		})
	})
}
