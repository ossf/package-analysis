#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis

for pkg in $(cat python.txt); do
  rand=$(cat /dev/urandom | tr -dc 'a-z0-9' | head -c 10)
  kubectl run pypi-$rand --image=$REGISTRY/python --restart='Never' \
      --labels=install=1,package_type=pypi \
      --annotations=package_name=$pkg \
      --annotations=package_version=test \
      --requests="cpu=250m" -- sh -c "analyze.py $pkg"
done
