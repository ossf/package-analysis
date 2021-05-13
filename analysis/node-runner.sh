#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis

for pkg in $(cat node.txt); do
  rand=$(cat /dev/urandom | tr -dc 'a-z0-9' | head -c 10)
  kubectl run npm-$rand --image=$REGISTRY/analysis --restart='Never' \
      --privileged \
      --labels=install=1,package_type=npm \
      --annotations=package_name=$pkg \
      --annotations=package_version=test \
      --requests="cpu=250m" -- \
      analyze \
      --package="npm/$pkg" \
      --upload="gs://ossf-malware-analysis-results-test/npm/$pkg/test/results.json"
done
