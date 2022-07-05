package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"time"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"

	"github.com/ossf/package-analysis/cmd/scheduler/proxy"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-feeds/pkg/feeds"
)

const (
	maxRetries    = 10
	retryInterval = 1
	retryExpRate  = 1.5
)

// supportedPkgManagers lists the package managers Package Analysis can
// analyze. Each entry can contain zero or more regexp filters that exclude
// unwanted versions to minimize noise.
var supportedPkgManagers = map[string][]*regexp.Regexp{
	"npm":       {},
	"pypi":      {},
	"rubygems":  {},
	"packagist": {regexp.MustCompile(`^dev-`), regexp.MustCompile(`\.x-dev$`)},
}

func main() {
	retryCount := 0
	subscriptionURL := os.Getenv("OSSMALWARE_SUBSCRIPTION_URL")
	topicURL := os.Getenv("OSSMALWARE_WORKER_TOPIC")
	log.Initalize(os.Getenv("LOGGER_ENV"))

	for retryCount <= maxRetries {
		err := listenLoop(subscriptionURL, topicURL)

		if err != nil {
			if retryCount++; retryCount >= maxRetries {
				log.Error("Retries exceeded",
					"error", err,
					"retryCount", retryCount)
				break
			}

			retryDuration := time.Second * time.Duration(retryDelay(retryCount))
			log.Error("Error encountered, retrying",
				"error", err,
				"retryCount", retryCount,
				"waitSeconds", retryDuration.Seconds())
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
	log.Info("Listening for messages to proxy...")

	err = srv.Listen(ctx, func(m *pubsub.Message) (*pubsub.Message, error) {
		log.Info("Handling message",
			"body", string(m.Body))
		pkg := feeds.Package{}
		if err := json.Unmarshal(m.Body, &pkg); err != nil {
			return nil, fmt.Errorf("error unmarshalling json: %w", err)
		}
		if filters, ok := supportedPkgManagers[pkg.Type]; !ok {
			return nil, fmt.Errorf("package type is not supported: %v", pkg.Type)
		} else {
			for _, filter := range filters {
				if filter.MatchString(pkg.Version) {
					return nil, fmt.Errorf("package version '%v' is filtered for type: %v", pkg.Version, pkg.Type)
				}
			}
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
