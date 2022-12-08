#!/bin/bash

RESULTS_DIR=${RESULTS_DIR:-"/tmp/results"}
STATIC_RESULTS_DIR=${STATIC_RESULTS_DIR:-"/tmp/staticResults"}
FILE_WRITE_RESULTS_DIR=${FILE_WRITE_RESULTS_DIR:-"/tmp/writeResults"}
LOGS_DIR=${LOGS_DIR:-"/tmp/dockertmp"}

# for pretty printing
LINE="-----------------------------------------"

function print_usage {
	echo "Usage: $0 [-dryrun] <analysis args ...>"
	echo "Pass -help to see a full list of arguments to the analysis program"
	echo
	echo "-dryrun prints out the commmand that would be executed without running anything"
}

function print_package_details {
	echo "Ecosystem:                $ECOSYSTEM"
	echo "Package:                  $PACKAGE"
	echo "Version:                  $VERSION"
	if [[ $LOCAL -eq 1 ]]; then
		LOCATION="$PKG_PATH"
	else
		LOCATION="remote"
	fi

	echo "Location:                 $LOCATION"
}

function print_results_dirs {
	echo "Dynamic analysis results: $RESULTS_DIR"
	echo "Static analysis results:  $STATIC_RESULTS_DIR"
	echo "File write results:       $FILE_WRITE_RESULTS_DIR"
	echo "Debug logs:               $LOGS_DIR"
}


args=("$@")

DRYRUN=0
LOCAL=0

ECOSYSTEM=""
PACKAGE=""
VERSION=""
PKG_PATH=""
MOUNTED_PKG_PATH=""

i=0
while [[ $i -lt $# ]]; do
	case "${args[$i]}" in
		"-dryrun")
			DRYRUN=1
			unset "args[i]" # this argument is not passed to analysis image
			;;
		"-local")
			LOCAL=1
			i=$((i+1))
			# -m preserves invalid/non-existent paths (which will be detected below)
			PKG_PATH=$(realpath -m "${args[$i]}")
			if [[ -z "$PKG_PATH" ]]; then
				echo "--local specified but no package path given"
				exit 255
			fi
			PKG_FILE=$(basename "$PKG_PATH")
			MOUNTED_PKG_PATH="/$PKG_FILE"
			# need to change the path passed to analysis image to the mounted one
			# which is stripped of host path info
			args[$i]="$MOUNTED_PKG_PATH"
			;;
		"-ecosystem")
			i=$((i+1))
			ECOSYSTEM="${args[$i]}"
			;;
		"-package")
			i=$((i+1))
			PACKAGE="${args[$i]}"
			;;
		"-version")
			i=$((i+1))
			VERSION="${args[$i]}"
			;;
	esac
	i=$((i+1))
done

DOCKER_OPTS=("run" "--cgroupns=host" "--privileged" "-ti")

DOCKER_MOUNTS=("-v" "/var/lib/containers:/var/lib/containers" "-v" "$RESULTS_DIR:/results" "-v" "$STATIC_RESULTS_DIR:/staticResults" "-v" "$FILE_WRITE_RESULTS_DIR:/writeResults" "-v" "$LOGS_DIR:/tmp")

ANALYSIS_IMAGE=gcr.io/ossf-malware-analysis/analysis

ANALYSIS_ARGS=("analyze" "-upload" "file:///results/" "-upload-file-write-info" "file:///writeResults/" "-upload-static" "file:///staticResults/")

# Add the remaining command line arguments
ANALYSIS_ARGS=("${ANALYSIS_ARGS[@]}" "${args[@]}")

if [[ $LOCAL -eq 1 ]]; then
	LOCATION="$PKG_PATH"

	# mount local package file in root of docker image
	DOCKER_MOUNTS+=("-v" "$PKG_PATH:$MOUNTED_PKG_PATH")
else
	LOCATION="remote"
fi

if [[ -n "$ECOSYSTEM" && -n "$PACKAGE" ]]; then
	PACKAGE_DEFINED=1
else
	PACKAGE_DEFINED=0
fi

if [[ $PACKAGE_DEFINED -eq 1 ]]; then
	echo $LINE
	echo "Package Details"
	print_package_details
	echo $LINE
fi

# If dry run, just print the command and exit
if [[ $DRYRUN -eq 1 ]]; then
	echo "Analysis command (dry run)"
	echo
	echo docker "${DOCKER_OPTS[@]}" "${DOCKER_MOUNTS[@]}" "$ANALYSIS_IMAGE" "${ANALYSIS_ARGS[@]}"

	echo
	exit 0
fi

# Else continue execution
if [[ $PACKAGE_DEFINED -eq 1 ]]; then
	echo "Analysing package"
	echo
fi

if [[ $LOCAL -eq 1 ]] && [[ ! -f "$PKG_PATH" || ! -r "$PKG_PATH" ]]; then
	echo "Error: path $PKG_PATH does not refer to a file or is not readable"
	echo
	exit 1
fi

sleep 1 # Allow time to read info above before executing

mkdir -p "$RESULTS_DIR"
mkdir -p "$STATIC_RESULTS_DIR"
mkdir -p "$FILE_WRITE_RESULTS_DIR"
mkdir -p "$LOGS_DIR"

docker "${DOCKER_OPTS[@]}" "${DOCKER_MOUNTS[@]}" "$ANALYSIS_IMAGE" "${ANALYSIS_ARGS[@]}"

DOCKER_EXIT_CODE=$?

if [[ $PACKAGE_DEFINED -eq 1 ]]; then
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
		rmdir --ignore-fail-on-non-empty "$STATIC_RESULTS_DIR"
		rmdir --ignore-fail-on-non-empty "$FILE_WRITE_RESULTS_DIR"
		rmdir --ignore-fail-on-non-empty "$LOGS_DIR"
	fi

echo $LINE
fi

exit $DOCKER_EXIT_CODE
