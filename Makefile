.DEFAULT_GOAL := help

BINARY := server
DIRECT_DEPS := $(shell awk '/^require \($$/ {p=1; next} p && /^\)$$/ {exit} p && /^\t/ && !/\/\/ indirect/ {print $$1}' go.mod | tr '\n' ' ')

.PHONY: build
build: ## Build server binary to bin/
	go build -o bin/$(BINARY) ./cmd/server

.PHONY: run
run: ## Run server (requires DATABASE_MASTER_URL)
	go run ./cmd/server

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: vet
vet: ## Run static analysis
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint with default config
	golangci-lint run

.PHONY: tidy
tidy: ## Clean up go.mod dependencies
	go mod tidy

.PHONY: show-outdated
show-outdated: ## Show outdated direct dependencies
	go list -u -m $(DIRECT_DEPS)

.PHONY: update-outdated
update-outdated: ## Update outdated direct dependencies to latest minor versions
	go get -u $(DIRECT_DEPS)
	go mod tidy

.PHONY: upgrade-dependencies
upgrade-dependencies: ## Update outdated direct dependencies to latest versions
	go get $(foreach m,$(DIRECT_DEPS), $m@latest)
	go mod tidy

.PHONY: audit
audit: ## Run dependencies security scan (govulncheck)
	go tool govulncheck ./...

.PHONY: generate
generate: ## Regenerate gomock mocks (uses go tool mockgen)
	go generate ./...

.PHONY: install-tools
install-tools: ## Install all tools from go.mod tool directive (goose, mockgen, govulncheck, protoc-gen-go, protoc-gen-go-grpc)
	go install tool

.PHONY: proto-generate
proto-generate: ## Generate Go code from protobuf definitions (requires buf; run make install-tools first)
	@PATH="$$PATH:$$(go env GOPATH)/bin:$$(go env GOBIN)" BUF_CACHE_DIR="$$(pwd)/.buf/cache" buf generate

.PHONY: migrate-status
migrate-status: ## Show migration status (requires DATABASE_MASTER_URL)
	@if [ -z "$$DATABASE_MASTER_URL" ]; then echo "DATABASE_MASTER_URL is required"; exit 1; fi
	go tool goose -dir migrations postgres "$$DATABASE_MASTER_URL" status

.PHONY: migrate-up
migrate-up: ## Apply migrations (requires DATABASE_MASTER_URL)
	@if [ -z "$$DATABASE_MASTER_URL" ]; then echo "DATABASE_MASTER_URL is required"; exit 1; fi
	go tool goose -dir migrations postgres "$$DATABASE_MASTER_URL" up

.PHONY: migrate-down
migrate-down: ## Rollback migrations (requires DATABASE_MASTER_URL)
	@if [ -z "$$DATABASE_MASTER_URL" ]; then echo "DATABASE_MASTER_URL is required"; exit 1; fi
	go tool goose -dir migrations postgres "$$DATABASE_MASTER_URL" down

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/
	rm -f server

.PHONY: help
help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
