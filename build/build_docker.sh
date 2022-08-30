#!/bin/bash -ex

nopush=${NOPUSH:-"true"}
tag=${RELEASE_TAG}

BASE_PATH="$(dirname $(dirname $(realpath $0)))"
REGISTRY=gcr.io/ossf-malware-analysis

# Mapping from the container name to the path containing the Dockerfile.
declare -A ANALYSIS_IMAGES=( [node]=npm [python]=pypi [ruby]=rubygems [packagist]=packagist )

pushd "$BASE_PATH/sandboxes"
for image in "${!ANALYSIS_IMAGES[@]}"; do
  extra_args=""
  if [ "$tag" != "" ]; then
    extra_args="-t $REGISTRY/$image:$tag"
  fi
  docker build $extra_args -t $REGISTRY/$image ${ANALYSIS_IMAGES[$image]}
  [[ "$nopush" == "false" ]]  && docker push $REGISTRY/$image
done
popd

# Mapping from the container name to the path containing the Dockerfile.
declare -A CMD_IMAGES=( [analysis]=analyze [scheduler]=scheduler )

pushd "$BASE_PATH"
for image in "${!CMD_IMAGES[@]}"; do
  extra_args=""
  if [ "$tag" != "" ]; then
    extra_args="-t $REGISTRY/$image:$tag --build-arg=SANDBOX_IMAGE_TAG=$tag"
  fi
  docker build $extra_args -t $REGISTRY/$image -f cmd/${CMD_IMAGES[$image]}/Dockerfile .
  [[ "$nopush" == "false" ]] && docker push $REGISTRY/$image
done
popd
