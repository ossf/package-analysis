#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis
IMAGES=(
  node
  python
)

for image in "${IMAGES[@]}"; do
  docker build -t $REGISTRY/$image $image
done

