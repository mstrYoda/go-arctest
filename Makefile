.PHONY: all build build-cli test clean example

all: build build-cli test

build:
	@echo "Building go-arctest..."
	@mkdir -p bin
	@go build -o bin/arctest ./cmd/arctest

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/

example:
	@echo "Running example architecture tests..."
	@go test ./examples/...