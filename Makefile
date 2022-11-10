REGISTRY=gcr.io/ossf-malware-analysis

.PHONY: all all_docker analysis_docker check_scripts run

all: all_docker

# This builds all sandbox images as well as the analysis container image
all_docker:
	bash build/build_docker.sh

# This builds just the analysis container image
analysis_docker:
	docker build -t ${REGISTRY}/analysis -f cmd/analyze/Dockerfile .

check_scripts:
	find -type f -name '*.sh' | xargs --no-run-if-empty shellcheck -S warning

run:
	@echo "To perform a local analysis, please use ./run_analysis.sh"
	@echo


