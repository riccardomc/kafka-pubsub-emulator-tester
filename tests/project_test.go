package tests

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/riccardomc/kafka-pubsub-emulator-tester/utils"
	. "github.com/smartystreets/goconvey/convey" // noqa
)

func TestProject(t *testing.T) {

	Convey("Given two Pub/Sub client to two different Projects", t, func() {
		rand.Seed(time.Now().UnixNano())
		ctx1 := context.Background()
		ctx2 := context.Background()
		projectID1 := createRandomProjectID()
		projectID2 := createRandomProjectID()
		client1, err1 := pubsub.NewClient(ctx1, projectID1)
		client2, err2 := pubsub.NewClient(ctx2, projectID2)
		So(err1, ShouldBeNil)
		So(err2, ShouldBeNil)

		Convey("When creating two topics with the same name in both projects", func() {
			topicName := "topic-both"

			// Create topic in projectID1 and defer deletion
			topicCheck := client1.Topic(topicName)
			topicCheck.Delete(ctx1)
			topic1, err := client1.CreateTopic(ctx1, topicName)
			So(err, ShouldBeNil)

			// Create topic in projectID2 and defer deletion
			topicCheck = client2.Topic(topicName)
			topicCheck.Delete(ctx2)
			topic2, err := client2.CreateTopic(ctx2, topicName)
			So(err, ShouldBeNil)

			Convey("When publishing to topic in ProjectID1", func() {
				publisher1, err := utils.NewPublisher(ctx1, projectID1, projectID1+"-publisher")
				So(err, ShouldBeNil)
				published := publisher1.Publish(ctx1, topicName, 1)
				So(published, ShouldNotBeEmpty)
				for _, r := range published {
					So(r.Err, ShouldBeNil)
				}
				Convey("Then receiving from ProjectID2 times out", func() {
					subscription, err := createRandomSubscription(ctx2, topic2.ID(), projectID2)
					So(err, ShouldBeNil)
					defer subscription.Delete(ctx2)
					subscriber, err := utils.NewSubscriber(ctx2, projectID2, projectID2+"-subscriber")
					So(err, ShouldBeNil)
					received, err := subscriber.Receive(ctx2, topic2.ID(), subscription.ID(), 10*time.Second)
					So(received, ShouldBeEmpty)
					Convey("And receiving from ProjectID1 succeeds", func() {
						subscription, err := createRandomSubscription(ctx1, topic1.ID(), projectID1)
						So(err, ShouldBeNil)
						defer subscription.Delete(ctx1)
						subscriber, err := utils.NewSubscriber(ctx1, projectID1, projectID1+"-subscriber")
						So(err, ShouldBeNil)
						received, err := subscriber.Receive(ctx1, topic1.ID(), subscription.ID(), 10*time.Second)
						So(err, ShouldBeNil)
						So(received, ShouldNotBeEmpty)
						So(len(received), ShouldEqual, 1)
					})
				})
			})
		})
	})
}
