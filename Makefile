REGISTRY=gcr.io/ossf-malware-analysis

.PHONY: all all_docker analysis_docker check_scripts run static_analysis_sandbox

all: build_dynamic_analysis_all

# This builds everything except the static analysis sandbox
build_dynamic_analysis_all:
	bash build/build_docker.sh

# This builds just the local analysis/worker image
build_local_worker:
	docker build -t ${REGISTRY}/analysis -f cmd/analyze/Dockerfile .

# This builds just the static analysis container image
build_static_analysis_sandbox:
	docker build -t ${REGISTRY}/static-analysis -f sandboxes/staticanalysis/Dockerfile .

build_dynamic_analysis_sandboxes:

# This updates the local sandbox image to use when running local static analysis
# Ensure 'nopull' is passed to run_analysis.sh
sync_static_analysis_sandbox_locally:
	sudo buildah pull docker-daemon:${REGISTRY}/static-analysis:latest

# This updates the local sandbox images to use when running local dynamic analysis
# Ensure 'nopull' is passed to run_analysis.sh
sync_dynamic_analysis_sandboxes_locally:
	sudo buildah pull docker-daemon:${REGISTRY}/node:latest
	sudo buildah pull docker-daemon:${REGISTRY}/python:latest
	sudo buildah pull docker-daemon:${REGISTRY}/ruby:latest
	sudo buildah pull docker-daemon:${REGISTRY}/packagist:latest
	sudo buildah pull docker-daemon:${REGISTRY}/crates.io:latest


check_scripts:
	find -type f -name '*.sh' | xargs --no-run-if-empty shellcheck -S warning

run:
	@echo "To perform a local analysis, please use ./run_analysis.sh"
	@echo


