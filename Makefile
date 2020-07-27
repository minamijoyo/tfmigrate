NAME := tfmigrate

ifndef GOBIN
GOBIN := $(shell echo "$${GOPATH%%:*}/bin")
endif

GOLINT := $(GOBIN)/golint

$(GOLINT): ; @go install golang.org/x/lint/golint

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
lint: $(GOLINT)
	golint $$(go list ./... | grep -v /vendor/)

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test: deps
	go test ./...

.PHONY: testacc
testacc: deps
	TEST_ACC=1 go test -count=1 ./...

.PHONY: check
check: lint vet test build
