ifneq (,$(wildcard ./.env))
	include .env
	export
endif

name ?=
steps ?= 1

.PHONY: help run build test fmt lint migrate-up migrate-down migrate-status migrate-create seed-admin

help: ## Display available targets
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the boilerplate HTTP server
	go run ./cmd/main.go serve

build: ## Build the CLI binary
	@mkdir -p bin
	go build -o ./bin/__PROJECT_NAME__ ./cmd/main.go

test: ## Run unit tests
	go test ./...

fmt: ## Format source files
	gofmt -w ./cmd ./internal

lint: ## Run a lightweight lint pass with go vet
	go vet ./...

migrate-up: ## Apply all pending SQL migrations
	go run ./cmd/main.go migrate up

migrate-down: ## Roll back SQL migrations (usage: make migrate-down steps=1)
	go run ./cmd/main.go migrate down --steps $(steps)

migrate-status: ## Show current migration version and dirty flag
	go run ./cmd/main.go migrate status

migrate-create: ## Create a new migration pair (usage: make migrate-create name=add_audit_log)
	@test -n "$(name)" || (echo "usage: make migrate-create name=add_audit_log" && exit 1)
	go run ./cmd/main.go migrate create "$(name)"

seed-admin: ## Create or update the bootstrap admin user from env values
	go run ./cmd/main.go seed admin

.DEFAULT_GOAL := help
