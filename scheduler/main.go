package main

import (
	"context"
	"log"
	"os"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

func main() {

	subscriptionUrl := os.Getenv("OSSMALWARE_SUBSCRIPTION_URL")
	ctx := context.Background()
	sub, err := pubsub.OpenSubscription(ctx, subscriptionUrl)
	if err != nil {
		panic(err)
	}

	for {
		msg, err := sub.Receive(ctx)
		if err != nil {
			log.Println("error receiving message: ", err)
			continue
		}
		go func(m *pubsub.Message) {
			defer m.Ack()
			log.Println("handling message: ", string(m.Body))
		}(msg)
	}
}
