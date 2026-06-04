# Variables
BINARY_NAME=bin/scheduler
MAIN_PATH=cmd/scheduler/main.go

.PHONY: all build run test test-cover clean

all: build

build:
	@echo "Building binary..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

run:
	@echo "Running scheduler..."
	go run $(MAIN_PATH)

test:
	@echo "Running tests..."
	go test -v ./...

test-cover:
	@echo "Running tests with coverage reporting..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin coverage.out
