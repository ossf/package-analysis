package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"

	"github.com/ossf/package-analysis/analysis"
)

func parseS3URL(s3URL string) string {
	// This assumes OSSF_MALWARE_ANALYSIS_RESULTS given in the form s3://endpoint:port/bucket
	// https://gocloud.dev/howto/blob/#s3-compatible

	parsedURL, err := url.Parse(s3URL)
	if err != nil {
		log.Printf("s3 url for OSSF_MALWARE_ANALYSIS_RESULTS could not be parsed: %v", err)
	}

	return parsedURL.Scheme + ":/" + parsedURL.Path + "?endpoint=" + parsedURL.Host + "&disableSSL=true&s3ForcePathStyle=true"
}

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

		err = analysis.UploadResults(ctx, resultsBucket, ecosystem+"/"+name, result)
		if err != nil {
			log.Panicf("Failed to upload to blobstore = %v\n", err)
		}

		err = analysis.WriteResultsToDocstore(ctx, docstorePath, result)
		if err != nil {
			log.Panicf("Failed to write to docstore = %v\n", err)
		}

		msg.Ack()
	}
}

func main() {
	ctx := context.Background()
	subURL := os.Getenv("OSSMALWARE_WORKER_SUBSCRIPTION")
	resultsBucket := os.Getenv("OSSF_MALWARE_ANALYSIS_RESULTS")
	if strings.HasPrefix(resultsBucket, "s3://") {
		// The env must also contain vars for:
		// AWS_SECRET_ACCESS_KEY
		// AWS_ACCESS_KEY_ID
		// AWS_REGION
		// https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
		resultsBucket = parseS3URL(resultsBucket)
	}
	docstorePath := os.Getenv("OSSMALWARE_DOCSTORE_URL")

	for {
		sub, err := pubsub.OpenSubscription(ctx, subURL)
		if err != nil {
			log.Panic(err)
		}

		messageLoop(ctx, sub, resultsBucket, docstorePath)
	}
}
