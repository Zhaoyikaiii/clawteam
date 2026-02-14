# ClawTeam Makefile
# Usage: make <target>

.PHONY: help dev test build run clean docker-up docker-down docker-logs deps fmt lint

.DEFAULT_GOAL := help

# Variables
BINARY_NAME_AGENT=agent-runtime
BINARY_NAME_IM=im-core
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

help:
	@echo '$(COLOR_SUCCESS)ClawTeam Development Commands$(COLOR_RESET)'
	@echo ''
	@echo '$(COLOR_INFO)Available targets:$(COLOR_RESET)'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_INFO)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'

## ===================================================================
## Development
## ===================================================================

dev:
	@echo '$(COLOR_INFO)Starting development environment...$(COLOR_RESET)'
	@$(MAKE) docker-up
	@echo '$(COLOR_SUCCESS)✓ Development environment is ready$(COLOR_RESET)'

deps:
	@echo '$(COLOR_INFO)Downloading dependencies...$(COLOR_RESET)'
	@$(GO) mod download
	@$(GO) mod tidy
	@echo '$(COLOR_SUCCESS)✓ Dependencies ready$(COLOR_RESET)'

## ===================================================================
## Build
## ===================================================================

build:
	@echo '$(COLOR_INFO)Building services...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)
	@echo '  Building $(BINARY_NAME_AGENT)...'
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_AGENT) ./$(CMD_DIR)/agent-runtime
	@echo '  Building $(BINARY_NAME_IM)...'
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_IM) ./$(CMD_DIR)/im-core
	@echo '$(COLOR_SUCCESS)✓ Build complete$(COLOR_RESET)'

build-linux:
	@echo '$(COLOR_INFO)Building Linux binaries...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)
	@echo '  Building $(BINARY_NAME_AGENT)...'
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_AGENT) ./$(CMD_DIR)/agent-runtime
	@echo '  Building $(BINARY_NAME_IM)...'
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_IM) ./$(CMD_DIR)/im-core
	@echo '$(COLOR_SUCCESS)✓ Linux build complete$(COLOR_RESET)'

## ===================================================================
## Run
## ===================================================================

run-agent:
	@echo '$(COLOR_INFO)Starting Agent Runtime...$(COLOR_RESET)'
	@$(GO) run ./$(CMD_DIR)/agent-runtime/main.go

run-im:
	@echo '$(COLOR_INFO)Starting IM Core Service...$(COLOR_RESET)'
	@$(GO) run ./$(CMD_DIR)/im-core/main.go

## ===================================================================
## Test
## ===================================================================

test:
	@echo '$(COLOR_INFO)Running tests...$(COLOR_RESET)'
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo '$(COLOR_SUCCESS)✓ Tests passed. Coverage report: coverage.html$(COLOR_RESET)'

test-short:
	@echo '$(COLOR_INFO)Running tests (fast)...$(COLOR_RESET)'
	@$(GO) test -v ./...

## ===================================================================
## Code Quality
## ===================================================================

fmt:
	@echo '$(COLOR_INFO)Formatting code...$(COLOR_RESET)'
	@$(GO) fmt ./...
	@echo '$(COLOR_SUCCESS)✓ Code formatted$(COLOR_RESET)'

lint:
	@echo '$(COLOR_INFO)Running linter...$(COLOR_RESET)'
	@golangci-lint run ./...
	@echo '$(COLOR_SUCCESS)✓ Lint passed$(COLOR_RESET)'

vet:
	@echo '$(COLOR_INFO)Running go vet...$(COLOR_RESET)'
	@$(GO) vet ./...
	@echo '$(COLOR_SUCCESS)✓ Vet passed$(COLOR_RESET)'

## ===================================================================
## Docker
## ===================================================================

docker-up:
	@echo '$(COLOR_INFO)Starting infrastructure...$(COLOR_RESET)'
	@$(DOCKER_COMPOSE) up -d redis qdrant
	@echo '$(COLOR_SUCCESS)✓ Infrastructure started$(COLOR_RESET)'

docker-down:
	@echo '$(COLOR_INFO)Stopping infrastructure...$(COLOR_RESET)'
	@$(DOCKER_COMPOSE) down
	@echo '$(COLOR_SUCCESS)✓ Infrastructure stopped$(COLOR_RESET)'

docker-restart: docker-down docker-up

docker-logs:
	@$(DOCKER_COMPOSE) logs -f

docker-ps:
	@$(DOCKER_COMPOSE) ps

## ===================================================================
## Database
## ===================================================================

db-connect:
	@echo '$(COLOR_INFO)Connecting to database...$(COLOR_RESET)'
	@psql -h localhost -U clawteam -d clawteam

db-migrate:
	@echo '$(COLOR_INFO)Running migrations...$(COLOR_RESET)'
	@psql -h localhost -U clawteam -d clawteam -f deploy/init-db.sql
	@echo '$(COLOR_SUCCESS)✓ Migrations complete$(COLOR_RESET)'

db-reset:
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

clean:
	@echo '$(COLOR_INFO)Cleaning build artifacts...$(COLOR_RESET)'
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@$(GO) clean
	@echo '$(COLOR_SUCCESS)✓ Clean complete$(COLOR_RESET)'

clean-all: clean
	@echo '$(COLOR_WARNING)Cleaning everything...$(COLOR_RESET)'
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@$(GO) clean -cache -modcache
	@echo '$(COLOR_SUCCESS)✓ Deep clean complete$(COLOR_RESET)'
