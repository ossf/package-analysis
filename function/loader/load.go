package loader

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func Load(ctx context.Context, m PubSubMessage) error {
	project := os.Getenv("GCP_PROJECT")
	bucket := os.Getenv("OSSF_MALWARE_ANALYSIS_RESULTS")

	bq, err := bigquery.NewClient(ctx, project)
	if err != nil {
		log.Panicf("Failed to create bq client: %v", err)
	}

	gcsRef := bigquery.NewGCSReference(fmt.Sprintf("gs://%s/*.json", bucket))
	gcsRef.AutoDetect = true
	gcsRef.SourceFormat = bigquery.JSON

	dataset := bq.Dataset("packages")
	loader := dataset.Table("analysis").LoaderFrom(gcsRef)
	loader.WriteDisposition = bigquery.WriteTruncate

	job, err := loader.Run(ctx)
	if err != nil {
		log.Panicf("Failed to create load job: %v", err)
	}

	log.Printf("Job created: %s", job.ID())
	return nil
}
