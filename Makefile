.PHONY: help up down dev db-up db-down build test lint migrate-up migrate-down migrate-create swagger run

# Default environment variables (copy .env.example to .env and override).
-include .env
export

API_CONTAINER = app-api
MIGRATE_DIR = migrations

# ==========================================
# MAIN COMMANDS
# ==========================================

help: ## Show this help menu
	@echo "Go Kit Backend Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Start production stack (Postgres + Redis + NATS + API)
	docker compose up --build -d

down: ## Stop all services
	docker compose down

dev: ## Start development stack with hot reload
	DOCKER_TARGET=development docker compose -f docker-compose.yml -f docker-compose.override.yml up --build -d

dev-down: ## Stop development stack
	docker compose -f docker-compose.yml -f docker-compose.override.yml down

db-up: ## Start only databases (Postgres + Redis + NATS)
	docker compose up -d postgres redis nats

db-down: ## Stop only databases
	docker compose down postgres redis nats

# ==========================================
# BUILD & TEST
# ==========================================

build: ## Build the API binary locally
	go build -o bin/api ./cmd/api

run: build ## Run the API binary locally (requires local Postgres + Redis + NATS)
	./bin/api

test: ## Run all tests
	go test -race -cover ./...

lint: ## Run go vet
	go vet ./...

# ==========================================
# MIGRATIONS
# ==========================================

migrate-up: ## Apply all pending Goose migrations
	goose -dir $(MIGRATE_DIR) postgres "$(POSTGRES_DSN)" up

migrate-down: ## Rollback the last Goose migration
	goose -dir $(MIGRATE_DIR) postgres "$(POSTGRES_DSN)" down

migrate-create: ## Create a new Goose migration (usage: make migrate-create NAME=add_orders_table)
	goose -dir $(MIGRATE_DIR) create $(NAME) sql

migrate-status: ## Show migration status
	goose -dir $(MIGRATE_DIR) postgres "$(POSTGRES_DSN)" status

# ==========================================
# SWAGGER
# ==========================================

swagger: ## Generate Swagger docs from annotations
	swag init -g cmd/api/main.go -o docs

# ==========================================
# UTILITIES
# ==========================================

logs: ## Tail API logs
	docker compose logs -f api

fmt: ## Format Go code
	go fmt ./...
