.PHONY: help generate migrate seed test gqltest build clean

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

generate: ## Generate GraphQL code
	@echo "Generating GraphQL API..."
	@go run github.com/99designs/gqlgen generate --config gqlgen.yml

migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run ./cmd/migrate -direction up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@go run ./cmd/migrate -direction down

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	@go run ./cmd/migrate -direction status

seed: ## Seed database with test data
	@echo "Seeding database..."
	@go run ./cmd/seed

seed-reproducible: ## Seed database with reproducible data (seed=42)
	@echo "Seeding database with seed 42..."
	@go run ./cmd/seed -seed 42

test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage report
	@go tool cover -html=coverage.out

gqltest: ## Run GraphQL tests in parallel
	@echo "Running GraphQL tests..."
	@go run ./cmd/gqltest --parallel testcases/

gqltest-update: ## Update GraphQL test snapshots
	@echo "Updating GraphQL test snapshots..."
	@go run ./cmd/gqltest-update testcases/

build: ## Build all binaries
	@echo "Building binaries..."
	@go build -o bin/migrate ./cmd/migrate
	@go build -o bin/server ./cmd/server
	@go build -o bin/seed ./cmd/seed
	@go build -o bin/debug ./cmd/debug

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

debug: ## Run debug TUI tool
	@echo "Starting debug tool..."
	@go run ./cmd/debug
