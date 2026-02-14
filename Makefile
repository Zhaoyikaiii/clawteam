# ClawTeam Makefile
# Usage: make <target>

.PHONY: help dev test build run clean docker-up docker-down docker-logs deps fmt lint

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME_AGENT=agent-runtime
BINARY_NAME_TASK=task-service
BUILD_DIR=build
CMD_DIR=cmd
DOCKER_COMPOSE=docker-compose
GO=go
GOFLAGS=-v

# Colors for output
COLOR_RESET=\033[0m
COLOR_INFO=\033[36m
COLOR_SUCCESS=\033[32m
COLOR_WARNING=\033[33m
COLOR_ERROR=\033[31m

help: ## Show this help message
	@echo '$(COLOR_SUCCESS)ClawTeam Development Commands$(COLOR_RESET)'
	@echo ''
	@echo '$(COLOR_INFO)Available targets:$(COLOR_RESET)'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_INFO)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'

## ===================================================================
## Development
## ===================================================================

dev: ## Start development environment (services + infrastructure)
	@echo '$(COLOR_INFO)Starting development environment...$(COLOR_RESET)'
	@$(MAKE) docker-up
	@echo '$(COLOR_SUCCESS)✓ Development environment is ready$(COLOR_RESET)'
	@echo ''
	@echo '$(COLOR_INFO)Services:$(COLOR_RESET)'
	@echo '  - PostgreSQL:  localhost:5432'
	@echo '  - NATS:       localhost:4222'
	@echo '  - Qdrant:     localhost:6333'
	@echo '  - Redis:      localhost:6379'
	@echo '  - Jaeger UI:  http://localhost:16686'
	@echo ''
	@echo '$(COLOR_INFO)Demo credentials:$(COLOR_RESET)'
	@echo '  Database: clawteam/clawteam/clawteam_dev_password'
	@echo '  Demo user: demo@clawteam.dev'

deps: ## Download Go dependencies
	@echo '$(COLOR_INFO)Downloading dependencies...$(COLOR_RESET)'
	@$(GO) mod download
	@$(GO) mod tidy
	@echo '$(COLOR_SUCCESS)✓ Dependencies ready$(COLOR_RESET)'

## ===================================================================
## Build
## ===================================================================

build: ## Build all service binaries
	@echo '$(COLOR_INFO)Building services...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)
	@echo '  Building $(BINARY_NAME_AGENT)...'
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_AGENT) ./$(CMD_DIR)/agent-runtime
	@echo '  Building $(BINARY_NAME_TASK)...'
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_TASK) ./$(CMD_DIR)/task-service
	@echo '$(COLOR_SUCCESS)✓ Build complete$(COLOR_RESET)'

build-linux: ## Build binaries for Linux (Docker)
	@echo '$(COLOR_INFO)Building Linux binaries...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_AGENT) ./$(CMD_DIR)/agent-runtime
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_TASK) ./$(CMD_DIR)/task-service
	@echo '$(COLOR_SUCCESS)✓ Linux build complete$(COLOR_RESET)'

## ===================================================================
## Run
## ===================================================================

run-agent: ## Run agent runtime locally
	@echo '$(COLOR_INFO)Starting Agent Runtime...$(COLOR_RESET)'
	@$(GO) run ./$(CMD_DIR)/agent-runtime/main.go

run-task: ## Run task service locally
	@echo '$(COLOR_INFO)Starting Task Service...$(COLOR_RESET)'
	@$(GO) run ./$(CMD_DIR)/task-service/main.go

## ===================================================================
## Test
## ===================================================================

test: ## Run all tests
	@echo '$(COLOR_INFO)Running tests...$(COLOR_RESET)'
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo '$(COLOR_SUCCESS)✓ Tests passed. Coverage report: coverage.html$(COLOR_RESET)'

test-short: ## Run tests without race detector
	@echo '$(COLOR_INFO)Running tests (fast)...$(COLOR_RESET)'
	@$(GO) test -v ./...

test-coverage: ## Show test coverage
	@echo '$(COLOR_INFO)Test coverage:$(COLOR_RESET)'
	@$(GO) test -cover ./...

## ===================================================================
## Code Quality
## ===================================================================

fmt: ## Format all Go code
	@echo '$(COLOR_INFO)Formatting code...$(COLOR_RESET)'
	@$(GO) fmt ./...
	@echo '$(COLOR_SUCCESS)✓ Code formatted$(COLOR_RESET)'

lint: ## Run linter (requires golangci-lint)
	@echo '$(COLOR_INFO)Running linter...$(COLOR_RESET)'
	@which golangci-lint > /dev/null || (echo '$(COLOR_ERROR)golangci-lint not installed. Run:$(COLOR_RESET) brew install golangci-lint' && exit 1)
	@golangci-lint run ./...
	@echo '$(COLOR_SUCCESS)✓ Lint passed$(COLOR_RESET)'

vet: ## Run go vet
	@echo '$(COLOR_INFO)Running go vet...$(COLOR_RESET)'
	@$(GO) vet ./...
	@echo '$(COLOR_SUCCESS)✓ Vet passed$(COLOR_RESET)'

## ===================================================================
## Docker
## ===================================================================

docker-up: ## Start all infrastructure services
	@echo '$(COLOR_INFO)Starting infrastructure...$(COLOR_RESET)'
	@$(DOCKER_COMPOSE) up -d postgres nats qdrant redis jaeger
	@echo '$(COLOR_SUCCESS)✓ Infrastructure started$(COLOR_RESET)'

docker-down: ## Stop all infrastructure services
	@echo '$(COLOR_INFO)Stopping infrastructure...$(COLOR_RESET)'
	@$(DOCKER_COMPOSE) down
	@echo '$(COLOR_SUCCESS)✓ Infrastructure stopped$(COLOR_RESET)'

docker-restart: docker-down docker-up ## Restart infrastructure

docker-logs: ## Show logs from all services
	@$(DOCKER_COMPOSE) logs -f

docker-logs-postgres: ## Show PostgreSQL logs
	@$(DOCKER_COMPOSE) logs -f postgres

docker-logs-nats: ## Show NATS logs
	@$(DOCKER_COMPOSE) logs -f nats

docker-ps: ## Show running containers
	@$(DOCKER_COMPOSE) ps

docker-build: ## Build Docker images
	@echo '$(COLOR_INFO)Building Docker images...$(COLOR_RESET)'
	@docker build -f deploy/Dockerfile.agent-runtime -t clawteam/agent-runtime:latest .
	@docker build -f deploy/Dockerfile.agent-runtime -t clawteam/task-service:latest --target=task-service .
	@echo '$(COLOR_SUCCESS)✓ Docker images built$(COLOR_RESET)'

docker-clean: ## Remove Docker volumes and containers
	@echo '$(COLOR_WARNING)This will delete all data. Continue? [y/N]$(COLOR_RESET)'
	@read -r confirmation; \
	if [ "$$confirmation" = "y" ] || [ "$$confirmation" = "Y" ]; then \
		$(DOCKER_COMPOSE) down -v; \
		echo '$(COLOR_SUCCESS)✓ Docker cleanup complete$(COLOR_RESET)'; \
	else \
		echo 'Aborted.'; \
	fi

## ===================================================================
## Database
## ===================================================================

db-connect: ## Connect to PostgreSQL (requires psql)
	@psql -h localhost -U clawteam -d clawteam

db-migrate: ## Run database migrations
	@echo '$(COLOR_INFO)Running migrations...$(COLOR_RESET)'
	@psql -h localhost -U clawteam -d clawteam -f deploy/init-db.sql
	@echo '$(COLOR_SUCCESS)✓ Migrations complete$(COLOR_RESET)'

db-reset: ## Reset database (dangerous!)
	@echo '$(COLOR_WARNING)This will delete all data. Continue? [y/N]$(COLOR_RESET)'
	@read -r confirmation; \
	if [ "$$confirmation" = "y" ] || [ "$$confirmation" = "Y" ]; then \
		psql -h localhost -U clawteam -d postgres -c "DROP DATABASE IF EXISTS clawteam;"; \
		psql -h localhost -U clawteam -d postgres -c "CREATE DATABASE clawteam;"; \
		$(MAKE) db-migrate; \
		echo '$(COLOR_SUCCESS)✓ Database reset complete$(COLOR_RESET)'; \
	else \
		echo 'Aborted.'; \
	fi

## ===================================================================
## Clean
## ===================================================================

clean: ## Clean build artifacts
	@echo '$(COLOR_INFO)Cleaning build artifacts...$(COLOR_RESET)'
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@$(GO) clean
	@echo '$(COLOR_SUCCESS)✓ Clean complete$(COLOR_RESET)'

clean-all: clean ## Clean everything including Docker
	@echo '$(COLOR_WARNING)Cleaning everything...$(COLOR_RESET)'
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@$(GO) clean -cache -modcache
	@echo '$(COLOR_SUCCESS)✓ Deep clean complete$(COLOR_RESET)'
