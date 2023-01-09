#! /bin/bash

# This script runs static analysis on all packages in a directory and
# creates a new directory with all the static analysis results for each package.
# Currently, it only supports NPM packages (as static analysis does).


ARCHIVES_DIR="$1"
RESULTS_DIR=${2:-"$ARCHIVES_DIR-results"}
START_LETTER="$3"

if [[ -z "$ARCHIVES_DIR" ]]; then
	echo "Archives dir not provided, please specify directory of .tgz archives"
	exit 1
fi

if [[ ! -d "$ARCHIVES_DIR" ]]; then
	echo "error: archives dir is not a directory"
	exit 1
fi


mkdir -p "$RESULTS_DIR"

function process_archive {
	ARCHIVE_PATH="$1"
	RESULTS_DIR="$2"
	START_LETTER="$3"
	if [[ -z "$ARCHIVE_PATH" ]]; then
		echo "Archive path is empty"
		return 1
	elif [[ -z "$RESULTS_DIR" ]]; then
		echo "Results dir is empty"
		return 1
	fi

	rm -rf /tmp/staticResults
	PACKAGE_VERSION_EXT=${ARCHIVE_PATH##"$ARCHIVES_DIR/"}
	PACKAGE_VERSION=${PACKAGE_VERSION_EXT%%.tgz}
	PACKAGE_FIRST_LETTER=${PACKAGE_VERSION:0:1}
	if [[ "$PACKAGE_FIRST_LETTER" < "$START_LETTER" ]]; then
		echo SKIP $PACKAGE_VERSION
		return
	fi
	# package name is everything before the last '-' character
	# package version is everything between the last '-' character and .tgz
	PACKAGE=$(python3 -c "print('-'.join(\"$PACKAGE_VERSION\".split('-')[:-1]))")
	VERSION=$(python3 -c "print(\"$PACKAGE_VERSION\".split('-')[-1])")
	echo "Package: $PACKAGE"
	echo "Version: $VERSION"

	OUTPUT_RESULTS_DIR=$(mktemp -d)
	STATIC_RESULTS_DIR=$OUTPUT_RESULTS_DIR ~/package-analysis/run_analysis.sh -ecosystem npm -package "$PACKAGE" -local "$ARCHIVE_PATH" -nopull -mode static -offline -fully-offline

	# python -m json.tool pretty prints JSON but it's a bit verbose; the
	# awk script prints some of the small JSON structs on a single line
	python3 -m json.tool "$OUTPUT_RESULTS_DIR/results.json" \
		| awk -f process-json.awk - \
		> "$RESULTS_DIR/$PACKAGE_VERSION-results.json"

	rm -rf "$OUTPUT_RESULTS_DIR"
}

for ARCHIVE_PATH in "$ARCHIVES_DIR"/*.tgz; do
	process_archive "$ARCHIVE_PATH" "$RESULTS_DIR" "$START_LETTER"
done

# TODO parallelise loop
# export -f process_archive
# shopt -s nullglob
# ARCHIVES=("$ARCHIVES_DIR"/*.tgz)
# parallel -i process_archive "{}" "$RESULTS_DIR" "$START_LETTER" ::: ${ARCHIVES[@]}
