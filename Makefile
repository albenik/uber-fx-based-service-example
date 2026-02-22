.PHONY: build run test test-core vet tidy generate proto-generate migrate-up migrate-down migrate-status install-goose install-govulncheck clean help show-outdated update-outdated upgrade-dependencies audit lint

.DEFAULT_GOAL := help

BINARY := server

build: ## Build server binary to bin/
	go build -o bin/$(BINARY) ./cmd/server

run: ## Run server (requires DATABASE_MASTER_URL)
	go run ./cmd/server

test: ## Run all tests
	go test ./...

vet: ## Run static analysis
	go vet ./...

lint: ## Run golangci-lint with default config
	golangci-lint run

tidy: ## Clean up go.mod dependencies
	go mod tidy

DIRECT_DEPS := $(shell awk '/^require \($$/ {p=1; next} p && /^\)$$/ {exit} p && /^\t/ && !/\/\/ indirect/ {print $$1}' go.mod | tr '\n' ' ')

show-outdated: ## Show outdated direct dependencies
	go list -u -m $(DIRECT_DEPS)

update-outdated: ## Update outdated direct dependencies to latest minor versions
	go get -u $(DIRECT_DEPS)
	go mod tidy

upgrade-dependencies: ## Update outdated direct dependencies to latest versions
	go get $(foreach m,$(DIRECT_DEPS), $m@latest)
	go mod tidy

audit: ## Run dependencies security scan (govulncheck)
	go tool govulncheck ./...

generate: ## Regenerate gomock mocks
	go generate ./...

proto-generate: ## Generate Go code from protobuf definitions (requires buf, protoc-gen-go, protoc-gen-go-grpc)
	@PATH="$$PATH:$$(go env GOPATH)/bin" BUF_CACHE_DIR="$$(pwd)/.buf/cache" buf generate

migrate-status: ## Show migration status (requires DATABASE_MASTER_URL)
	@if [ -z "$$DATABASE_MASTER_URL" ]; then echo "DATABASE_MASTER_URL is required"; exit 1; fi
	go tool goose -dir migrations postgres "$$DATABASE_MASTER_URL" status

migrate-up: ## Apply migrations (requires DATABASE_MASTER_URL)
	@if [ -z "$$DATABASE_MASTER_URL" ]; then echo "DATABASE_MASTER_URL is required"; exit 1; fi
	go tool goose -dir migrations postgres "$$DATABASE_MASTER_URL" up

migrate-down: ## Rollback migrations (requires DATABASE_MASTER_URL)
	@if [ -z "$$DATABASE_MASTER_URL" ]; then echo "DATABASE_MASTER_URL is required"; exit 1; fi
	go tool goose -dir migrations postgres "$$DATABASE_MASTER_URL" down

clean: ## Remove build artifacts
	rm -rf bin/
	rm -f server

help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
