.PHONY: help build test test-coverage lint fmt clean

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build:
	go build -o bin/url-shortener cmd/api/main.go

test:
	go test -v ./...

test-coverage:
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

lint:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin/ coverage/

.DEFAULT_GOAL := help
