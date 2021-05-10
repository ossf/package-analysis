# Configs and code to run analysis jobs

## Cluster

To access the cluster, run:

```shell
gcloud container clusters get-credentials analysis-cluster --zone=us-central1-c --project=ossf-malware-analysis
```

### Setup

Workload Identity is enabled for uploads to GCS.

## Deployment

The code in this directory is by building the docker image in `build/analysis`.

## Analysis

The two scripts in this repo can be run directly:

```shell
./python-runner.sh
./node-runner.sh
```

To clean up, run:

```shell
kubectl delete pod -l install=1
```

The .txt files contain package data from [NPM](https://medium.com/r/?url=https%3A%2F%2Fwww.npmjs.com%2Fbrowse%2Fdepended) and [PyPI](https://medium.com/r/?url=https%3A%2F%2Fhugovk.github.io%2Ftop-pypi-packages%2Ftop-pypi-packages-30-days.json).
