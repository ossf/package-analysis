# Server

This directory contains a simple HTTP server to serve queries against
the package analysis results.

## Deployment

This is deployed on Cloud Run. Deploy by running:

```bash
$ gcloud run deploy server \
  --image=gcr.io/ossf-malware-analysis/server \
  --allow-unauthenticated \
  --set-env-vars=GOOGLE_CLOUD_PROJECT=ossf-malware-analysis \
  --platform managed \
  --project=ossf-malware-analysis
