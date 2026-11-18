#!/bin/bash
#
# Load locally-built sandbox images into a Docker named volume for use on
# macOS (Apple Silicon or Intel).
#
# Background: the `analysis` image runs `podman` internally to launch the
# dynamic/static sandbox containers. That nested podman reads its image store
# from /var/lib/containers, which scripts/run_analysis.sh bind-mounts from the
# host. On Linux dev boxes this is populated with `make sync` (sudo podman load
# on the host). macOS has no host podman, and a plain host-directory bind mount
# uses virtiofs, on which podman's overlay storage driver cannot create
# whiteouts ("kernel does not support overlay fs").
#
# The fix is to back that store with a Docker *named volume* (stored on the
# Docker Desktop VM's own filesystem, where overlay works) and to pre-load the
# sandbox images into it once. run_analysis.sh treats a CONTAINER_DIR_OVERRIDE
# value with no leading slash as a named volume, so:
#
#   export CONTAINER_DIR_OVERRIDE=pa-containers
#   scripts/run_analysis.sh -nopull -local <pkg> ...
#
# will then find the images this script loads.
#
# Prerequisites: `make build` has produced the three local images
# (analysis, dynamic-analysis, static-analysis).

set -euo pipefail

REGISTRY="gcr.io/ossf-malware-analysis"
VOLUME="${CONTAINER_DIR_OVERRIDE:-pa-containers}"
ANALYSIS_IMAGE="${REGISTRY}/analysis:latest"
DYNAMIC_IMAGE="${REGISTRY}/dynamic-analysis:latest"
STATIC_IMAGE="${REGISTRY}/static-analysis:latest"

if ! docker info >/dev/null 2>&1; then
	echo "Error: Docker is not running. Start Docker Desktop first." >&2
	exit 1
fi

for img in "$ANALYSIS_IMAGE" "$DYNAMIC_IMAGE" "$STATIC_IMAGE"; do
	if ! docker image inspect "$img" >/dev/null 2>&1; then
		echo "Error: image '$img' not found locally. Run 'make build' first." >&2
		exit 1
	fi
done

# podman preserves the image config digest as its IMAGE ID, which matches
# docker's image ID, so we can re-tag reliably by ID after loading.
DYNAMIC_ID="$(docker image inspect "$DYNAMIC_IMAGE" --format '{{.Id}}' | cut -d: -f2 | cut -c1-12)"
STATIC_ID="$(docker image inspect "$STATIC_IMAGE" --format '{{.Id}}' | cut -d: -f2 | cut -c1-12)"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "Saving sandbox images (the dynamic image is several GB; this takes a while)..."
docker save "$DYNAMIC_IMAGE" -o "$TMPDIR/dynamic.tar"
docker save "$STATIC_IMAGE" -o "$TMPDIR/static.tar"

echo "Creating named volume '$VOLUME' and loading images into the nested podman store..."
docker volume create "$VOLUME" >/dev/null

docker run --rm --privileged \
	-v "$VOLUME:/var/lib/containers" \
	-v "$TMPDIR/dynamic.tar:/dynamic.tar:ro" \
	-v "$TMPDIR/static.tar:/static.tar:ro" \
	-e "DYNAMIC_IMAGE=$DYNAMIC_IMAGE" \
	-e "STATIC_IMAGE=$STATIC_IMAGE" \
	-e "DYNAMIC_ID=$DYNAMIC_ID" \
	-e "STATIC_ID=$STATIC_ID" \
	"$ANALYSIS_IMAGE" \
	bash -c '
		set -e
		podman load -i /dynamic.tar
		podman load -i /static.tar
		# podman load can lose the original repo tags; re-tag by image ID.
		podman tag "$DYNAMIC_ID" "$DYNAMIC_IMAGE"
		podman tag "$STATIC_ID" "$STATIC_IMAGE"
		echo "--- sandbox images now available to podman ---"
		podman images --format "{{.Repository}}:{{.Tag}} {{.ID}}"
	'

echo
echo "Done. To run an analysis using these images:"
echo
echo "  export CONTAINER_DIR_OVERRIDE=$VOLUME"
echo "  scripts/run_analysis.sh -nopull -local <package> -ecosystem <eco> -package <name>"
