#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis

for pkg in $(cat rubygems.txt); do
  rand=$(cat /dev/urandom | tr -dc 'a-z0-9' | head -c 10)
  kubectl run rubygems-$rand --image=$REGISTRY/analysis --restart='Never' \
      --privileged \
      --labels=install=1,package_type=rubygems \
      --annotations=package_name=$pkg \
      --annotations=package_version=test \
      --requests="cpu=250m" -- \
      analyze \
      --package="rubygems/$pkg" \
      --upload="gs://ossf-malware-analysis-results-test/rubygems/$pkg"
done
