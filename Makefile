SHELL := /usr/bin/env bash
.DEFAULT_GOAL := all

MAKEFLAGS += --no-print-directory

PROJECT_ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: help # Print this help message.
help:
	@grep -E '^\.PHONY: [a-zA-Z0-9_-]+ .*?# .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = "(: |#)"}; {printf "%-30s %s\n", $$2, $$3}'

.PHONY: all # Build the application.
all: build

.PHONY: build # Build the binary for the current platform.
build:
	./tools/goreleaser.sh build --snapshot --clean --single-target

.PHONY: build-all # Build binaries for all platforms.
build-all:
	./tools/goreleaser.sh build --snapshot --clean

.PHONY: release-snapshot # Create a full release snapshot (binaries, archives, packages).
release-snapshot:
	./tools/goreleaser.sh release --snapshot --clean --skip=publish

.PHONY: run # Run the application locally.
run: build
	./dist/admiral_$(shell go env GOOS)_$(shell go env GOARCH)*/admiral

.PHONY: test # Run unit tests.
test:
	go test -race -covermode=atomic ./...

.PHONY: test-verbose # Run unit tests with verbose output.
test-verbose:
	go test -v -race -covermode=atomic ./...

.PHONY: lint # Lint the code.
lint:
	./tools/golangci-lint.sh run --timeout 2m30s

.PHONY: lint-fix # Lint and fix the code.
lint-fix:
	./tools/golangci-lint.sh run --fix
	go mod tidy

.PHONY: fmt # Format the code.
fmt:
	go fmt ./...

.PHONY: verify # Verify go modules are tidy.
verify:
	go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod or go.sum is not tidy" && exit 1)

.PHONY: release # Tag and push the next version (auto-detected from commits).
release:
	@echo "Current version: $$(./tools/svu.sh current)"
	@echo "Next version:    $$(./tools/svu.sh next)"
	@echo ""
	@read -p "Proceed? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	git tag -a $$(./tools/svu.sh next) -m "Release $$(./tools/svu.sh next)"
	git push origin $$(./tools/svu.sh next)

.PHONY: release-patch # Tag and push a patch release.
release-patch:
	@echo "Current version: $$(./tools/svu.sh current)"
	@echo "Next version:    $$(./tools/svu.sh patch)"
	git tag -a $$(./tools/svu.sh patch) -m "Release $$(./tools/svu.sh patch)"
	git push origin $$(./tools/svu.sh patch)

.PHONY: release-minor # Tag and push a minor release.
release-minor:
	@echo "Current version: $$(./tools/svu.sh current)"
	@echo "Next version:    $$(./tools/svu.sh minor)"
	git tag -a $$(./tools/svu.sh minor) -m "Release $$(./tools/svu.sh minor)"
	git push origin $$(./tools/svu.sh minor)

.PHONY: release-major # Tag and push a major release.
release-major:
	@echo "Current version: $$(./tools/svu.sh current)"
	@echo "Next version:    $$(./tools/svu.sh major)"
	git tag -a $$(./tools/svu.sh major) -m "Release $$(./tools/svu.sh major)"
	git push origin $$(./tools/svu.sh major)

.PHONY: version # Show current and next version.
version:
	@echo "Current: $$(./tools/svu.sh current)"
	@echo "Next:    $$(./tools/svu.sh next)"

.PHONY: deps # Download dependencies.
deps:
	go mod download
	go mod tidy
