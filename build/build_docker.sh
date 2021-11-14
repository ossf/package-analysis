#!/bin/bash -ex

nopush=${NOPUSH:-"false"}

BASE_PATH="$(dirname $(dirname $(realpath $0)))"
REGISTRY=gcr.io/ossf-malware-analysis

# TODO: rename the Docker images
declare -A ANALYSIS_IMAGES=( [node]=npm [python]=pypi [ruby]=rubygems )

pushd "$BASE_PATH/sandboxes"
for image in "${!ANALYSIS_IMAGES[@]}"; do
  docker build -t $REGISTRY/$image ${ANALYSIS_IMAGES[$image]}
  [[ "$nopush" == "false" ]]  && docker push $REGISTRY/$image
done
popd

OTHER_IMAGES=(
  analysis
)

for image in "${OTHER_IMAGES[@]}"; do
  pushd "$BASE_PATH/$image"
  docker build --build-arg NO_PODMAN_PULL=$NO_PODMAN_PULL -t $REGISTRY/$image .
  [[ "$nopush" == "false" ]] && docker push $REGISTRY/$image
  popd
done

CMD_IMAGES=(
  scheduler
  server
)

pushd "$BASE_PATH"
for image in "${CMD_IMAGES[@]}"; do
  docker build --build-arg NO_PODMAN_PULL=$NO_PODMAN_PULL -t $REGISTRY/$image -f cmd/$image/Dockerfile .
  [[ "$nopush" == "false" ]] && docker push $REGISTRY/$image
done
popd
