# Package Analysis Infrastructure

This directory contains all the configuration, documentation and scripts needed
to manage the package analysis infrastructure.

## Production Cluster

The Production cluster runs in GCP.

To access the cluster, run:

```shell
$ gcloud container clusters get-credentials analysis-cluster --zone=us-central1-c --project=ossf-malware-analysis
```

### Updating Container Images

To update container images, run:

```shell
$ cd build
$ make push_all_images
```
