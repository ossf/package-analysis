# This Makefile contains commands for building the example custom sandbox and syncing it to the local container cache

# Registry for Docker images built and used by package analysis
REGISTRY := gcr.io/ossf-malware-analysis
IMAGE_NAME := dynamic-analysis-custom

# Build the sandbox
build_example_sandbox: DOCKERFILE=$(SANDBOX_DIR)/example/Dockerfile
	docker build -t ${REGISTRY}/$(IMAGE_NAME)

# Update (sync) locally built sandbox images from Docker to podman.
# This is needed for local analysis; in order to use the updated image,
# pass '-nopull' to scripts/run_analysis.sh
#
sync_example_sandbox:
	sudo buildah pull docker-daemon:${REGISTRY}/${IMAGE_NAME}
