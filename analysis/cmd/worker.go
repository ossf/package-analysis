package main

import (
	"context"
	"log"
	"os"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"

	"github.com/ossf/package-analysis/analysis"
)

func messageLoop(ctx context.Context, sub *pubsub.Subscription, resultsBucket, docstorePath string) {
	for {
		msg, err := sub.Receive(ctx)
		if err != nil {
			// All subsequent receive calls will return the same error, so we bail out.
			log.Printf("error receiving message: %v", err)
			return
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
				log.Panicf("Failed to upload to blobstore = %v\n", err)
			}
		}

		if docstorePath != "" {
			err = analysis.WriteResultsToDocstore(ctx, docstorePath, result)
			if err != nil {
				log.Panicf("Failed to write to docstore = %v\n", err)
			}
		}

		msg.Ack()
	}
}

func main() {
	ctx := context.Background()
	subURL := os.Getenv("OSSMALWARE_WORKER_SUBSCRIPTION")
	resultsBucket := os.Getenv("OSSF_MALWARE_ANALYSIS_RESULTS")
	docstorePath := os.Getenv("OSSMALWARE_DOCSTORE_URL")

	for {
		sub, err := pubsub.OpenSubscription(ctx, subURL)
		if err != nil {
			log.Panic(err)
		}

		messageLoop(ctx, sub, resultsBucket, docstorePath)
	}
}
