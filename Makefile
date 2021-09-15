
CC = go
GOOS = linux
GOPKG = github.com/deptofdefense/awslogin

ROOT = $(shell pwd)
BINDIR = $(ROOT)/bin

GIT_COMMIT ?= $(shell git rev-list -1 HEAD)
COMMON_LDFLAGS=-s -w -X $(GOPKG)/pkg/version.commit=$(GIT_COMMIT)
ifdef CIRCLECI
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		LDFLAGS=-linkmode external -extldflags -static
	endif
endif

.PHONY: help
help: ## Print the help documentation
	@grep -E '^[\/a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# ----- CLI Targets -----

bin/awslogin: ## Build awslogin
	GOARCH=amd64 $(CC) build -ldflags "$(LDFLAGS) $(COMMON_LDFLAGS)" -o $@ $(GOPKG)/cmd/$(notdir $@)

# ----- Other Targets -----

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: test
test: ## Tests for the project
	go test ./pkg/... -count=1

.PHONY: test_coverage
test_coverage: ## Tests with coverage
	go test ./pkg/... -cover  -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: clean
clean: ## Clean up built items
	-(rm -rf $(BINDIR))
	-(rm -f coverage.out)
