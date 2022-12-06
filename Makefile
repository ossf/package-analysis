# This Makefile contains common development / build commands for Package Analysis. For everything to work properly, it needs to be kept in the top-level project directory.

# outermost abspath removes the trailing slash from the directory path
PREFIX := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
SANDBOX_DIR := $(PREFIX)/sandboxes

REGISTRY := gcr.io/ossf-malware-analysis
NODE_IMAGE_NAME := node
PYTHON_IMAGE_NAME := python
RUBY_IMAGE_NAME := ruby
PACKAGIST_IMAGE_NAME := packagist
CRATES_IMAGE_NAME := crates.io
STATIC_ANALYSIS_IMAGE_NAME := static-analysis

#
# This is just the old 'build everything script'
#
.PHONY: legacy_build_docker
legacy_build_docker:
	bash build/build_docker.sh

.PHONY: all
all: docker_build_all

#
# These recipes build all the docker images via a common
# pattern recipe 'docker_build_%'
# TODO grab release tag from env vars

docker_build_%:
	docker build -t ${REGISTRY}/$(IMAGE_NAME) -f $(DOCKERFILE) $(DIR)

docker_build_analysis_image: DIR=$(PREFIX)
docker_build_analysis_image: DOCKERFILE=$(PREFIX)/cmd/analyze/Dockerfile
docker_build_analysis_image: IMAGE_NAME=analysis

docker_build_scheduler_image: DIR=$(PREFIX)
docker_build_scheduler_image: DOCKERFILE=$(PREFIX)/cmd/scheduler/Dockerfile
docker_build_scheduler_image: IMAGE_NAME=scheduler

docker_build_node_sandbox: DIR=$(SANDBOX_DIR)/npm
docker_build_node_sandbox: DOCKERFILE=$(SANDBOX_DIR)/npm/Dockerfile
docker_build_node_sandbox: IMAGE_NAME=$(NODE_IMAGE_NAME)

docker_build_python_sandbox: DIR=$(SANDBOX_DIR)/pypi
docker_build_python_sandbox: DOCKERFILE=$(SANDBOX_DIR)/pypi/Dockerfile
docker_build_python_sandbox: IMAGE_NAME=$(PYTHON_IMAGE_NAME)

docker_build_ruby_sandbox: DIR=$(SANDBOX_DIR)/rubygems
docker_build_ruby_sandbox: DOCKERFILE=$(SANDBOX_DIR)/rubygems/Dockerfile
docker_build_ruby_sandbox: IMAGE_NAME=$(RUBY_IMAGE_NAME)

docker_build_packagist_sandbox: DIR=$(SANDBOX_DIR)/packagist
docker_build_packagist_sandbox: DOCKERFILE=$(SANDBOX_DIR)/packagist/Dockerfile
docker_build_packagist_sandbox:	IMAGE_NAME=$(PACKAGIST_IMAGE_NAME)

docker_build_crates_sandbox: DIR=$(SANDBOX_DIR)/crates.io
docker_build_crates_sandbox: DOCKERFILE=$(SANDBOX_DIR)/crates.io/Dockerfile
docker_build_crates_sandbox: IMAGE_NAME=$(CRATES_IMAGE_NAME)

docker_build_static_analysis_sandbox: DIR=$(PREFIX)
docker_build_static_analysis_sandbox: DOCKERFILE=$(SANDBOX_DIR)/staticanalysis/Dockerfile
docker_build_static_analysis_sandbox: IMAGE_NAME=$(STATIC_ANALYSIS_IMAGE_NAME)

docker_build_all_sandboxes: docker_build_node_sandbox docker_build_python_sandbox docker_build_ruby_sandbox docker_build_packagist_sandbox docker_build_crates_sandbox docker_build_static_analysis_sandbox

docker_build_all: docker_build_all_sandboxes docker_build_analysis_image docker_build_scheduler_image

#
# These recipes update (sync) sandbox images built locally by Docker
# to podman, which is what actually runs them. This is needed for
# local analyses; in order to use these updated images, pass
# 'nopull' to run_analysis.sh
#

sync_%_sandbox:
	sudo buildah pull docker-daemon:${REGISTRY}/${IMAGE_NAME}:latest

sync_node_sandbox: IMAGE_NAME=${NODE_IMAGE_NAME}

sync_python_sandbox: IMAGE_NAME=${PYTHON_IMAGE_NAME}

sync_ruby_sandbox: IMAGE_NAME=${RUBY_IMAGE_NAME}

sync_packagist_sandbox: IMAGE_NAME=${PACKAGIST_IMAGE_NAME}

sync_crates_sandbox: IMAGE_NAME=${CRATES_IMAGE_NAME}

sync_sandbox_static_analysis: IMAGE_NAME=${STATIC_ANALYSIS_IMAGE_NAME}

sync_all_sandboxes: sync_node_sandbox sync_python_sandbox sync_ruby_sandbox sync_packagist_sandbox sync_crates_sandbox sync_static_analysis_sandbox

#
# This runs a lint check on all shell scripts in the repo
#
.PHONY: check_scripts
check_scripts:
	find -type f -name '*.sh' | xargs --no-run-if-empty shellcheck -S warning

.PHONY: run
run:
	@echo "To perform a local analysis, please use ./run_analysis.sh"
	@echo

#
# These recipes control docker-compose, which is used for
# end-to-end testing of the complete scheduler/worker system
#
.PHONY: docker_compose_start
docker_compose_start:
	cd ./examples/e2e && docker-compose up -d
	sleep 3
	curl localhost:8080
	@echo
	@echo "To see analysis results, go to http://localhost:9000/minio/package-analysis"
	@echo "Remember to run `make docker_compose_stop` when done!"

.PHONY: docker_compose_logs
docker_compose_logs:
	cd ./examples/e2e && docker-compose logs

.PHONY: docker_compose_stop
docker_compose_stop:
	cd ./examples/e2e && docker-compose down

