# Loader

This runs periodically as a Cloud Function to load analysis results into
BigQuery.

We use this instead of the BigQuery Data Transfer service as it does not support
load jobs with WRITE_TRUNCATE.
