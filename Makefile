GO ?= go
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build check format-lint fmt clean

build:
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o bin/facade ./cmd/facade

check: format-lint
	$(GO) vet ./...
	$(GO) test ./...

format-lint:
	@unformatted="$$(gofmt -l .)" || exit 1; \
	if [ -n "$$unformatted" ]; then \
		printf 'Go files need formatting:\n%s\nRun make fmt to fix them.\n' "$$unformatted"; \
		exit 1; \
	fi

fmt:
	$(GO) fmt ./...

clean:
	rm -rf bin
