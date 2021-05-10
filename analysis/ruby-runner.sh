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
      --image=$REGISTRY/ruby --command="analyze.rb $pkg" \
      --bucket=gs://ossf-malware-analysis-results \
      --upload="rubygems/$pkg/test/results.json"
done
