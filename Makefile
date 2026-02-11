# NER Studio Makefile

# Default to help
.DEFAULT_GOAL := help

# Colors
BLUE := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RESET := \033[0m

## help: Show this help message
.PHONY: help
help:
	@echo "$(BLUE)NER Studio - Available Commands:$(RESET)"
	@grep -E '^##' Makefile | sed 's/## /  /'

## dev: Start both frontend and backend development servers
.PHONY: dev
dev:
	@echo "$(GREEN)Starting development servers...$(RESET)"
	@make dev-api & make dev-web
	@wait

## dev-api: Start backend with hot reload
.PHONY: dev-api
dev-api:
	@echo "$(GREEN)Starting API server...$(RESET)"
	@cd apps/api && air || go run cmd/server/main.go

## dev-web: Start frontend dev server
.PHONY: dev-web
dev-web:
	@echo "$(GREEN)Starting web server...$(RESET)"
	@cd apps/web && npm run dev

## build-api: Build Go binary
.PHONY: build-api
build-api:
	@echo "$(GREEN)Building API...$(RESET)"
	@cd apps/api && go build -o dist/server cmd/server/main.go

## build-web: Build React for production
.PHONY: build-web
build-web:
	@echo "$(GREEN)Building Web...$(RESET)"
	@cd apps/web && npm run build

## build: Build both frontend and backend
.PHONY: build
build: build-api build-web

## test-api: Run Go tests
.PHONY: test-api
test-api:
	@echo "$(GREEN)Running API tests...$(RESET)"
	@cd apps/api && go test ./...

## test-web: Run React tests
.PHONY: test-web
test-web:
	@echo "$(GREEN)Running Web tests...$(RESET)"
	@cd apps/web && npm test

## test: Run all tests
.PHONY: test
test: test-api test-web

## lint-api: Lint Go code
.PHONY: lint-api
lint-api:
	@echo "$(GREEN)Linting API...$(RESET)"
	@cd apps/api && golangci-lint run || go vet ./...

## lint-web: Lint frontend
.PHONY: lint-web
lint-web:
	@echo "$(GREEN)Linting Web...$(RESET)"
	@cd apps/web && npm run lint

## lint: Run all linters
.PHONY: lint
lint: lint-api lint-web

## db-start: Start Supabase local development
.PHONY: db-start
db-start:
	@echo "$(GREEN)Starting Supabase...$(RESET)"
	@supabase start

## db-stop: Stop Supabase local development
.PHONY: db-stop
db-stop:
	@echo "$(YELLOW)Stopping Supabase...$(RESET)"
	@supabase stop

## db-migrate: Run database migrations
.PHONY: db-migrate
db-migrate:
	@echo "$(GREEN)Running migrations...$(RESET)"
	@supabase db push

## db-reset: Reset database and re-run migrations
.PHONY: db-reset
db-reset:
	@echo "$(YELLOW)Resetting database...$(RESET)"
	@supabase db reset

## db-seed: Seed development data
.PHONY: db-seed
db-seed:
	@echo "$(GREEN)Seeding database...$(RESET)"
	@supabase seed apply

## deps-api: Install Go dependencies
.PHONY: deps-api
deps-api:
	@echo "$(GREEN)Installing API dependencies...$(RESET)"
	@cd apps/api && go mod download

## deps-web: Install frontend dependencies
.PHONY: deps-web
deps-web:
	@echo "$(GREEN)Installing Web dependencies...$(RESET)"
	@cd apps/web && npm install

## deps: Install all dependencies
.PHONY: deps
deps: deps-api deps-web
