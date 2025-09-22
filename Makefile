# Variables
BINARY_NAME=vetbot
DOCKER_IMAGE=vetbot/app
MIGRATIONS_DIR=./migrations

.PHONY: help build run test clean migrate-up migrate-down docker-build docker-run

help: ## Show this help
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./cmd/vetbot

run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	go run ./cmd/vetbot

test: ## Run tests
	@echo "Running tests..."
	go test ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@for file in $(shell ls $(MIGRATIONS_DIR)/*.sql | sort); do \
		echo "Applying $$(basename $$file)..."; \
		psql $$DATABASE_URL -f $$file || exit 1; \
	done

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@echo "Warning: This will drop all tables!"
	psql $$DATABASE_URL -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --env-file .env -p 8080:8080 $(DOCKER_IMAGE)

compose-up: ## Start with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

compose-down: ## Stop Docker Compose services
	@echo "Stopping services..."
	docker-compose down

compose-logs: ## Show Docker Compose logs
	docker-compose logs -f

dev: ## Start development environment
	@echo "Starting development environment..."
	docker-compose up -d postgres
	@sleep 5
	@echo "Database is ready, now run: make run"

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...