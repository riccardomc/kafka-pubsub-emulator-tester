package main

import (
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	. "github.com/smartystreets/goconvey/convey" // noqa
)

var (
	projectID    = "riccardomc-playground" //FIXME: this should be configurable
	topic        = "topic1"
	subscription = "subscription-topic1"
)

func TestTopic(t *testing.T) {

}

func TestSubscription(t *testing.T) {
	Convey("Given a Pub/Sub client", t, func() {
		ctx := context.Background()
		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		Convey("When the client subscribes to a not existent topic", func() {
			subscriptionConfig := pubsub.SubscriptionConfig{
				Topic: client.Topic("idontexist"),
			}
			_, err := client.CreateSubscription(ctx, "subscription-idontexist", subscriptionConfig)

			Convey("Then subscription fails with 'Topic not found'", func() {
				So(err.Error(), ShouldContainSubstring, "not found")
			})
		})
	})

	Convey("Given a Pub/Sub client", t, func() {
		ctx := context.Background()
		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		Convey("When the client creates a subscription to an existing topic", func() {
			//Attempting to delete the subscription before testing
			subCheck := client.Subscription("subscription-new")
			err := subCheck.Delete(ctx)
			if err != nil {
				log.Printf("Delete subscription before test failed: %v", err)
			}

			subscriptionConfig := pubsub.SubscriptionConfig{
				Topic: client.Topic(topic),
			}
			sub, err := client.CreateSubscription(ctx, "subscription-new", subscriptionConfig)
			So(err, ShouldBeNil)
			So(sub, ShouldNotBeNil)

			Convey("Then subscription is created with no errors", func() {
				exists, err := sub.Exists(ctx)
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
				Convey("And deleting succeeds with no errors", func() {
					err = sub.Delete(ctx)
					So(err, ShouldBeNil)
					exists, err := sub.Exists(ctx)
					So(err, ShouldBeNil)
					So(exists, ShouldBeFalse)
				})
			})

		})
	})
}

func TestPubSub(t *testing.T) {
	Convey("Given a Publisher", t, func() {
		ctx := context.Background()
		publisher, err := NewPublisher(ctx, projectID, "pub1")
		if err != nil {
			t.Fatalf("Unable to create client: %v", err)
		}
		Convey("When the publisher publishes to an existing topic", func() {
			errors := publisher.Publish(ctx, topic, 10)
			Convey("Then no errors", func() {
				So(errors, ShouldEqual, 0)
			})
		})
	})

	Convey("Given a Receiver", t, func() {
		ctx := context.Background()
		subscriber, err := NewSubscriber(ctx, projectID, "sub1")
		if err != nil {
			t.Fatalf("Unable to create client: %v", err)
		}
		Convey("When subscriber receives on an existing subscription", func() {
			messages, err := subscriber.Receive(ctx, topic, subscription, 10*time.Second)
			Convey("Then no errors", func() {
				So(err, ShouldBeNil)
				So(len(messages), ShouldNotEqual, 0)
			})
		})
	})
}
