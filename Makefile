# Variables
BINARY_NAME=vetbot
DOCKER_IMAGE=vetbot/app
MIGRATIONS_DIR=./migrations
TEST_DB_URL=postgres://vetbot_user:vetbot_password@localhost:5433/vetbot_test?sslmode=disable

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

test: ## Run unit tests
	@echo "Running unit tests..."
	go test ./... -short

test-integration: test-db-up ## Run integration tests (requires test database)
	@echo "Running integration tests..."
	@echo "Waiting for test database to be ready..."
	@sleep 3
	TEST_DATABASE_URL=$(TEST_DB_URL) go test ./internal/database/ -v
	@$(MAKE) test-db-down

test-all: ## Run all tests (unit + integration)
	@echo "Running all tests..."
	@$(MAKE) test
	@$(MAKE) test-integration

test-db-up: ## Start test database
	@echo "Starting test database..."
	docker-compose -f docker-compose.test.yml up -d
	@echo "Test database started on port 5433"

test-db-down: ## Stop test database
	@echo "Stopping test database..."
	docker-compose -f docker-compose.test.yml down
	@echo "Test database stopped"

test-db-logs: ## Show test database logs
	docker-compose -f docker-compose.test.yml logs -f

test-db-reset: ## Reset test database (stop, remove volumes, start)
	@echo "Resetting test database..."
	docker-compose -f docker-compose.test.yml down -v
	docker-compose -f docker-compose.test.yml up -d
	@sleep 5
	@echo "Test database reset complete"

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

migrate-test: test-db-up ## Run migrations on test database
	@echo "Running migrations on test database..."
	@sleep 3
	@for file in $(shell ls $(MIGRATIONS_DIR)/*.sql | sort); do \
		echo "Applying $$(basename $$file) to test database..."; \
		psql $(TEST_DB_URL) -f $$file || exit 1; \
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

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: help build run test test-integration test-all test-db-up test-db-down test-db-logs test-db-reset clean migrate-up migrate-test migrate-down docker-build docker-run compose-up compose-down compose-logs dev lint fmt coverage