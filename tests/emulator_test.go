package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strconv"
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

func TestSinglePublisherSingleSubscriber(t *testing.T) {
	var publishResults []*PublishResult
	var subscription string
	var topic string

	// Setup a new topic and subscription and delete them when done
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()
	newTopic, err := createRandomTopic(ctx)
	if newTopic != nil && err == nil {
		topic = newTopic.ID()
		defer newTopic.Delete(ctx)
	} else {
		log.Fatalf("Failed to init tests: %v", err)
	}
	newSubscription, err := createRandomSubscription(ctx, newTopic.ID())
	if newSubscription != nil && err == nil {
		subscription = newSubscription.ID()
		defer newSubscription.Delete(ctx)
	} else {
		log.Fatalf("Failed to init tests: %v", err)
	}

	Convey("Given a Publisher", t, func() {
		ctx := context.Background()
		publisher, err := NewPublisher(ctx, projectID, "pub1")
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
		subscriber, err := NewSubscriber(ctx, projectID, "sub1")
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

func createRandomTopic(ctx context.Context) (*pubsub.Topic, error) {
	topic := "randomtopic-" + strconv.Itoa(rand.Int()%10000)
	log.Println("New Topic: " + topic)
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client.CreateTopic(ctx, topic)
}

func createRandomSubscription(ctx context.Context, topic string) (*pubsub.Subscription, error) {
	subscription := "subscription-" + strconv.Itoa(rand.Int()%10000)
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
