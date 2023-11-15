# Loader

This runs periodically as a Cloud Function to load analysis results into
BigQuery.

We use this instead of the BigQuery Data Transfer service as it does not support
load jobs with `WRITE_TRUNCATE`.

To deploy, run the following command in this directory (/function/loader):

## Dynamic analysis results

```bash
gcloud functions deploy load-data \
    --region=us-central1 \
    --project=ossf-malware-analysis \
    --entry-point=Load \
    --memory=512MB \
    --runtime=go121 \
    --timeout=120s \
    --trigger-topic=load-data \
    --set-env-vars=OSSF_MALWARE_ANALYSIS_RESULTS=ossf-malware-analysis-results,GCP_PROJECT=ossf-malware-analysis
```

## Static analysis results

```bash
gcloud functions deploy load-staticanalysis-data \
    --region=us-central1 \
    --project=ossf-malware-analysis \
    --entry-point=LoadStaticAnalysis \
    --memory=512MB \
    --runtime=go121 \
    --timeout=120s \
    --trigger-topic=load-data \
    --set-env-vars=OSSF_MALWARE_STATIC_ANALYSIS_RESULTS=ossf-malware-static-analysis-results-v1,GCP_PROJECT=ossf-malware-analysis
```
