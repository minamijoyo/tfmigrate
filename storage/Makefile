NAME := tfmigrate-storage

.DEFAULT_GOAL := check

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./...

.PHONY: check
check: lint test
