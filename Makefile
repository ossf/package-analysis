# This Makefile contains common development / build commands for Package Analysis. For everything to work properly, it needs to be kept in the top-level project directory.

# Get absolute path to top-level package analysis project directory
# outermost abspath removes the trailing slash from the directory path
PREFIX := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
SANDBOX_DIR := $(PREFIX)/sandboxes

# Registry for Docker images built and used by package analysis
REGISTRY := gcr.io/ossf-malware-analysis

# If RELEASE_TAG environment variable is not specified, images will be tagged
# as 'latest' which is equivalent to just tagging without specifying a version
TAG := ${RELEASE_TAG}
ifeq ($(TAG), )
	TAG := latest
	BUILD_ARG=
else
	# pass tag into analysis image build
	BUILD_ARG=--build-arg=SANDBOX_IMAGE_TAG=$(TAG)
endif

# whether to push docker images
DOCKER_PUSH := ${PUSH}
ifeq ($(DOCKER_PUSH), )
	DOCKER_PUSH := false
endif


.PHONY: all
all: docker_build_all

#
# These recipes build all the top-level docker images

docker_build_%_image:
	@# if TAG is 'latest', the two -t arguments are equivalent and do the same thing
	docker build $(BUILD_ARG) -t ${REGISTRY}/$(IMAGE_NAME) -t ${REGISTRY}/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) $(DIR)
	if [[ "$(PUSH)" == "true" ]]; then docker push --all-tags ${REGISTRY}/$(IMAGE_NAME); fi

#
# These build the sandbox images and also update (sync) them locally
# from Docker to podman. This is needed for local analyses; in order
# to use these updated images, pass 'nopull' to run_analysis.sh
#
docker_build_%_sandbox:
	@# if TAG is 'latest', the two -t arguments are equivalent and do the same thing
	docker build -t ${REGISTRY}/$(IMAGE_NAME) -t ${REGISTRY}/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) $(DIR)
	if [[ "$(PUSH)" == "true" ]]; then docker push --all-tags ${REGISTRY}/$(IMAGE_NAME); fi

	sudo buildah pull docker-daemon:${REGISTRY}/${IMAGE_NAME}:$(TAG)

docker_build_analysis_image: DIR=$(PREFIX)
docker_build_analysis_image: DOCKERFILE=$(PREFIX)/cmd/analyze/Dockerfile
docker_build_analysis_image: IMAGE_NAME=analysis

docker_build_scheduler_image: DIR=$(PREFIX)
docker_build_scheduler_image: DOCKERFILE=$(PREFIX)/cmd/scheduler/Dockerfile
docker_build_scheduler_image: IMAGE_NAME=scheduler

docker_build_node_sandbox: DIR=$(SANDBOX_DIR)/npm
docker_build_node_sandbox: DOCKERFILE=$(SANDBOX_DIR)/npm/Dockerfile
docker_build_node_sandbox: IMAGE_NAME=node

docker_build_python_sandbox: DIR=$(SANDBOX_DIR)/pypi
docker_build_python_sandbox: DOCKERFILE=$(SANDBOX_DIR)/pypi/Dockerfile
docker_build_python_sandbox: IMAGE_NAME=python

docker_build_ruby_sandbox: DIR=$(SANDBOX_DIR)/rubygems
docker_build_ruby_sandbox: DOCKERFILE=$(SANDBOX_DIR)/rubygems/Dockerfile
docker_build_ruby_sandbox: IMAGE_NAME=ruby

docker_build_packagist_sandbox: DIR=$(SANDBOX_DIR)/packagist
docker_build_packagist_sandbox: DOCKERFILE=$(SANDBOX_DIR)/packagist/Dockerfile
docker_build_packagist_sandbox:	IMAGE_NAME=packagist

docker_build_crates_sandbox: DIR=$(SANDBOX_DIR)/crates.io
docker_build_crates_sandbox: DOCKERFILE=$(SANDBOX_DIR)/crates.io/Dockerfile
docker_build_crates_sandbox: IMAGE_NAME=crates.io

docker_build_static_analysis_sandbox: DIR=$(PREFIX)
docker_build_static_analysis_sandbox: DOCKERFILE=$(SANDBOX_DIR)/staticanalysis/Dockerfile
docker_build_static_analysis_sandbox: IMAGE_NAME=static-analysis

.PHONY: docker_build_all_sandboxes
docker_build_all_sandboxes: docker_build_node_sandbox docker_build_python_sandbox docker_build_ruby_sandbox docker_build_packagist_sandbox docker_build_crates_sandbox docker_build_static_analysis_sandbox

.PHONY: docker_build_all
docker_build_all: docker_build_all_sandboxes docker_build_analysis_image docker_build_scheduler_image


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


.PHONY: test
test:
	go test -v ./...

