INSTALL_DIR ?= $(if $(GOBIN),$(GOBIN),$(shell go env GOPATH)/bin)

.PHONY: help build install test fmt lint smoke

help: ## Display available targets
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the generator binary
	@mkdir -p bin
	go build -o ./bin/go-template ./cmd/go-template

install: ## Install the generator binary into GOBIN or GOPATH/bin
	@mkdir -p "$(INSTALL_DIR)"
	go build -o "$(INSTALL_DIR)/go-template" ./cmd/go-template
	@echo "installed $(INSTALL_DIR)/go-template"

test: ## Run unit tests
	go test ./...

fmt: ## Format source files
	gofmt -w ./cmd ./internal

lint: ## Run a lightweight lint pass with go vet
	go vet ./...

smoke: build ## Generate a sample project into ./tmp-smoke
	rm -rf ./tmp-smoke
	./bin/go-template init ./tmp-smoke --module github.com/example/tmp-smoke
	cd ./tmp-smoke && go test ./... && ./generate.sh
