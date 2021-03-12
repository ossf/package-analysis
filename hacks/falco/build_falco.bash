#!/bin/bash
# Build falco with increased ring buffer size for the eBPF probe.
# https://github.com/ossf/package-analysis/issues/10

FALCO_VERSION=0.27.0
REV="-3"
DOCKER_IMAGE=gcr.io/ossf-malware-analysis/falco:$FALCO_VERSION$REV
FALCO_GIT=https://github.com/falcosecurity/falco

if [ -d falco ]; then
  cd falco
  git checkout .
  git fetch
else
  git clone $FALCO_GIT
  cd falco
fi

git checkout $FALCO_VERSION

# Replace the existing libscap patch with ours to patch the ring buffer size.
cp ../libscap.patch cmake/modules/sysdig-repo/patch/libscap.patch

mkdir -p build
docker pull falcosecurity/falco-builder
docker run -ti --rm -v $PWD/..:/source -v $PWD/build:/build falcosecurity/falco-builder cmake
docker run -ti --rm -v $PWD/..:/source -v $PWD/build:/build falcosecurity/falco-builder package
cd ../

docker build -t $DOCKER_IMAGE --build-arg FALCO_VERSION=$FALCO_VERSION .
docker push $DOCKER_IMAGE
