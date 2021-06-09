package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"

	"github.com/ossf/package-analysis/analysis"
)

const (
	maxRetries    = 10
	retryInterval = 1
	retryExpRate  = 1.5
)

func messageLoop(ctx context.Context, subURL, resultsBucket, docstorePath string) error {
	sub, err := pubsub.OpenSubscription(ctx, subURL)
	if err != nil {
		return err
	}

	fmt.Println("Awaiting message from subscription...")
	for {
		msg, err := sub.Receive(ctx)
		if err != nil {
			// All subsequent receive calls will return the same error, so we bail out.
			return fmt.Errorf("error receiving message: %w", err)
		}

		name := msg.Metadata["name"]
		if name == "" {
			log.Printf("name is empty")
			msg.Ack()
			continue
		}

		ecosystem := msg.Metadata["ecosystem"]
		if ecosystem == "" {
			log.Printf("ecosystem is empty")
			msg.Ack()
			continue
		}

		manager, ok := analysis.SupportedPkgManagers[ecosystem]
		if !ok {
			log.Printf("Unsupported pkg manager %s", manager)
			msg.Ack()
			continue
		}
		log.Printf("Got request %s/%s", ecosystem, name)

		version := msg.Metadata["version"]
		if version == "" {
			version = manager.GetLatest(name)
		}
		log.Printf("Installing version %s", version)

		log.Printf("Got request %s/%s at version %s", ecosystem, name, version)
		result := analysis.Run(ecosystem, name, version, manager.Image, manager.CommandFmt(name, version))

		if resultsBucket != "" {
			err = analysis.UploadResults(ctx, resultsBucket, ecosystem+"/"+name, result)
			if err != nil {
				return fmt.Errorf("failed to upload to blobstore = %w", err)
			}
		}

		if docstorePath != "" {
			err = analysis.WriteResultsToDocstore(ctx, docstorePath, result)
			if err != nil {
				return fmt.Errorf("failed to write to docstore = %v\n", err)
			}
		}

		msg.Ack()
	}
	return nil
}

func main() {
	retryCount := 0
	ctx := context.Background()
	subURL := os.Getenv("OSSMALWARE_WORKER_SUBSCRIPTION")
	resultsBucket := os.Getenv("OSSF_MALWARE_ANALYSIS_RESULTS")
	docstorePath := os.Getenv("OSSMALWARE_DOCSTORE_URL")

	for {
		err := messageLoop(ctx, subURL, resultsBucket, docstorePath)
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

func retryDelay(retryCount int) int {
	return int(math.Floor(retryInterval * math.Pow(retryExpRate, float64(retryCount))))
}
