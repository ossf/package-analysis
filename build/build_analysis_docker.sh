#!/bin/bash -ex

# Rebuilds the local analysis image only

BASE_PATH="$(dirname $(dirname $(realpath $0)))"
REGISTRY=gcr.io/ossf-malware-analysis

pushd "$BASE_PATH"
docker build -t $REGISTRY/analysis -f cmd/analyze/Dockerfile .
popd
