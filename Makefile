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

.PHONY: legacy-tfstate
legacy-tfstate:
	# Generate a 0.12.31 tfstate file for use in replace-provider tests.
	docker run \
		--interactive \
		--rm \
		--tty \
		--volume $(shell pwd):/src \
		--workdir /src/test-fixtures/legacy-tfstate \
		--entrypoint /bin/sh \
		hashicorp/terraform:0.12.31 \
			-c \
				"terraform init && \
				terraform apply -auto-approve"

# Update the install-binary target to use a symlink in ~/.local/bin
.PHONY: install-binary
install-binary: build
	mkdir -p $(HOME)/.local/bin
	ln -sf $(PWD)/bin/$(NAME) $(HOME)/.local/bin/$(NAME)
