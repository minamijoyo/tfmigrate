NAME := tfmigrate

.DEFAULT_GOAL := build

.PHONY: deps
deps:
	go mod download

.PHONY: build
build: deps
	go build -o bin/$(NAME)

.PHONY: install
install: deps
	go install

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test: build
	go test ./...

.PHONY: testacc
testacc: build
	TEST_ACC=1 go test -count=1 -failfast -timeout=20m ./...

.PHONY: check
check: lint test
