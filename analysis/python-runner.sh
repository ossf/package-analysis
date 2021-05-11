#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis

for pkg in $(cat python.txt); do
  rand=$(cat /dev/urandom | tr -dc 'a-z0-9' | head -c 10)
  kubectl run pypi-$rand --image=$REGISTRY/analysis --restart='Never' \
      --privileged \
      --labels=install=1,package_type=pypi \
      --annotations=package_name=$pkg \
      --annotations=package_version=test \
      --requests="cpu=250m" -- \
      analyze \
      --image=$REGISTRY/python --command="analyze.py $pkg" \
      --bucket=gs://ossf-malware-analysis-results \
      --upload="pypi/$pkg/test/results.json"
done
