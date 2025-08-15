# RiskMatrix Makefile
# Security detection management platform

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Project info
BINARY_NAME=server
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
MAIN_PATH=./cmd/server

# Database
DB_PATH=data/riskmatrix.db
TEST_DB_PATH=data/test_riskmatrix.db

# Docker
DOCKER_IMAGE=riskmatrix
DOCKER_TAG=latest

.PHONY: all build clean test coverage deps run dev help install import-mitre docker-build docker-run docker-stop

# Default target
all: clean deps test build

## Build commands
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=1 $(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) $(MAIN_PATH)

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_WINDOWS) $(MAIN_PATH)

build-all: build build-linux build-windows ## Build for all platforms

## Development commands
run: build ## Build and run the server
	@echo "Starting server..."
	./$(BINARY_NAME)

dev: ## Run in development mode with hot reload (requires air)
	@echo "Starting development server..."
	@if command -v air > /dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

serve: build ## Build and serve on custom port (make serve PORT=9000)
	@echo "Starting server on port $(or $(PORT),8080)..."
	./$(BINARY_NAME) -addr :$(or $(PORT),8080)

## Database commands
import-mitre: ## Import MITRE ATT&CK data
	@echo "Importing MITRE ATT&CK data..."
	CGO_ENABLED=1 $(GOBUILD) -o mitre-importer ./cmd/import-mitre
	./mitre-importer -db $(DB_PATH) -csv data/mitre.csv
	@rm -f mitre-importer

import-mitre-force: ## Force re-import MITRE ATT&CK data (clears existing data)
	@echo "Force importing MITRE ATT&CK data..."
	@read -p "This will delete existing MITRE data. Continue? (y/N): " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		CGO_ENABLED=1 $(GOBUILD) -o mitre-importer ./cmd/import-mitre; \
		sqlite3 $(DB_PATH) "DELETE FROM mitre_techniques;" 2>/dev/null || true; \
		./mitre-importer -db $(DB_PATH) -csv data/mitre.csv; \
		rm -f mitre-importer; \
	else \
		echo "Import cancelled"; \
	fi

seed-db: ## Seed database with test data
	@echo "Seeding database with test data..."
	@if [ -f "scripts/seed-data.sql" ]; then \
		sqlite3 $(DB_PATH) < scripts/seed-data.sql; \
		echo "Database seeded successfully"; \
	else \
		echo "Seed data file not found"; \
	fi

reset-db: ## Reset database (WARNING: This will delete all data)
	@echo "Resetting database..."
	@read -p "Are you sure you want to reset the database? This will delete ALL data (y/N): " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		rm -f $(DB_PATH) $(TEST_DB_PATH); \
		echo "Database reset complete"; \
	else \
		echo "Database reset cancelled"; \
	fi

## Testing commands
test: ## Run tests
	@echo "Running tests..."
	CGO_ENABLED=1 $(GOTEST) -v ./...

test-short: ## Run short tests only
	@echo "Running short tests..."
	CGO_ENABLED=1 $(GOTEST) -short -v ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	CGO_ENABLED=1 $(GOTEST) -v ./internal/integration_test.go

coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	CGO_ENABLED=1 $(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	CGO_ENABLED=1 $(GOTEST) -bench=. -benchmem ./...

## Code quality commands
fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

vet: ## Vet code
	@echo "Vetting code..."
	$(GOCMD) vet ./...

lint: ## Run golangci-lint (requires golangci-lint)
	@if command -v golangci-lint > /dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

check: fmt vet test ## Format, vet, and test

## Dependency commands
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u all

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) verify

## Docker commands
docker-build: ## Build Docker image for current platform
	@echo "Building Docker image for current platform..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-build-arm: ## Build Docker image for ARM64 (Apple Silicon)
	@echo "Building Docker image for ARM64..."
	docker buildx build --platform linux/arm64 -t $(DOCKER_IMAGE):$(DOCKER_TAG)-arm64 .

docker-build-multi: ## Build multi-platform Docker image (AMD64 + ARM64)
	@echo "Building multi-platform Docker image..."
	docker buildx build --platform linux/amd64,linux/arm64 -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-buildx-setup: ## Set up Docker buildx for multi-platform builds
	@echo "Setting up Docker buildx..."
	docker buildx create --name multiplatform --use --platform linux/amd64,linux/arm64
	docker buildx inspect --bootstrap

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	@if [[ "$$(uname -m)" == "arm64" ]]; then \
		TARGETARCH=arm64 docker-compose up -d; \
	else \
		TARGETARCH=amd64 docker-compose up -d; \
	fi

docker-run-arm: ## Run Docker container optimized for ARM64
	@echo "Running Docker container for ARM64..."
	TARGETARCH=arm64 docker-compose up -d

docker-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker resources..."
	docker-compose down --volumes --rmi all

docker-push-multi: docker-build-multi ## Build and push multi-platform image
	@echo "Pushing multi-platform Docker image..."
	docker buildx build --platform linux/amd64,linux/arm64 -t $(DOCKER_IMAGE):$(DOCKER_TAG) --push .

docker-import-mitre: ## Import MITRE data in running Docker container
	@echo "Importing MITRE data in Docker container..."
	docker exec riskmatrix ./mitre-importer -db /app/data/riskmatrix.db -csv /app/data/mitre.csv

docker-shell: ## Open shell in running Docker container
	@echo "Opening shell in Docker container..."
	docker exec -it riskmatrix /bin/sh

## Utility commands
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(BINARY_UNIX) $(BINARY_WINDOWS)
	rm -f mitre-importer coverage.out coverage.html

install: build ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	cp $(BINARY_NAME) $(GOPATH)/bin/

status: ## Show project status
	@echo "=== RiskMatrix Project Status ==="
	@echo "Go version: $$(go version)"
	@echo "Project directory: $$(pwd)"
	@echo "Database exists: $$(if [ -f $(DB_PATH) ]; then echo 'Yes'; else echo 'No'; fi)"
	@echo "Binary exists: $$(if [ -f $(BINARY_NAME) ]; then echo 'Yes'; else echo 'No'; fi)"
	@echo "Git status:"
	@git status --porcelain || echo "Not a git repository"

logs: ## Tail server logs (if running with systemd or similar)
	@echo "Showing recent logs..."
	@if [ -f "/var/log/riskmatrix.log" ]; then \
		tail -f /var/log/riskmatrix.log; \
	else \
		echo "No log file found. Server may not be running or logging to stdout."; \
	fi

## Quick start commands
quickstart: deps import-mitre build run ## Quick start: deps, import data, build, and run

fresh-start: clean deps test import-mitre build ## Fresh start: clean, deps, test, import data, and build

help: ## Show this help message
	@echo "RiskMatrix - Security Detection Management Platform"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make quickstart     # Quick setup and run"
	@echo "  make dev           # Development mode with hot reload"
	@echo "  make test coverage # Run tests with coverage report"
	@echo "  make docker-run    # Run with Docker"
	@echo "  make serve PORT=9000 # Run on custom port"