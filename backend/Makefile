.PHONY: help generate generate-user generate-admin generate-m2m migrate test build clean

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

generate: generate-user generate-admin generate-m2m ## Generate all GraphQL code for all three APIs

generate-user: ## Generate User API GraphQL code
	@echo "Generating User API..."
	@go run github.com/99designs/gqlgen generate --config gqlgen.yml

generate-admin: ## Generate Admin API GraphQL code
	@echo "Generating Admin API..."
	@go run github.com/99designs/gqlgen generate --config gqlgen.admin.yml

generate-m2m: ## Generate M2M API GraphQL code
	@echo "Generating M2M API..."
	@go run github.com/99designs/gqlgen generate --config gqlgen.m2m.yml

migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run ./cmd/migrate -direction up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@go run ./cmd/migrate -direction down

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	@go run ./cmd/migrate -direction status

test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage report
	@go tool cover -html=coverage.out

build: ## Build all binaries
	@echo "Building binaries..."
	@go build -o bin/migrate ./cmd/migrate
	@go build -o bin/server ./cmd/server

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

tidy: ## Tidy Go modules
	@echo "Tidying modules..."
	@go mod tidy

dev: ## Run development server
	@echo "Starting development server..."
	@go run ./cmd/server
