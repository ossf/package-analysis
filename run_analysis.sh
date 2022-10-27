#!/bin/bash

NOPULL=0
DRYRUN=0

ECOSYSTEM="$1"
shift
PACKAGE="$1"
shift

while [[ $# -gt 0 ]]; do
	case "$1" in
		"--nopull")
			NOPULL=1
			shift
			;;
		"-d"|"--dryrun")
			DRYRUN=1
			shift
			;;
		*)
			if [[ -z "$PKG_PATH" ]]; then
				# -m preserves invalid/non-existent paths (which will be detected below)
				PKG_PATH=$(realpath -m "$1")
				shift
			else
				echo "Extra/unrecognised argument $1 (local package path already set to $PKG_PATH)"
				exit 255
			fi
			;;
	esac
done

if [[ -z "$ECOSYSTEM" || -z "$PACKAGE" ]]; then
	echo "Usage: $0 (npm|packagist|pypi|rubygems) <package name> [/path/to/local/package.file] [--nopull] [-d|--dryrun]"
	exit 255
fi



RESULTS_DIR="/tmp/results"
LOGS_DIR="/tmp/dockertmp"

LINE="-----------------------------------------"

DOCKER_OPTS=(run --cgroupns=host --privileged -ti)

DOCKER_MOUNTS=(-v /var/lib/containers:/var/lib/containers -v "$RESULTS_DIR":/results -v "$LOGS_DIR":/tmp)

ANALYSIS_IMAGE=gcr.io/ossf-malware-analysis/analysis

ANALYSIS_ARGS=(analyze -package "$PACKAGE" -ecosystem "$ECOSYSTEM" -upload file:///results/)

if [[ $NOPULL -eq 1 ]]; then
	ANALYSIS_ARGS+=(-nopull)
fi


if [[ -n "$PKG_PATH" ]]; then
	# local mode

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

if [[ $DRYRUN -eq 1 ]]; then
	echo "Analysis command (dry run)"
	echo
	echo docker "${DOCKER_OPTS[@]}" "${DOCKER_MOUNTS[@]}" "$ANALYSIS_IMAGE" "${ANALYSIS_ARGS[@]}"

	echo
	exit 0
else
	echo "Analysing package"
	echo

	if [[ -n "$PKG_PATH" ]] && [[ ! -f "$PKG_PATH" || ! -r "$PKG_PATH" ]]; then
		echo "Error: path $PKG_PATH does not refer to a file or is not readable"
		exit 1
	fi

	sleep 1 # Allow time to read info above before executing

	mkdir -p "$RESULTS_DIR"
	mkdir -p "$LOGS_DIR"

	docker "${DOCKER_OPTS[@]}" "${DOCKER_MOUNTS[@]}" "$ANALYSIS_IMAGE" "${ANALYSIS_ARGS[@]}"
fi

DOCKER_EXIT_CODE=$?

echo
echo $LINE

if [[ $DOCKER_EXIT_CODE -eq 0 ]]; then
	echo "Finished analysis"
	echo
	echo "Ecosystem:   $ECOSYSTEM"
	echo "Package:     $PACKAGE"
	echo "Results dir: $RESULTS_DIR"
	echo "Logs dir:    $LOGS_DIR"
else
	echo "Analysis failed"
	echo
	echo "docker process exited with code $DOCKER_EXIT_CODE"
	echo
	echo "Ecosystem:   $ECOSYSTEM"
	echo "Package:     $PACKAGE"
	rmdir --ignore-fail-on-non-empty "$RESULTS_DIR"
	rmdir --ignore-fail-on-non-empty "$LOGS_DIR"
fi

echo $LINE
exit $DOCKER_EXIT_CODE
