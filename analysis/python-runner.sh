#!/bin/bash

for pkg in $(cat python.txt); do
  rand=$(cat /dev/urandom | tr -dc 'a-z0-9' | head -c 10)
  kubectl run pypi-$rand --image=gcr.io/ossf-malware-analysis/python --restart='Never' \
      --labels=install=1,package_type=pypi \
      --annotations=package_name=$pkg \
      --annotations=package_version=test \
      --requests="cpu=250m" -- sh -c "mkdir -p /app && cd /app && analyze.py $pkg"
done
