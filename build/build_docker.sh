#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis
IMAGES=(
  node
  python
  ruby
)

for image in "${IMAGES[@]}"; do
  docker build --squash -t $REGISTRY/$image $image
  docker push $REGISTRY/$image
done

rm -rf analysis/analysis
cp -r ../analysis analysis/
docker build --squash -t $REGISTRY/analysis analysis
docker push $REGISTRY/analysis
