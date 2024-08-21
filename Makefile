.DEFAULT_GOAL := build

.PHONY: fmt
fmt:
	go fmt

.PHONY: lint
lint:
	staticcheck
	go vet

.PHONY: test
test: fmt lint
	go test

.PHONY: build
build: test
	go mod tidy
	go build
