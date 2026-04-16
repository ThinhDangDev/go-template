.PHONY: build install test lint clean

BINARY_NAME=go-template
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

install:
	go install $(LDFLAGS) .

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
	go clean

dev: build
	./bin/$(BINARY_NAME)
