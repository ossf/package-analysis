REGISTRY=gcr.io/ossf-malware-analysis

.PHONY: all all_docker analysis_docker check_scripts run static_analysis_sandbox

all: all_docker

# This builds all sandbox images except for static analysis, as well as the analysis container image
all_docker:
	bash build/build_docker.sh

# This builds just the analysis container image
analysis_docker:
	docker build -t ${REGISTRY}/analysis -f cmd/analyze/Dockerfile .

# This builds just the static analysis container image
static_analysis_sandbox:
	docker build -t ${REGISTRY}/static-analysis -f sandboxes/staticanalysis/Dockerfile sandboxes/staticanalysis

check_scripts:
	find -type f -name '*.sh' | xargs --no-run-if-empty shellcheck -S warning

run:
	@echo "To perform a local analysis, please use ./run_analysis.sh"
	@echo

docker_compose_start:
	cd ./examples/e2e && docker-compose up -d
	sleep 3
	curl localhost:8080
	@echo
	@echo "To see analysis results, go to http://localhost:9000/minio/package-analysis"
	@echo "Remember to run "make docker_compose_stop" when done!"

docker_compose_logs:
	cd ./examples/e2e && docker-compose logs

docker_compose_stop:
	cd ./examples/e2e && docker-compose down
