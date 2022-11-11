#!/bin/bash

NOPULL=0
DRYRUN=0
LOCAL=0
VERSION_SET=0

ECOSYSTEM="$1"
shift
PACKAGE="$1"
shift

VERSION="latest"
PKG_PATH=""

RESULTS_DIR="/tmp/results"
LOGS_DIR="/tmp/dockertmp"

# for pretty printing
LINE="-----------------------------------------"


function print_usage {
	echo "Usage: $0 (npm|packagist|pypi|rubygems) <package name> [version] [--local /path/to/local/package] [--nopull] [-d|--dryrun]"
}


function print_package_details {
	echo "Ecosystem:   $ECOSYSTEM"
	echo "Package:     $PACKAGE"
	echo "Version:     $VERSION"
	if [[ $LOCAL -eq 1 ]]; then
		LOCATION="$PKG_PATH"
	else
		LOCATION="remote"
	fi

	echo "Location:    $LOCATION"
}

function print_results_dirs {
	echo "Results dir: $RESULTS_DIR"
	echo "Logs dir:    $LOGS_DIR"
}


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
		"-l"|"--local")
			LOCAL=1
			shift
			# -m preserves invalid/non-existent paths (which will be detected below)
			PKG_PATH=$(realpath -m "$1")
			if [[ -z "$PKG_PATH" ]]; then
				echo "--local specified but no package path given"
				exit 255
			fi
			shift
			;;
		*)
			if [[ $VERSION_SET -eq 0 ]]; then
				VERSION_SET=1
				VERSION="$1"
				shift
			else
				echo "Extra/unrecognised argument $1 (version already set to $VERSION)"
				exit 255
			fi
			;;
	esac
done

if [[ -z "$ECOSYSTEM" || -z "$PACKAGE" ]]; then
	print_usage
	exit 255
fi



DOCKER_OPTS=("run" "--cgroupns=host" "--privileged" "-ti")

DOCKER_MOUNTS=("-v" "/var/lib/containers:/var/lib/containers" "-v" "$RESULTS_DIR:/results" "-v" "$LOGS_DIR:/tmp")

ANALYSIS_IMAGE=gcr.io/ossf-malware-analysis/analysis

ANALYSIS_ARGS=("analyze" "-package" "$PACKAGE" "-ecosystem" "$ECOSYSTEM" "-upload" "file:///results/")

if [[ "$VERSION" != "latest" ]]; then
	ANALYSIS_ARGS+=("-version" "$VERSION")
fi

if [[ $NOPULL -eq 1 ]]; then
	ANALYSIS_ARGS+=("-nopull")
fi

if [[ $LOCAL -eq 1 ]]; then
	PKG_FILE=$(basename "$PKG_PATH")
	LOCATION="$PKG_PATH"

	# mount local package file in root of docker image
	DOCKER_MOUNTS+=("-v" "$PKG_PATH:/$PKG_FILE")

	# tell analyis to use mounted package file
	ANALYSIS_ARGS+=("-local" "/$PKG_FILE")
else
	LOCATION="remote"
fi

echo $LINE
echo "Package Details"
print_package_details
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

	if [[ $LOCAL -eq 1 ]] && [[ ! -f "$PKG_PATH" || ! -r "$PKG_PATH" ]]; then
		echo "Error: path $PKG_PATH does not refer to a file or is not readable"
		echo
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
	print_package_details
	print_results_dirs
else
	echo "Analysis failed"
	echo
	echo "docker process exited with code $DOCKER_EXIT_CODE"
	echo
	print_package_details
	rmdir --ignore-fail-on-non-empty "$RESULTS_DIR"
	rmdir --ignore-fail-on-non-empty "$LOGS_DIR"
fi

echo $LINE
exit $DOCKER_EXIT_CODE
