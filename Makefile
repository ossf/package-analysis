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


#
# This recipe builds and pushes images for production. Note: RELEASE_TAG must be set
#
.PHONY: cloudbuild
cloudbuild: require_release_tag push_prod_images

.PHONY: require_release_tag
require_release_tag:
ifndef RELEASE_TAG
	$(error RELEASE_TAG must be set)
endif


#
# These recipes build all the top-level docker images

build_%_image:
	@# if TAG is 'latest', the two -t arguments are equivalent and do the same thing
	docker build $(BUILD_ARG) -t ${REGISTRY}/$(IMAGE_NAME) -t ${REGISTRY}/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) $(DIR)

#
# These recipes build the sandbox images.
#
build_%_sandbox:
	@# if TAG is 'latest', the two -t arguments are equivalent and do the same thing
	docker build -t ${REGISTRY}/$(IMAGE_NAME) -t ${REGISTRY}/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) $(DIR)

build_analysis_image: DIR=$(PREFIX)
build_analysis_image: DOCKERFILE=$(PREFIX)/cmd/analyze/Dockerfile
build_analysis_image: IMAGE_NAME=analysis

build_scheduler_image: DIR=$(PREFIX)
build_scheduler_image: DOCKERFILE=$(PREFIX)/cmd/scheduler/Dockerfile
build_scheduler_image: IMAGE_NAME=scheduler

build_static_analysis_sandbox: DIR=$(PREFIX)
build_static_analysis_sandbox: DOCKERFILE=$(SANDBOX_DIR)/staticanalysis/Dockerfile
build_static_analysis_sandbox: IMAGE_NAME=static-analysis

build_dynamic_analysis_sandbox: DIR=$(SANDBOX_DIR)/dynamicanalysis
build_dynamic_analysis_sandbox: DOCKERFILE=$(SANDBOX_DIR)/dynamicanalysis/Dockerfile
build_dynamic_analysis_sandbox: IMAGE_NAME=dynamic-analysis

.PHONY: build_prod_images
build_prod_images: build_dynamic_analysis_sandbox build_static_analysis_sandbox build_analysis_image build_scheduler_image

#
# Builds then pushes analysis and sandbox images
#

push_%:
	docker push --all-tags ${REGISTRY}/$(IMAGE_NAME)

push_analysis_image: IMAGE_NAME=analysis
push_analysis_image: build_analysis_image

push_scheduler_image: IMAGE_NAME=scheduler
push_scheduler_image: build_scheduler_image

push_dynamic_analysis_sandbox: IMAGE_NAME=dynamic-analysis
push_dynamic_analysis_sandbox: build_dynamic_analysis_sandbox

push_static_analysis_sandbox: IMAGE_NAME=static-analysis
push_static_analysis_sandbox: build_static_analysis_sandbox

.PHONY: push_prod_sandboxes
push_prod_sandboxes: push_dynamic_analysis_sandbox push_static_analysis_sandbox

.PHONY: push_prod_images
push_prod_images: push_prod_sandboxes push_analysis_image push_scheduler_image


#
# These update (sync) locally built sandbox images from Docker to
# podman. In order to use locally built sandbox images for analysis,
# pass '-nopull' to scripts/run_analysis.sh
#
sync_%_sandbox:
	sudo buildah pull docker-daemon:${REGISTRY}/${IMAGE_NAME}:$(TAG)

sync_dynamic_analysis_sandbox: IMAGE_NAME=dynamic-analysis
sync_dynamic_analysis_sandbox: build_dynamic_analysis_sandbox

sync_static_analysis_sandbox: IMAGE_NAME=static-analysis
sync_static_analysis_sandbox: build_static_analysis_sandbox

.PHONY: sync_prod_sandboxes
sync_prod_sandboxes: sync_dynamic_analysis_sandbox sync_static_analysis_sandbox


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

E2E_TEST_COMPOSE_ARGS := -p pa-e2e-testing -f ./configs/e2e/docker-compose.yml -f ./test/e2e/docker-compose.test.yml

.PHONY: e2e_test_start
e2e_test_start:
	docker-compose $(E2E_TEST_COMPOSE_ARGS) up -d
	@echo
	@echo "To see analysis results, go to http://localhost:9000/minio/package-analysis"
	@echo
	@echo "Remember to run 'make e2e_test_stop' when done!"
	@sleep 5
	@echo
	curl localhost:8080

.PHONY: e2e_test_stop
e2e_test_stop:
	docker-compose $(E2E_TEST_COMPOSE_ARGS) down

.PHONY: e2e_test_logs_all
e2e_test_logs_all:
	docker-compose $(E2E_TEST_COMPOSE_ARGS) logs

.PHONY: e2e_test_logs_feeds
e2e_test_logs_feeds:
	docker-compose $(E2E_TEST_COMPOSE_ARGS) logs -f feeds

.PHONY: e2e_test_logs_scheduler
e2e_test_logs_scheduler:
	docker-compose $(E2E_TEST_COMPOSE_ARGS) logs -f scheduler

.PHONY: e2e_test_logs_analysis
e2e_test_logs_analysis:
	docker-compose $(E2E_TEST_COMPOSE_ARGS) logs -f analysis

.PHONY: test_go
test_go:
	go test -v ./...

.PHONY: test_dynamic_analysis
test_dynamic_analysis:
	@echo -e "\n##\n## Test NPM \n##\n"
	scripts/run_analysis.sh -mode dynamic -nopull -ecosystem npm -package async
	@echo -e "\n##\n## Test PyPI \n##\n"
	scripts/run_analysis.sh -mode dynamic -nopull -ecosystem pypi -package requests
	@echo -e "\n##\n## Test Packagist \n##\n"
	scripts/run_analysis.sh -mode dynamic -nopull -ecosystem packagist -package symfony/deprecation-contracts
	@echo -e "\n##\n## Test Crates.io \n##\n"
	scripts/run_analysis.sh -mode dynamic -nopull -ecosystem crates.io -package itoa
	@echo -e "\n##\n## Test RubyGems \n##\n"
	scripts/run_analysis.sh -mode dynamic -nopull -ecosystem rubygems -package guwor_palindrome
	@echo "Dynamic analysis test passed"

.PHONY: test_static_analysis
test_static_analysis:
	@echo -e "\n##\n## Test NPM \n##\n"
	scripts/run_analysis.sh -mode static -nopull -ecosystem npm -package async
	@echo "Static analysis test passed"
