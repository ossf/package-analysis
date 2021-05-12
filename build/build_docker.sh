#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis
IMAGES=(
  node
  python
  ruby
  analysis
)

rm -rf analysis/analysis
cp -r ../analysis analysis/

for image in "${IMAGES[@]}"; do
  docker build --squash -t $REGISTRY/$image $image
done
