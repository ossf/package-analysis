#!/bin/bash
# Injects packages to analyze into the PubSub queue for workers to analyze.
# E.g.: ./analysis-runner.sh npm node.txt

TOPIC=${OSSMALWARE_WORKER_TOPIC:-projects/ossf-malware-analysis/topics/workers}

function usage {
  echo "Usage: $(basename $0) ECOSYSTEM [PACKAGES_FILE]"
  echo "  Where:"
  echo "   ECOSYSTEM is one of: npm, pypi, rubygems."
  echo "   PACKAGES_FILE contains a list of package namess, one per line."
  echo "   If omitted STDIN will be used instead."
}

if [ "$#" -eq 0 -o "$#" -ge 3 ]; then
  echo "ERROR: Invalid arguments"
  usage
  exit 1
fi

ecosystem="$1"
pkg_file="$2"

# Make sure the passed in ecosystem is valid.
if [ "$ecosystem" != "npm" -a "$ecosystem" != "pypi" -a "$ecosystem" != "rubygems" ]; then
  echo "ERROR: Invalid ECOSYSTEM"
  usage
  exit 1
fi

# Make sure, if there is a file arg, that it exists and is a file.
if [ "$pkg_file" != "" -a ! -f $pkg_file ]; then
  echo "ERROR: Invalid PACKAGES_FILE"
  usage
  exit 1
fi

# If the file is set, then read it, otherwise we'll read in from stdin.
# The extra condition at the end of the while ensures that a line without a
# trailing \n (i.e. at the end of the file) is still processed.
if [ "$pkg_file" != "" ]; then
  cat "$pkg_file"
else
  cat
fi | while IFS= read -r pkg || [[ -n "$pkg" ]]; do
  gcloud pubsub topics publish "$TOPIC" --attribute=name=$pkg,ecosystem=$ecosystem
done
