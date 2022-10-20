#!/bin/bash

ECOSYSTEM="$1"
PACKAGE="$2"
PKG_PATH="$3"

NOPULL=0
# Check for --nopull as last arg
if [[ ${@:$#:$#} == "--nopull" ]]; then
	NOPULL=1
	# remove last arg (set args to arg[0:arg[-1]])
	set -- "${@:1:$(($#-1))}"
fi


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

if [[ $NOPULL -eq 1 ]]; then
	ANALYSIS_ARGS+=(-nopull)
fi



if [[ -n "$PKG_PATH" ]]; then
	# local mode

	if [[ ! -f "$PKG_PATH" || ! -r "$PKG_PATH" ]]; then
		echo "Error: path $PKG_PATH does not refer to a file or is not readable"
		exit 1
	fi

	PKG_FILE=$(basename "$PKG_PATH")
	LOCATION="$PKG_PATH"

	# mount local package file in root of docker image
	DOCKER_MOUNTS+=(-v "$PKG_PATH:/$PKG_FILE")

	# tell analyis to use mounted package file
	ANALYSIS_ARGS+=(-local "/$PKG_FILE")
else
	# remote mode
	LOCATION="remote (upstream $ECOSYSTEM)"
fi

echo $LINE
echo "Package Details"
echo "Ecosystem: $ECOSYSTEM"
echo "Name:      $PACKAGE"
echo "Location:  $LOCATION"
echo $LINE

echo "Analysing package"
echo

# Print out command and allow time to read
echo docker ${DOCKER_OPTS[@]} ${DOCKER_MOUNTS[@]} $ANALYSIS_IMAGE ${ANALYSIS_ARGS[@]}
sleep 1

docker ${DOCKER_OPTS[@]} ${DOCKER_MOUNTS[@]} $ANALYSIS_IMAGE ${ANALYSIS_ARGS[@]}

DOCKER_EXIT_CODE=$?

echo

echo $LINE

PACKAGE_SUMMARY="$PACKAGE [$ECOSYSTEM]"

if [[ $DOCKER_EXIT_CODE == 0 ]]; then
	echo "Finished analysis"
	echo "Package:     $PACKAGE_SUMMARY"
	echo "Results dir: $RESULTS_DIR"
	echo "Logs dir:    $LOGS_DIR"
else
	echo "Docker process exited with nonzero exit code $DOCKER_EXIT_CODE"
	echo
	echo "Analysis failed"
	echo "Package: $PACKAGE_SUMMARY"
	rmdir --ignore-fail-on-non-empty "$RESULTS_DIR"
	rmdir --ignore-fail-on-non-empty "$LOGS_DIR"
	exit $DOCKER_EXIT_CODE
fi
echo $LINE
