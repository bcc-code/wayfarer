.PHONY: help generate migrate seed test build clean

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

clear-db: ## Clear all database tables
	@echo "Clearing database..."
	@go run ./cmd/clear

seed: ## Seed database with test data
	@echo "Seeding database..."
	@go run ./cmd/seed

seed-reproducible: ## Seed database with reproducible data (seed=42)
	@echo "Seeding database with seed 42..."
	@go run ./cmd/seed -seed 42

seed-large: ## Seed database with large dataset for performance testing
	@echo "Seeding database with large dataset (10k users, 50 projects)..."
	@go run ./cmd/seed -users 10000 -projects 50 -churches 100 -superteams 5 -achievements 50 -team-size 8

test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage report
	@go tool cover -html=coverage.out

build: ## Build all binaries
	@echo "Building binaries..."
	@go build -o bin/migrate ./cmd/migrate
	@go build -o bin/server ./cmd/server
	@go build -o bin/seed ./cmd/seed
	@go build -o bin/clear ./cmd/clear
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

## Observability Commands

jaeger-up: ## Start Jaeger tracing
	@echo "Starting Jaeger..."
	@cd ../docker/jaeger && docker compose up -d
	@echo "Jaeger UI available at: http://localhost:16686"

jaeger-down: ## Stop Jaeger tracing
	@echo "Stopping Jaeger..."
	@cd ../docker/jaeger && docker compose down

jaeger-restart: ## Restart Jaeger (clears in-memory traces)
	@echo "Restarting Jaeger..."
	@cd ../docker/jaeger && docker compose restart

jaeger-logs: ## View Jaeger logs
	@cd ../docker/jaeger && docker compose logs -f

jaeger-ps: ## Show Jaeger container status
	@cd ../docker/jaeger && docker compose ps

## Load Testing Commands

loadtest-gen: ## Generate JWT tokens for load testing (run after seed-large)
	@echo "Generating load test tokens..."
	@go run ./cmd/loadtest/tokengen -limit 1000 -output ./cmd/loadtest/config.json

loadtest-gen-all: ## Generate tokens for all seeded users (up to 10k)
	@echo "Generating load test tokens for all users..."
	@go run ./cmd/loadtest/tokengen -limit 10000 -output ./cmd/loadtest/config.json

loadtest: loadtest-gen ## Run default load test (50 VUs, 5 minutes)
	@echo "Running load test..."
	@k6 run ./cmd/loadtest/k6/steady.js

loadtest-quick: loadtest-gen ## Run quick load test (10 VUs, 1 minute)
	@echo "Running quick load test..."
	@k6 run --env STEADY_VUS=10 --env DURATION=1m ./cmd/loadtest/k6/steady.js

loadtest-stress: loadtest-gen-all ## Run stress test (200 VUs, 10 minutes)
	@echo "Running stress test..."
	@k6 run --env STEADY_VUS=200 --env DURATION=10m ./cmd/loadtest/k6/steady.js

loadtest-spike: loadtest-gen ## Run spike test only (500 VUs)
	@echo "Running spike test..."
	@k6 run --env SPIKE_PEAK=500 --env SPIKE_HOLD=2m ./cmd/loadtest/k6/spike.js

loadtest-leaderboard: loadtest-gen ## Run leaderboard stress test only (100 RPS)
	@echo "Running leaderboard stress test..."
	@k6 run --env LEADERBOARD_RPS=100 --env LEADERBOARD_DURATION=5m ./cmd/loadtest/k6/leaderboard.js
