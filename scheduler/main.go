package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"

	"github.com/ossf/package-analysis/scheduler/proxy"
	"github.com/ossf/package-feeds/feeds"
)

var supportedPkgManagers = map[string]bool{
	"npm":      true,
	"pypi":     true,
	"rubygems": true,
}

func main() {
	subscriptionURL := os.Getenv("OSSMALWARE_SUBSCRIPTION_URL")
	ctx := context.Background()
	sub, err := pubsub.OpenSubscription(ctx, subscriptionURL)
	if err != nil {
		panic(err)
	}

	topicURL := os.Getenv("OSSMALWARE_WORKER_TOPIC")
	topic, err := pubsub.OpenTopic(ctx, topicURL)
	if err != nil {
		panic(err)
	}
	srv := proxy.New(topic, sub)
	fmt.Println("Listening for messages to proxy...")

	err = srv.Listen(ctx, func(m *pubsub.Message) (*pubsub.Message, error) {
		log.Println("Handling message: ", string(m.Body))
		pkg := feeds.Package{}
		if err := json.Unmarshal(m.Body, &pkg); err != nil {
			return nil, fmt.Errorf("error unmarshalling json: %w", err)
		}
		if _, ok := supportedPkgManagers[pkg.Type]; !ok {
			return nil, fmt.Errorf("package type is not supported: %v", pkg.Type)
		}
		return &pubsub.Message{
			Body: []byte{},
			Metadata: map[string]string{
				"name":      pkg.Name,
				"ecosystem": pkg.Type,
				"version":   pkg.Version,
			},
		}, nil
	})

	if err != nil {
		panic(err)
	}
}
