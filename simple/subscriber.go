package main

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
)

func main() {
	project := "riccardomc-playground"
	subName := "subscription-topic1"
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		log.Fatalf("Client error: %v", err)
	}
	sub := client.Subscription(subName)
	config, err := sub.Config(ctx)
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}
	log.Println(config)
}
