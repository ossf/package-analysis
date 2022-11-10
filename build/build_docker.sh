#!/bin/bash -ex

push=${PUSH:-"false"}
tag=${RELEASE_TAG}

BASE_PATH="$(dirname "$(dirname "$(realpath "$0")")")"
REGISTRY=gcr.io/ossf-malware-analysis

# Mapping from the container name to the path containing the Dockerfile.
declare -A ANALYSIS_IMAGES=( [node]=npm [python]=pypi [ruby]=rubygems [packagist]=packagist [crates.io]=crates.io )

pushd "$BASE_PATH/sandboxes"
for image in "${!ANALYSIS_IMAGES[@]}"; do
  extra_args=""
  if [ "$tag" != "" ]; then
    extra_args="-t $REGISTRY/$image:$tag"
  fi
  docker build $extra_args -t $REGISTRY/$image ${ANALYSIS_IMAGES[$image]}
  [[ "$push" == "true" ]] && docker push --all-tags $REGISTRY/$image
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
  [[ "$push" == "true" ]] && docker push --all-tags $REGISTRY/$image
done
popd
