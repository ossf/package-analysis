
REGISTRY=gcr.io/ossf-malware-analysis

.PHONY: all all_docker analysis_docker

all: all_docker

all_docker:
	bash build/build_docker.sh

analysis_docker:
	docker build -t ${REGISTRY}/analysis -f cmd/analyze/Dockerfile .


run:
	@echo "To run analysis locally, please use the ./run_analysis.sh script"
	@echo
