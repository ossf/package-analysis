package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"

	"github.com/ossf/package-analysis/cmd/scheduler/proxy"
	"github.com/ossf/package-feeds/pkg/feeds"
)

const (
	maxRetries    = 10
	retryInterval = 1
	retryExpRate  = 1.5
)

var supportedPkgManagers = map[string]bool{
	"npm":      true,
	"pypi":     true,
	"rubygems": true,
}

func main() {
	retryCount := 0
	subscriptionURL := os.Getenv("OSSMALWARE_SUBSCRIPTION_URL")
	topicURL := os.Getenv("OSSMALWARE_WORKER_TOPIC")

	for retryCount <= maxRetries {
		err := listenLoop(subscriptionURL, topicURL)

		if err != nil {
			if retryCount++; retryCount >= maxRetries {
				log.Printf("Retries exceeded, Error: %v\n", err)
				break
			}
			log.Printf("Warning: %v\n", err)

			retryDuration := time.Second * time.Duration(retryDelay(retryCount))
			log.Printf("Error encountered, will try again after %v seconds\n", retryDuration.Seconds())
			time.Sleep(retryDuration)
		}
	}
}

func listenLoop(subUrl, topicURL string) error {
	ctx := context.Background()

	sub, err := pubsub.OpenSubscription(ctx, subUrl)
	if err != nil {
		return err
	}

	topic, err := pubsub.OpenTopic(ctx, topicURL)
	if err != nil {
		return err
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

	return err
}

func retryDelay(retryCount int) int {
	return int(math.Floor(retryInterval * math.Pow(retryExpRate, float64(retryCount))))
}
