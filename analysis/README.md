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

## Local usage

To run the analysis code locally, the easiest way is to use the Docker image
`gcr.io/ossf-malware-analysis/analysis`. This can be built from
`../build/build_docker.sh`.

This container uses `podman` to run a nested, sandboxed ([gVisor]) container for
analysis.

The following commands will dump the JSON result to `/tmp/results`.

[gVisor]: https://gvisor.dev/

### Live package
To run this on a live package (e.g. the "Django" package on https://pypi.org)

```bash
$ mkdir /tmp/results
$ docker run --privileged -ti -v \
    /tmp/results:/results \
    gcr.io/ossf-malware-analysis/analysis analyze \
    -package Django -ecosystem pypi \
    -upload file:///results/
```

### Local package
To run this on a local package archive (e.g. `/path/to/test.whl`), it needs to
be mounted into the the container.

```bash
$ mkdir /tmp/results
$ docker run --privileged -ti -v \
    /tmp/results:/results \
    /path/to/test.whl:/test.whl \
    gcr.io/ossf-malware-analysis/analysis analyze \
    -local /test.whl -ecosystem pypi \
    -upload file:///results/
```
