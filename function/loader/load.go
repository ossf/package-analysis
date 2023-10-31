package loader

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
)

//go:embed dynamic-analysis-schema.json
var dynamicAnalysisSchemaJSON []byte

//go:embed static-analysis-schema.json
var staticAnalysisSchemaJSON []byte

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

	schema, err := bigquery.SchemaFromJSON(dynamicAnalysisSchemaJSON)
	if err != nil {
		log.Panicf("Failed to decode schema: %v", err)
	}

	gcsRef := bigquery.NewGCSReference(fmt.Sprintf("gs://%s/*.json", bucket))
	gcsRef.Schema = schema
	gcsRef.SourceFormat = bigquery.JSON
	gcsRef.MaxBadRecords = 10000

	dataset := bq.Dataset("packages")
	loader := dataset.Table("analysis").LoaderFrom(gcsRef)
	loader.WriteDisposition = bigquery.WriteTruncate
	loader.TimePartitioning = &bigquery.TimePartitioning{
		Type:  bigquery.DayPartitioningType,
		Field: "CreatedTimestamp",
	}

	job, err := loader.Run(ctx)
	if err != nil {
		log.Panicf("Failed to create load job: %v", err)
	}

	log.Printf("Job created: %s", job.ID())
	return nil
}

func LoadStaticAnalysis(ctx context.Context, m PubSubMessage) error {
	project := os.Getenv("GCP_PROJECT")
	bucket := os.Getenv("OSSF_MALWARE_STATIC_ANALYSIS_RESULTS")

	bq, err := bigquery.NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to create BigQuery client: %w", err)

	}
	defer bq.Close()

	schema, err := bigquery.SchemaFromJSON(staticAnalysisSchemaJSON)
	if err != nil {
		return fmt.Errorf("failed to decode schema: %w", err)

	}

	gcsRef := bigquery.NewGCSReference(fmt.Sprintf("gs://%s/*.json", bucket))
	gcsRef.Schema = schema
	gcsRef.SourceFormat = bigquery.JSON
	gcsRef.MaxBadRecords = 10000

	dataset := bq.Dataset("packages")
	loader := dataset.Table("staticanalysis").LoaderFrom(gcsRef)
	loader.WriteDisposition = bigquery.WriteTruncate
	loader.TimePartitioning = &bigquery.TimePartitioning{
		Type:  bigquery.DayPartitioningType,
		Field: "created",
	}

	job, err := loader.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to create load job: %v", err)
	}

	fmt.Printf("load job created: %s\n", job.ID())

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for job: %w", err)
	}

	if status.Err() != nil {
		fmt.Printf("job completed with %d errors\n", len(status.Errors))
		for idx, err := range status.Errors {
			fmt.Printf("error %d: %v\n", idx, err)
		}

		return status.Err()
	}

	return nil
}
