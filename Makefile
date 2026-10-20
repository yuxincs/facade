GO ?= go
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)
GOLANGCI_LINT_VERSION := $(shell cat .golangciversion)
GOLANGCI_LINT := $(CURDIR)/bin/golangci-lint

.PHONY: build check cover format-lint lint install-golangci-lint fmt clean

build:
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o bin/facade ./cmd/facade

check: format-lint lint
	$(GO) vet ./...
	$(GO) test ./...

cover:
	rm -rf .coverage
	mkdir -p .coverage
	FACADE_COVERAGE_DIR="$(CURDIR)/.coverage" $(GO) test -count=1 -covermode=atomic -coverpkg=./... ./... -args -test.gocoverdir="$(CURDIR)/.coverage"
	# Build the standard coverage tool for SDKs that omit its prebuilt binary.
	$(GO) run cmd/covdata textfmt -i=.coverage -o=coverage.out

format-lint: install-golangci-lint
	"$(GOLANGCI_LINT)" fmt --diff

lint: install-golangci-lint
	"$(GOLANGCI_LINT)" run

install-golangci-lint:
	@if [ "$$('$(GOLANGCI_LINT)' version --short 2>/dev/null)" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		installer=$$(mktemp) || exit 1; \
		trap 'rm -f "$$installer"' EXIT; \
		curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/v$(GOLANGCI_LINT_VERSION)/install.sh" -o "$$installer" && \
		sh "$$installer" -b "$(CURDIR)/bin" "v$(GOLANGCI_LINT_VERSION)"; \
	fi

fmt: install-golangci-lint
	"$(GOLANGCI_LINT)" fmt

clean:
	rm -rf bin .coverage coverage.out
