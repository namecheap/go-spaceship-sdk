SHELL := /bin/bash

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Compile all packages
	go build ./...

.PHONY: test
test: ## Run unit tests (excludes live/acceptance tests)
	go test -skip 'TestAcc' ./...

.PHONY: test-cover
test-cover: ## Run unit tests with a coverage report
	go test -skip 'TestAcc' -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

.PHONY: testacc
testacc: ## Run live tests against the real Spaceship API (loads .env if present)
	@set -a; [[ -f .env ]] && source .env; set +a; \
	for v in SPACESHIP_API_KEY SPACESHIP_API_SECRET; do \
		if [[ -z "$${!v}" ]]; then \
			echo "$$v must be set — export it or add it to .env"; exit 1; \
		fi; \
	done; \
	go test -run TestAcc ./... -v -count=1 -timeout=10m

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run linters and the formatter check
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Auto-fix formatting
	golangci-lint fmt ./...
