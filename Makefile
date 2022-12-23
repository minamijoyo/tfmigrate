NAME := tfmigrate

.DEFAULT_GOAL := build

.PHONY: deps
deps:
	go mod download

.PHONY: build
build: deps
	go build -buildvcs=false -o bin/$(NAME)

.PHONY: install
install: deps
	go install

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test: build
	go test ./...

.PHONY: generate-plugin-cache
generate-plugin-cache:
	scripts/testacc/generate_plugin_cache.sh

.PHONY: testacc
testacc: build generate-plugin-cache
	TEST_ACC=1 go test -count=1 -failfast -timeout=20m ./...

.PHONY: check
check: lint test
