# IT 02 — sign up and sign in.
#
# Run `make` or `make help` to see every target.
#
# Backend configuration is read from a local `.env` (see `.env.example`), or
# injected from Infisical by the `-secure` targets (see the README).

SHELL := /bin/bash
.DEFAULT_GOAL := help

BACKEND_DIR  := backend
FRONTEND_DIR := frontend
BACKEND_BIN  := $(BACKEND_DIR)/it02-auth-api

# Load `.env` when it exists so `make backend` needs no exported variables.
# Write values unquoted, for example: JWT_SECRET=abc123
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: help install dev backend frontend build build-backend build-frontend \
        test test-backend test-frontend typecheck fmt vet tidy check clean \
        dev-secure backend-secure secrets \
        check-env check-deps check-infisical

## ---------------------------------------------------------------- help ----

help: ## Show this help
	@echo "IT 02 — ระบบสมาชิก"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@awk 'BEGIN { FS = ":.*## " } \
	  /^[a-zA-Z0-9_-]+:.*## / { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }' \
	  $(MAKEFILE_LIST)
	@echo ""

## ------------------------------------------------------------- running ----

# The '-' on the trap line below is deliberate: Ctrl-C leaves the shell with
# exit 130, which is the normal way to stop this target rather than a failure.
dev: check-env check-deps ## Run backend and frontend together (Ctrl-C stops both)
	@echo "→ backend  http://localhost:4201"
	@echo "→ frontend http://localhost:4200"
	@echo ""
	@echo "(Ctrl-C stops both)"
	@echo ""
	-@trap 'kill 0' EXIT INT TERM; \
	( cd $(BACKEND_DIR) && go run ./cmd/api ) & \
	( cd $(FRONTEND_DIR) && npm run dev ) & \
	wait

backend: check-env ## Run the Go API on :4201
	@cd $(BACKEND_DIR) && go run ./cmd/api

frontend: check-deps ## Run the Vite dev server on :4200
	@cd $(FRONTEND_DIR) && npm run dev

## --------------------------------------------- running with Infisical ----

# Which Infisical environment to pull from: dev, staging, prod, ...
# Override per run:  make backend-secure INFISICAL_ENV=staging
INFISICAL_ENV ?= dev

# `--project-config-dir=..` because the API is started from $(BACKEND_DIR)
# while `.infisical.json` sits at the repository root, which is the monorepo
# case that flag exists for.
#
# Unset local secret values before Infisical injects them, so a stale `.env`
# value can never silently win over the selected environment.
INFISICAL_RUN = env -u JWT_SECRET -u DATABASE_URL infisical run \
                  --project-config-dir=.. --env=$(INFISICAL_ENV) --

backend-secure: check-infisical ## Run the API with secrets injected by Infisical
	@cd $(BACKEND_DIR) && $(INFISICAL_RUN) go run ./cmd/api

dev-secure: check-infisical check-deps ## Like `dev`, with the API's secrets from Infisical
	@echo "→ backend  http://localhost:4201   (secrets: Infisical $(INFISICAL_ENV))"
	@echo "→ frontend http://localhost:4200"
	@echo ""
	@echo "(Ctrl-C stops both)"
	@echo ""
	-@trap 'kill 0' EXIT INT TERM; \
	( cd $(BACKEND_DIR) && $(INFISICAL_RUN) go run ./cmd/api ) & \
	( cd $(FRONTEND_DIR) && npm run dev ) & \
	wait

secrets: check-infisical ## List the secrets Infisical would inject
	@infisical secrets --env=$(INFISICAL_ENV)

## -------------------------------------------------------------- setup -----

install: ## Install dependencies for both sides
	@cd $(FRONTEND_DIR) && npm install
	@cd $(BACKEND_DIR) && go mod download

## ------------------------------------------------------------ building ----

build: build-backend build-frontend ## Build both sides for production

build-backend: ## Compile the API binary
	@cd $(BACKEND_DIR) && go build -o ../$(BACKEND_BIN) ./cmd/api
	@echo "built $(BACKEND_BIN)"

build-frontend: check-deps ## Production build of the web app
	@cd $(FRONTEND_DIR) && npm run build

## ------------------------------------------------------------- quality ----

check: fmt vet test typecheck ## Everything CI should run

test: test-backend test-frontend ## Run both test suites

test-backend: ## Run the Go tests
	@cd $(BACKEND_DIR) && go test ./...

test-frontend: check-deps ## Run the Vitest suite
	@cd $(FRONTEND_DIR) && npm test

typecheck: check-deps ## Type check the frontend, templates included
	@cd $(FRONTEND_DIR) && npm run typecheck
	@echo "frontend type check passed"

fmt: ## Format the Go code
	@cd $(BACKEND_DIR) && gofmt -w .

vet: ## Run go vet
	@cd $(BACKEND_DIR) && go vet ./...

tidy: ## Tidy the Go module
	@cd $(BACKEND_DIR) && go mod tidy

clean: ## Remove generated build output
	@rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules/.tmp \
	        $(BACKEND_BIN)
	@echo "cleaned"

## --------------------------------------------------------- preconditions --

check-env:
	@if [ -z "$$JWT_SECRET" ]; then \
	  echo "JWT_SECRET is required."; \
	  echo ""; \
	  echo "  cp .env.example .env"; \
	  echo "  openssl rand -base64 48        # put the result in JWT_SECRET"; \
	  echo ""; \
	  echo "There is deliberately no generated fallback: a secret invented at"; \
	  echo "start up would invalidate every token issued before a restart."; \
	  echo ""; \
	  exit 1; \
	fi
	@if [ $${#JWT_SECRET} -lt 32 ]; then \
	  echo "JWT_SECRET is $${#JWT_SECRET} bytes; at least 32 are required."; \
	  echo ""; \
	  echo "  openssl rand -base64 48"; \
	  echo ""; \
		exit 1; \
	fi
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "DATABASE_URL is required."; \
		echo "Apply the Supabase migration and copy a connection string into .env."; \
		echo "See backend/DATABASE.md."; \
		exit 1; \
	fi

check-deps:
	@if [ ! -d "$(FRONTEND_DIR)/node_modules" ]; then \
	  echo "Frontend dependencies are missing. Run: make install"; \
	  exit 1; \
	fi

check-infisical:
	@command -v infisical >/dev/null 2>&1 || { \
	  echo "The Infisical CLI is not installed."; \
	  echo ""; \
	  echo "  brew install infisical/get-cli/infisical"; \
	  echo ""; \
	  echo "Other platforms: https://infisical.com/docs/cli/overview"; \
	  echo ""; \
	  echo "To run without Infisical, use 'make dev' with a local .env instead."; \
	  echo ""; \
	  exit 1; \
	}
	@if [ ! -f .infisical.json ]; then \
	  echo "This project is not linked to an Infisical project yet."; \
	  echo ""; \
	  echo "  infisical login     # once per machine"; \
	  echo "  infisical init      # writes .infisical.json with your project id"; \
	  echo ""; \
	  echo "See .infisical.example.json for the shape of that file."; \
	  echo ""; \
	  exit 1; \
	fi
