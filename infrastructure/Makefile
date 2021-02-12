GO ?= go

BIN := bin
GOOS   ?= $(shell uname | tr A-Z a-z)

GOLANGCI_LINT_VERSION = v1.35.2
$(BIN_DIR):
	mkdir -p bin

.PHONY: help
help:  ## Display this help
	@awk \
		-v "col=${COLOR}" -v "nocol=${NOCOLOR}" \
		' \
			BEGIN { \
				FS = ":.*##" ; \
				printf "Available targets:\n"; \
			} \
			/^[a-zA-Z0-9_-]+:.*?##/ { \
				printf "  %s%-25s%s %s\n", col, $$1, nocol, $$2 \
			} \
			/^##@/ { \
				printf "\n%s%s%s\n", col, substr($$0, 5), nocol \
			} \
		' $(MAKEFILE_LIST)

.PHONY: build
build: $(BIN_DIR)
	pulumi preview

.PHONY: clean
clean: ## Clean the build directory
	rm -rf $(BIN)

.PHONY: go-mod
go-mod: ## Cleanup and verify go modules
	export GO111MODULE=on && \
		$(GO) mod tidy && \
		$(GO) mod verify

# Verification targets
.PHONY: verify
verify: verify-go-mod verify-go-lint ## Run all verification targets

.PHONY: verify-go-mod
verify-go-mod: go-mod ## Verify the go modules
	../hacks/tree-status

.PHONY: verify-go-lint
verify-go-lint: golangci-lint ## Verify the golang code by linting
	$(BIN)/golangci-lint run 

golangci-lint: $(BIN_DIR)
	export \
		VERSION=$(GOLANGCI_LINT_VERSION) \
		URL=https://raw.githubusercontent.com/golangci/golangci-lint \
		BINDIR=$(BIN) && \
	curl -sfL $$URL/$$VERSION/install.sh | sh -s $$VERSION
