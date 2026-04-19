.PHONY: run build test lint clean

BINARY_NAME=test-phase08

run:
	go run cmd/server/main.go

build:
	go build -o bin/$(BINARY_NAME) cmd/server/main.go

test:
	go test -v -race -coverprofile=coverage.txt ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
	go clean

docker-build:
	docker build -t $(BINARY_NAME) .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
