GO ?= go
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)
GOLANGCI_LINT_VERSION := $(shell cat .golangciversion)
GOLANGCI_LINT := $(CURDIR)/bin/golangci-lint

.PHONY: build check format-lint lint install-golangci-lint fmt clean

build:
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o bin/facade ./cmd/facade

check: format-lint lint
	$(GO) vet ./...
	$(GO) test ./...

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
	rm -rf bin
