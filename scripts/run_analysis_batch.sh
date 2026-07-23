#!/bin/bash

RESULTS_DIR=${RESULTS_DIR:-"/tmp/results"}
STATIC_RESULTS_DIR=${STATIC_RESULTS_DIR:-"/tmp/staticResults"}
FILE_WRITE_RESULTS_DIR=${FILE_WRITE_RESULTS_DIR:-"/tmp/writeResults"}
ANALYZED_PACKAGES_DIR=${ANALYZED_PACKAGES_DIR:-"/tmp/analyzedPackages"}
LOGS_DIR=${LOGS_DIR:-"/tmp/dockertmp"}
STRACE_LOGS_DIR=${STRACE_LOGS_DIR:-"/tmp/straceLogs"}


# function to create directory if it doesn't exist
function create_dir_if_not_exists {
	local dir_path=$1
	if [[ ! -d "$dir_path" ]]; then
		mkdir -p "$dir_path"
		echo "Directory created: $dir_path"
	else
		echo "Directory already exists: $dir_path"
	fi
}


# for pretty printing
LINE="-----------------------------------------"

function print_usage {
	echo "Usage: $0 [-dryrun] [-fully-offline] <analyze args...>"
	echo
	echo $LINE
	echo "Script options"
	echo "  -dryrun"
	echo "    	prints commmand that would be executed and exits"
	echo "  -fully-offline"
	echo "    	completely disables network access for the container runtime"
	echo "    	Analysis will only work when using -local <pkg path> and -nopull."
	echo "    	(see also: -offline)"
	echo "  -nointeractive"
	echo "          disables TTY input and prevents allocating pseudo-tty"
	echo $LINE
	echo
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
	echo "Analyzed package saved:   $ANALYZED_PACKAGES_DIR"
	echo "Debug logs:               $LOGS_DIR"
	echo "Strace logs:              $STRACE_LOGS_DIR"
}


args=("$@")

HELP=0
DRYRUN=0
LOCAL=0
DOCKER_OFFLINE=0
INTERACTIVE=1

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
		"-fully-offline")
			DOCKER_OFFLINE=1
			unset "args[i]" # this argument is not passed to analysis image
			;;
		"-nointeractive")
			INTERACTIVE=0
			unset "args[i]" # this argument is not passed to analysis image
			;;
		"-help")
			HELP=1
			;;
		"-local")
			# need to create a mount to pass the package archive to the docker image
			LOCAL=1
			i=$((i+1))
			# -m preserves invalid/non-existent paths (which will be detected below)
			PKG_PATH=$(realpath -m "${args[$i]}")
			if [[ -z "$PKG_PATH" ]]; then
				echo "-local specified but no package path given"
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

