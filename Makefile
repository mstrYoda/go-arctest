.PHONY: all build test clean example

all: build test

build:
	@echo "Building go-arctest..."
	@go build -o bin/arctest ./main.go

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin/

example:
	@echo "Running example architecture tests..."
	@go test ./examples/... 