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

.PHONY: all
all: docker_build_all

#
# These recipes build all the top-level docker images

docker_build_%_image:
	@# if TAG is 'latest', the two -t arguments are equivalent and do the same thing
	docker build $(BUILD_ARG) -t ${REGISTRY}/$(IMAGE_NAME) -t ${REGISTRY}/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) $(DIR)

#
# These build the sandbox images and also update (sync) them locally
# from Docker to podman. This is needed for local analyses; in order
# to use these updated images, pass 'nopull' to run_analysis.sh
#
docker_build_%_sandbox:
	@# if TAG is 'latest', the two -t arguments are equivalent and do the same thing
	docker build -t ${REGISTRY}/$(IMAGE_NAME) -t ${REGISTRY}/$(IMAGE_NAME):$(TAG) -f $(DOCKERFILE) $(DIR)
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
# Builds then pushes analysis and sandbox images
#

docker_push_%:
	docker push --all-tags ${REGISTRY}/$(IMAGE_NAME)

docker_push_analysis_image: docker_build_analysis_image
docker_push_scheduler_image: docker_build_scheduler_image
docker_push_node_sandbox: docker_build_node_sandbox
docker_push_python_sandbox: docker_build_python_sandbox
docker_push_ruby_sandbox: docker_build_ruby_sandbox
docker_push_packagist_sandbox: docker_build_packagist_sandbox
docker_push_crates_sandbox: docker_build_crates_sandbox
docker_push_static_analysis_sandbox: docker_build_static_analysis_sandbox

docker_push_analysis_image: IMAGE_NAME=analysis
docker_push_scheduler_image: IMAGE_NAME=scheduler
docker_push_node_sandbox: IMAGE_NAME=node
docker_push_python_sandbox: IMAGE_NAME=python
docker_push_ruby_sandbox: IMAGE_NAME=ruby
docker_push_packagist_sandbox:	IMAGE_NAME=packagist
docker_push_crates_sandbox: IMAGE_NAME=crates.io
docker_push_static_analysis_sandbox: IMAGE_NAME=static-analysis

.PHONY: docker_push_all_sandboxes
docker_push_all_sandboxes: docker_push_node_sandbox docker_push_python_sandbox docker_push_ruby_sandbox docker_push_packagist_sandbox docker_push_crates_sandbox docker_push_static_analysis_sandbox

.PHONY: docker_push_all
docker_push_all: docker_push_all_sandboxes docker_push_analysis_image docker_push_scheduler_image


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
E2E_TEST_DIR := ./test/e2e

.PHONY: e2e_test_start
e2e_test_start:
	cd $(E2E_TEST_DIR) && docker-compose up -d
	@echo
	@echo "To see analysis results, go to http://localhost:9000/minio/package-analysis"
	@echo "Remember to run 'make e2e_test_stop' when done!"
	sleep 5
	curl localhost:8080

.PHONY: e2e_test_stop
e2e_test_stop:
	cd $(E2E_TEST_DIR) && docker-compose down

.PHONY: e2e_test_logs_all
e2e_test_logs_all:
	cd $(E2E_TEST_DIR) && docker-compose logs

.PHONY: e2e_test_logs_feeds
e2e_test_logs_feeds:
	cd $(E2E_TEST_DIR) && docker-compose logs -f feeds

.PHONY: e2e_test_logs_scheduler
e2e_test_logs_scheduler:
	cd $(E2E_TEST_DIR) && docker-compose logs -f scheduler

.PHONY: e2e_test_logs_analysis:
e2e_test_logs_analysis:
	cd $(E2E_TEST_DIR) && docker-compose logs -f analysis


.PHONY: test
test:
	go test -v ./...

