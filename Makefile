GO ?= go
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build check fmt clean

build:
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o bin/facade ./cmd/facade

check:
	test -z "$$(gofmt -l cmd)"
	$(GO) vet ./...
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf bin
