#!/bin/bash -ex

nopush=${NOPUSH:-"false"}

REGISTRY=gcr.io/ossf-malware-analysis
ANALYSIS_IMAGES=(
  node
  python
  ruby
)

for image in "${ANALYSIS_IMAGES[@]}"; do
  docker build --squash -t $REGISTRY/$image $image
  [[ "$nopush" == "false" ]]  && docker push $REGISTRY/$image
done

OTHER_IMAGES=(
  analysis
  scheduler
  server
)

for image in "${OTHER_IMAGES[@]}"; do
  pushd ../$image
  docker build --squash -t $REGISTRY/$image .
  [[ "$nopush" == "false" ]] && docker push $REGISTRY/$image
  popd
done
