#!/bin/bash

ECOSYSTEM="$1"
PACKAGE="$2"

if [[ $# != 2 ]]; then
	echo "Usage: $0 (npm|packagist|pypi|rubygems) <package name>"
	exit -1
fi

RESULTS_DIR="/tmp/results"
LOGS_DIR="/tmp/dockertmp"

mkdir -p "$RESULTS_DIR"
mkdir -p "$LOGS_DIR"

docker run --cgroupns=host --privileged -ti \
    -v "$RESULTS_DIR":/results \
    -v "$LOGS_DIR":/tmp \
    -v /var/lib/containers:/var/lib/containers \
    gcr.io/ossf-malware-analysis/analysis analyze \
    -package "$PACKAGE" -ecosystem "$ECOSYSTEM" \
    -upload file:///results/

EXIT_CODE=$?

echo
if [[ $EXIT_CODE == 0 ]]; then
	echo "Finished"
	echo "Results dir: $RESULTS_DIR"
	echo "Logs dir: $LOGS_DIR"
else
	echo "Docker process exited with nonzero exit code $EXIT_CODE"
	rmdir --ignore-fail-on-non-empty "$RESULTS_DIR"
	rmdir --ignore-fail-on-non-empty "$LOGS_DIR"
	exit $EXIT_CODE
fi
