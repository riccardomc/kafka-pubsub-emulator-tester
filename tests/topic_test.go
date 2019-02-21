package tests

import (
	"context"
	"log"
	"testing"

	"cloud.google.com/go/pubsub"

	. "github.com/smartystreets/goconvey/convey" // noqa
)

func TestTopic(t *testing.T) {
	Convey("Given a Pub/Sub client", t, func() {
		ctx := context.Background()
		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		client.CreateTopic(ctx, "topic1")
		Convey("When the client creates an new topic", func() {
			//Attempting to delete the topic before testing
			topicCheck := client.Topic("topic-new")
			err := topicCheck.Delete(ctx)
			if err != nil {
				log.Printf("Delete topic before test failed: %v", err)
			}
			topic, err := client.CreateTopic(ctx, "topic-new")
			Convey("Then the topic is created with no errors", func() {
				So(err, ShouldBeNil)
				So(topic, ShouldNotBeNil)
				exists, err := topic.Exists(ctx)
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
				Convey("And deleting succeeds with no errors", func() {
					err = topic.Delete(ctx)
					So(err, ShouldBeNil)
					exists, err := topic.Exists(ctx)
					So(err, ShouldBeNil)
					So(exists, ShouldBeFalse)
				})
			})
		})
	})
}
