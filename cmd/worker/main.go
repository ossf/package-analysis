package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"

	"github.com/ossf/package-analysis/internal/analysis"
	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/pkgecosystem"
	"github.com/ossf/package-analysis/internal/sandbox"
)

const (
	maxRetries    = 10
	retryInterval = 1
	retryExpRate  = 1.5
)

func messageLoop(ctx context.Context, subURL, resultsBucket string) error {
	sub, err := pubsub.OpenSubscription(ctx, subURL)
	if err != nil {
		return err
	}

	log.Info("Listening for messages to process...")
	for {
		msg, err := sub.Receive(ctx)
		if err != nil {
			// All subsequent receive calls will return the same error, so we bail out.
			return fmt.Errorf("error receiving message: %w", err)
		}

		name := msg.Metadata["name"]
		if name == "" {
			log.Warn("name is empty")
			msg.Ack()
			continue
		}

		ecosystem := msg.Metadata["ecosystem"]
		if ecosystem == "" {
			log.Warn("ecosystem is empty",
				"name", name)
			msg.Ack()
			continue
		}

		manager, ok := pkgecosystem.SupportedPkgManagers[ecosystem]
		if !ok {
			log.Warn("Unsupported pkg manager",
				"ecosystem", ecosystem,
				"name", name)
			msg.Ack()
			continue
		}

		version := msg.Metadata["version"]
		if version == "" {
			version = manager.GetLatest(name)
		}

		log.Info("Got request",
			"ecosystem", ecosystem,
			"name", name,
			"version", version)
		sb := sandbox.New(manager.Image)
		result := analysis.RunLive(ecosystem, name, version, sb, manager.CommandFmt(name, version))

		if resultsBucket != "" {
			err = analysis.UploadResults(ctx, resultsBucket, ecosystem+"/"+name, result)
			if err != nil {
				return fmt.Errorf("failed to upload to blobstore = %w", err)
			}
		}

		msg.Ack()
	}
}

func main() {
	retryCount := 0
	ctx := context.Background()
	subURL := os.Getenv("OSSMALWARE_WORKER_SUBSCRIPTION")
	resultsBucket := os.Getenv("OSSF_MALWARE_ANALYSIS_RESULTS")
	log.Initalize(os.Getenv("LOGGER_ENV") == "prod")

	for {
		err := messageLoop(ctx, subURL, resultsBucket)
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

func retryDelay(retryCount int) int {
	return int(math.Floor(retryInterval * math.Pow(retryExpRate, float64(retryCount))))
}
