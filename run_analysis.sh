#!/bin/bash

ECOSYSTEM="$1"
PACKAGE="$2"
PKG_PATH="$3"

if [[ $# != 2 && $# != 3 ]]; then
	echo "Usage: $0 (npm|packagist|pypi|rubygems) <package name> [/path/to/local/package.file]"
	exit 255
fi

RESULTS_DIR="/tmp/results"
LOGS_DIR="/tmp/dockertmp"

LINE="-----------------------------------------"

mkdir -p "$RESULTS_DIR"
mkdir -p "$LOGS_DIR"


DOCKER_OPTS=(run --cgroupns=host --privileged -ti)

DOCKER_MOUNTS=(-v /var/lib/containers:/var/lib/containers -v "$RESULTS_DIR":/results -v "$LOGS_DIR":/tmp)

ANALYSIS_IMAGE=gcr.io/ossf-malware-analysis/analysis

ANALYSIS_ARGS=(analyze -package "$PACKAGE" -ecosystem "$ECOSYSTEM" -upload file:///results/)

echo $LINE

if [[ -n "$PKG_PATH" ]]; then 
	# local mode
	PKG_FILE=$(basename "$PKG_PATH")
	echo "Analysing $ECOSYSTEM package $PACKAGE from local file $PKG_FILE"
	echo "Path: $PKG_PATH"

	# mount local package file in root of docker image
	DOCKER_MOUNTS+=(-v "$PKG_PATH:/$PKG_FILE")

	# tell analyis to use mounted package file
	ANALYSIS_ARGS+=(-local "/$PKG_FILE")
else
	# remote mode
	echo "Analysing $ECOSYSTEM package $PACKAGE (remote)"
fi

echo $LINE
echo

sleep 1

docker ${DOCKER_OPTS[@]} ${DOCKER_MOUNTS[@]} $ANALYSIS_IMAGE ${ANALYSIS_ARGS[@]}
DOCKER_EXIT_CODE=$?

echo

if [[ $DOCKER_EXIT_CODE == 0 ]]; then
	echo "Finished"
	echo "Results dir: $RESULTS_DIR"
	echo "Logs dir: $LOGS_DIR"
else
	echo "Docker process exited with nonzero exit code $DOCKER_EXIT_CODE"
	rmdir --ignore-fail-on-non-empty "$RESULTS_DIR"
	rmdir --ignore-fail-on-non-empty "$LOGS_DIR"
	exit $DOCKER_EXIT_CODE
fi
