#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis
IMAGES=(
  analysis
  node
  python
  ruby
)

rm -rf analysis/analysis
cp -r ../analysis analysis/

for image in "${IMAGES[@]}"; do
  docker build -t $REGISTRY/$image $image
done

