package tests

import (
	"context"
	"log"
	"testing"

	"cloud.google.com/go/pubsub"
	. "github.com/smartystreets/goconvey/convey" // noqa
)

func TestSubscription(t *testing.T) {
	Convey("Given a Pub/Sub client", t, func() {
		ctx := context.Background()
		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		client.CreateTopic(ctx, "topic1")

		Convey("When the client subscribes to a not existent topic", func() {
			subscriptionConfig := pubsub.SubscriptionConfig{
				Topic: client.Topic("idontexist"),
			}
			_, err := client.CreateSubscription(ctx, "subscription-idontexist", subscriptionConfig)

			Convey("Then subscription fails with 'Topic not found'", func() {
				So(err.Error(), ShouldContainSubstring, "NotFound")

			})
		})
	})

	Convey("Given a Pub/Sub client", t, func() {
		ctx := context.Background()
		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		client.CreateTopic(ctx, "topic1")
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
