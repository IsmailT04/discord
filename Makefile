# =============================================================================
# Discord-like monorepo — root Makefile
# Layout: backend/ (Go) · frontend/ (Vite) · database/ · livekit/
# Compose: backend/deploy/compose/docker-compose.yml (add in Update 0/1)
# =============================================================================

.DEFAULT_GOAL := help

# ------------------------------------------------------------------------------
# Variables
# ------------------------------------------------------------------------------
ENV            ?= dev
BACKEND_DIR    := backend
FRONTEND_DIR   := frontend
DATABASE_DIR   := database
COMPOSE_DIR    := backend/deploy/compose
COMPOSE_FILE   := $(COMPOSE_DIR)/docker-compose.yml
COMPOSE        := docker compose -f $(COMPOSE_FILE) --project-directory $(COMPOSE_DIR)

# Optional: NAME=add_users make migrate-create
NAME           ?=

.PHONY: help bootstrap env-check \
	ensure-compose-env up down logs ps \
	migrate-up migrate-down migrate-status migrate-create \
	migrate-drop migrate-reset migrate-fresh \
	dev-api dev-web \
	test test-backend test-frontend \
	test-coverage test-coverage-html \
	lint lint-backend lint-frontend fmt tidy \
	build build-api build-web clean \
	signoz-up signoz-down signoz-logs \
	confirm confirm-prod

# ------------------------------------------------------------------------------
# Safety
# ------------------------------------------------------------------------------
confirm: ## Ask for confirmation (used by destructive targets)
	@read -p "Are you sure? [y/N] " ans; \
	if [ "$$ans" != "y" ]; then echo "Aborted."; exit 1; fi

confirm-prod: ## Extra confirmation when ENV=prod
	@if [ "$(ENV)" = "prod" ]; then \
		read -p "WARNING: ENV=prod. Type 'yes' to continue: " ans; \
		if [ "$$ans" != "yes" ]; then echo "Aborted."; exit 1; fi; \
	fi

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------
help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ------------------------------------------------------------------------------
# Setup & bootstrap
# ------------------------------------------------------------------------------
bootstrap: ## Install frontend deps and tidy Go modules
	@echo "==> frontend: npm install"
	cd $(FRONTEND_DIR) && npm install
	@echo "==> backend: go mod tidy"
	cd $(BACKEND_DIR) && go mod tidy
	@echo "==> Done. Copy *.env.example files to .env before running services."

env-check: ## Remind which env example files to copy
	@echo "Expected env templates (create when ready):"
	@echo "  $(BACKEND_DIR)/.env.example  -> $(BACKEND_DIR)/.env"
	@echo "  $(FRONTEND_DIR)/.env.example -> $(FRONTEND_DIR)/.env"
	@echo "  livekit/.env.example         -> livekit/.env  (optional)"
	@echo "  $(COMPOSE_DIR)/.env.example  -> $(COMPOSE_DIR)/.env  (compose)"

# ------------------------------------------------------------------------------
# Local infrastructure (Compose)
# Requires: $(COMPOSE_FILE) — see backend/deploy/compose/README.md
# ------------------------------------------------------------------------------
ensure-compose-env: ## Copy compose .env.example → .env if missing
	@if [ ! -f "$(COMPOSE_DIR)/.env" ]; then \
		cp "$(COMPOSE_DIR)/.env.example" "$(COMPOSE_DIR)/.env"; \
		echo "==> Created $(COMPOSE_DIR)/.env from .env.example"; \
	fi

up: ensure-compose-env ## Start local infra (Postgres, Redis)
	$(COMPOSE) up -d --wait

down: ## Stop local infra
	$(COMPOSE) down

logs: ## Follow Compose logs
	$(COMPOSE) logs -f

ps: ## Show Compose service status
	$(COMPOSE) ps

# ------------------------------------------------------------------------------
# Database migrations (host Go CLI → database/migrations)
# Requires: cmd/migrate implemented; Postgres up; backend/.env loaded by you/shell
# ------------------------------------------------------------------------------
migrate-up: ## Apply migrations
	cd $(BACKEND_DIR) && go run ./cmd/migrate up

migrate-down: ## Roll back the last migration
	cd $(BACKEND_DIR) && go run ./cmd/migrate down

migrate-status: ## Show migration status
	cd $(BACKEND_DIR) && go run ./cmd/migrate status

migrate-create: ## Create migration pair (NAME=add_foo)
	@test -n "$(NAME)" || (echo "Usage: make migrate-create NAME=add_foo"; exit 1)
	cd $(BACKEND_DIR) && go run ./cmd/migrate create $(NAME)

migrate-drop: confirm confirm-prod ## Drop database (destructive)
	cd $(BACKEND_DIR) && go run ./cmd/migrate drop

migrate-reset: confirm confirm-prod ## Reset database (destructive)
	cd $(BACKEND_DIR) && go run ./cmd/migrate reset

migrate-fresh: confirm confirm-prod ## Drop all + migrate up (destructive)
	cd $(BACKEND_DIR) && go run ./cmd/migrate fresh

# ------------------------------------------------------------------------------
# Development loops (host processes; infra via Compose)
# ------------------------------------------------------------------------------
dev-api: ## Run Go API (cmd/api)
	cd $(BACKEND_DIR) && go run ./cmd/api

dev-web: ## Run Vite frontend
	cd $(FRONTEND_DIR) && npm run dev

# ------------------------------------------------------------------------------
# Quality
# ------------------------------------------------------------------------------
test: test-backend ## Run backend tests (default)

test-backend: ## Run Go tests
	cd $(BACKEND_DIR) && go test ./...

test-frontend: ## Run frontend lint as smoke check (add unit tests later)
	cd $(FRONTEND_DIR) && npm run lint

test-coverage: ## Go tests with coverage profile
	cd $(BACKEND_DIR) && go test ./... -coverprofile=coverage.out

test-coverage-html: test-coverage ## Coverage + HTML report
	cd $(BACKEND_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Wrote $(BACKEND_DIR)/coverage.html"

lint: lint-backend lint-frontend ## Lint backend + frontend

lint-backend: ## go vet
	cd $(BACKEND_DIR) && go vet ./...

lint-frontend: ## oxlint
	cd $(FRONTEND_DIR) && npm run lint

fmt: ## Format Go sources
	cd $(BACKEND_DIR) && go fmt ./...

tidy: ## go mod tidy
	cd $(BACKEND_DIR) && go mod tidy

# ------------------------------------------------------------------------------
# Build & clean
# ------------------------------------------------------------------------------
build: build-api build-web ## Build API binary + frontend dist

build-api: ## Compile API to backend/bin/api
	mkdir -p $(BACKEND_DIR)/bin
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api

build-web: ## Production frontend build
	cd $(FRONTEND_DIR) && npm run build

clean: ## Remove build artifacts and coverage files
	rm -rf $(BACKEND_DIR)/bin \
		$(BACKEND_DIR)/coverage.out \
		$(BACKEND_DIR)/coverage.html \
		$(FRONTEND_DIR)/dist \
		$(FRONTEND_DIR)/dist-ssr

# ------------------------------------------------------------------------------
# Observability (SigNoz) — compose profile "observability"
# ------------------------------------------------------------------------------
signoz-up: ensure-compose-env ## Start OTel collector + Jaeger (observability profile)
	$(COMPOSE) --profile observability up -d --wait

signoz-down: ## Stop observability profile services
	$(COMPOSE) --profile observability stop

signoz-logs: ## Follow observability-related Compose logs
	$(COMPOSE) --profile observability logs -f
