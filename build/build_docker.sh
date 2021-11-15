#!/bin/bash -ex

nopush=${NOPUSH:-"false"}

BASE_PATH="$(dirname $(dirname $(realpath $0)))"
REGISTRY=gcr.io/ossf-malware-analysis

# Mapping from the container name to the path containing the Dockerfile.
declare -A ANALYSIS_IMAGES=( [node]=npm [python]=pypi [ruby]=rubygems )

pushd "$BASE_PATH/sandboxes"
for image in "${!ANALYSIS_IMAGES[@]}"; do
  docker build -t $REGISTRY/$image ${ANALYSIS_IMAGES[$image]}
  [[ "$nopush" == "false" ]]  && docker push $REGISTRY/$image
done
popd

# Mapping from the container name to the path containing the Dockerfile.
declare -A CMD_IMAGES=( [analysis]=analyze [scheduler]=scheduler [server]=server )

pushd "$BASE_PATH"
for image in "${!CMD_IMAGES[@]}"; do
  docker build --build-arg NO_PODMAN_PULL=$NO_PODMAN_PULL -t $REGISTRY/$image -f cmd/${CMD_IMAGES[$image]}/Dockerfile .
  [[ "$nopush" == "false" ]] && docker push $REGISTRY/$image
done
popd
