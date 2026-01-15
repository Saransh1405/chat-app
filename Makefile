.PHONY: help build run test clean docker-up docker-down migrate migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/chat-app ./cmd/server

run: ## Run the application
	@echo "Running application..."
	@go run ./cmd/server

test: ## Run tests
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-up: ## Start Docker containers (PostgreSQL, Redis)
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs: ## View Docker container logs
	@docker-compose logs -f

migrate: migrate-up ## Alias for migrate-up

migrate-up: ## Run database migrations up
	@echo "Running database migrations..."
	@psql -h localhost -U chat_app_user -d chat_app_db -f internal/database/migrations/001_initial_schema.sql || echo "Migration already applied or error occurred. Check database connection."

migrate-down: ## Rollback database migrations (manual - implement if needed)
	@echo "Rollback migrations - manual operation required"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run ./... || echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

dev: docker-up run ## Start Docker containers and run application

.DEFAULT_GOAL := help

