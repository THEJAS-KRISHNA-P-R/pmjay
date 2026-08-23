# PMJAY Point-of-Denial Advocate — backend
#
# Zero third-party Go dependencies (see ../ARCHITECTURE.md for why), so
# every target here is plain `go` tooling — no external build system,
# no Makefile-generator, nothing to install beyond Go itself.

BINARY       := pmjay-advocate-server
PKG          := ./cmd/server
LDFLAGS      := -s -w
DATA_FILE    ?= ./data/cases.json

.PHONY: help build build-arm64 run test test-race vet lint fmt clean docker-build

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*## ' $(MAKEFILE_LIST) | sed 's/:.*## /|/' | sort | awk -F'|' '{printf "  %-14s %s\n", $$1, $$2}'

build: ## Build a native binary into ./bin
	@mkdir -p bin
	go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY) $(PKG)
	@echo "built bin/$(BINARY)"

build-arm64: ## Cross-compile a stripped linux/arm64 binary (t4g.micro / Graviton target) — no C toolchain needed
	@mkdir -p bin
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY)-linux-arm64 $(PKG)
	@echo "built bin/$(BINARY)-linux-arm64 ($$(du -h bin/$(BINARY)-linux-arm64 | cut -f1))"

build-amd64: ## Cross-compile a stripped linux/amd64 binary (generic x86_64 VM target)
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY)-linux-amd64 $(PKG)
	@echo "built bin/$(BINARY)-linux-amd64 ($$(du -h bin/$(BINARY)-linux-amd64 | cut -f1))"

run: ## Run the server locally (reads .env if present via `make run ENV_FILE=.env`)
	@if [ -f "$(ENV_FILE)" ]; then set -a; . ./$(ENV_FILE); set +a; fi; \
	DATA_FILE_PATH=$${DATA_FILE_PATH:-$(DATA_FILE)} go run $(PKG)

test: ## Run the full test suite with coverage
	go test ./... -cover

test-race: ## Run the full test suite with coverage AND the race detector (what CI should run)
	go test ./... -cover -race

test-verbose: ## Run tests with per-test pass/fail output
	go test ./... -v

vet: ## Run go vet
	go vet ./...

fmt: ## Check formatting (fails if any file needs gofmt)
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt-formatted:"; echo "$$unformatted"; exit 1; \
	fi
	@echo "all files gofmt-clean"

check: vet fmt test-race ## Run everything CI should run before a merge: vet + fmt + race tests

clean: ## Remove build artifacts
	rm -rf bin/

docker-build: ## Build the production Docker image (see Dockerfile)
	docker build -t pmjay-advocate-backend:local .