if [[ $# -eq 0 ]]; then
	HELP=1
fi

DOCKER_OPTS=("run" "--cgroupns=host" "--privileged" "--rm")

# On development systems, we mount /var/lib/containers so that sandbox images can be
# shared between the host system and the analysis image. However, this requires the
# directory to be backed by a non-overlay filesystem.
# In some environments, e.g. GitHub Codespaces, this is not the case, and we need to
# specify a different mount dir which is backed by a non-overlay filesystem.

# Checks that the given mountpoint has the given filesystem mount type
function is_mount_type() {
	if [[ $(findmnt -T "$2" -n -o FSTYPE) == "$1" ]]; then
		return 0
	else
		return 1
	fi
}

CONTAINER_MOUNT_DIR="/var/lib/containers"

if [[ -n "$CONTAINER_DIR_OVERRIDE" ]]; then
	CONTAINER_MOUNT_DIR="$CONTAINER_DIR_OVERRIDE"
elif [[ $CODESPACES == "true" ]]; then
	CONTAINER_MOUNT_DIR=$(mktemp -d)
	echo "GitHub Codespaces environment detected, using $CONTAINER_MOUNT_DIR for container mount"
elif is_mount_type overlay /var/lib; then
	if is_mount_type overlay /tmp && ! is_mount_type tmpfs /tmp; then
		CONTAINER_MOUNT_DIR=$(mktemp -d)
		echo "Warning: /var/lib is an overlay mount, using $CONTAINER_MOUNT_DIR for container mount"
	else
		echo "Environment error: /var/lib is an overlay mount, please set CONTAINER_DIR_OVERRIDE to a directory that is backed by a non-overlay filesystem"
		exit 1
	fi
fi


DOCKER_MOUNTS=("-v" "$CONTAINER_MOUNT_DIR:/var/lib/containers" "-v" "$RESULTS_DIR:/results" "-v" "$STATIC_RESULTS_DIR:/staticResults" "-v" "$FILE_WRITE_RESULTS_DIR:/writeResults" "-v" "$LOGS_DIR:/tmp" "-v" "$ANALYZED_PACKAGES_DIR:/analyzedPackages" "-v" "$STRACE_LOGS_DIR:/straceLogs")

ANALYSIS_IMAGE=gcr.io/ossf-malware-analysis/analysis

ANALYSIS_ARGS=("analyze" "-dynamic-bucket" "file:///results/" "-file-writes-bucket" "file:///writeResults/" "-static-bucket" "file:///staticResults/" "-analyzed-pkg-bucket" "file:///analyzedPackages/" "-execution-log-bucket" "file:///results")

# Add the remaining command line arguments
ANALYSIS_ARGS=("${ANALYSIS_ARGS[@]}" "${args[@]}")

if [[ $HELP -eq 1 ]]; then
	print_usage
fi

if [[ $INTERACTIVE -eq 1 ]]; then
	DOCKER_OPTS+=("-ti")
fi

if [[ $LOCAL -eq 1 ]]; then
	LOCATION="$PKG_PATH"

	# mount local package file in root of docker image
	DOCKER_MOUNTS+=("-v" "$PKG_PATH:$MOUNTED_PKG_PATH")
else
	LOCATION="remote"
fi

if [[ $DOCKER_OFFLINE -eq 1 ]]; then
	DOCKER_OPTS+=("--network" "none")
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

create_dir_if_not_exists "$RESULTS_DIR"
create_dir_if_not_exists "$STATIC_RESULTS_DIR"
create_dir_if_not_exists "$FILE_WRITE_RESULTS_DIR"
create_dir_if_not_exists "$ANALYZED_PACKAGES_DIR"
create_dir_if_not_exists "$LOGS_DIR"
create_dir_if_not_exists "$STRACE_LOGS_DIR"

docker "${DOCKER_OPTS[@]}" "${DOCKER_MOUNTS[@]}" "$ANALYSIS_IMAGE" "${ANALYSIS_ARGS[@]}"

DOCKER_EXIT_CODE=$?
# define the results naming convention
RESULTS_PREFIX="${ECOSYSTEM}-${PACKAGE}-${VERSION}"

if [[ $PACKAGE_DEFINED -eq 1 ]]; then
echo
echo $LINE
	if [[ $DOCKER_EXIT_CODE -eq 0 ]]; then
		echo "Finished analysis"
		echo
		print_package_details
		print_results_dirs

		# rename and move each type of result file with the prefix
		# Rename and move each type of result file with the prefix
		for dir in "$RESULTS_DIR" "$STATIC_RESULTS_DIR" "$FILE_WRITE_RESULTS_DIR" "$ANALYZED_PACKAGES_DIR" "$LOGS_DIR" "$STRACE_LOGS_DIR"; do
    			for file in "$dir"/*; do
        			if [[ -e "$file" ]]; then
            				filename=$(basename "$file")
            				extension="${filename##*.}"
            				new_name="$dir/${RESULTS_PREFIX}.${extension}"

            				# Check if the new filename already exists
            				if [[ -e "$new_name" ]]; then
                				echo "Warning: $new_name already exists. Skipping rename for $file."
            				else
                				mv "$file" "$new_name"
                				echo "Renamed $file to $new_name"
            				fi
        			fi
    			done
		done

#	else
#		echo "Analysis failed"
#		echo
#		echo "docker process exited with code $DOCKER_EXIT_CODE"
#		echo
#		print_package_details
#		rmdir --ignore-fail-on-non-empty "$RESULTS_DIR"
#		rmdir --ignore-fail-on-non-empty "$STATIC_RESULTS_DIR"
#		rmdir --ignore-fail-on-non-empty "$FILE_WRITE_RESULTS_DIR"
#		rmdir --ignore-fail-on-non-empty "$ANALYZED_PACKAGES_DIR"
#		rmdir --ignore-fail-on-non-empty "$LOGS_DIR"
#		rmdir --ignore-fail-on-non-empty "$STRACE_LOGS_DIR"
	fi

echo $LINE
fi

exit $DOCKER_EXIT_CODE
