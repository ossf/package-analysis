#!/bin/bash

REGISTRY=gcr.io/ossf-malware-analysis
IMAGES=(
  node
  python
  ruby
)

for image in "${IMAGES[@]}"; do
  docker build --squash -t $REGISTRY/$image $image

  # Flatten the image.
  container=$(docker create $REGISTRY/$image)
  docker export $container | docker import - $REGISTRY/$image:flat
  docker rm $container

  docker push $REGISTRY/$image
  docker push $REGISTRY/$image:flat
done

rm -rf analysis/analysis
cp -r ../analysis analysis/
docker build --squash -t $REGISTRY/analysis analysis
docker push $REGISTRY/analysis
