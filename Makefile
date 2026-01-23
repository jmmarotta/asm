.DEFAULT_GOAL := all

GO ?= go
BIN ?= asm
PKG ?= ./...
RUN ?=

.PHONY: all check build bin test test-race test-cover test-one fmt fmt-check vet tidy clean help

all: fmt vet test build

check: fmt-check vet test build

build:
	$(GO) build ./...

bin:
	$(GO) build -o $(BIN) ./cmd/asm

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-cover:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -func=coverage.out

test-one:
	@if [ -n "$(RUN)" ]; then \
		$(GO) test $(PKG) -run $(RUN) -count=1; \
	else \
		$(GO) test $(PKG) -count=1; \
	fi

fmt:
	gofmt -w cmd internal

fmt-check:
	@if [ -n "`gofmt -l cmd internal`" ]; then \
		echo "gofmt required on cmd/ or internal/"; \
		exit 1; \
	fi

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

clean:
	rm -f $(BIN) coverage.out

help:
	@echo "Targets:" \
		"\n  all          - fmt, vet, test, build" \
		"\n  check        - fmt-check, vet, test, build" \
		"\n  build        - go build ./..." \
		"\n  bin          - build ./cmd/asm to ./$(BIN)" \
		"\n  test         - go test ./..." \
		"\n  test-one     - go test $(PKG) (-run $(RUN))" \
		"\n  test-race    - go test -race ./..." \
		"\n  test-cover   - go test with coverage" \
		"\n  fmt          - gofmt -w cmd internal" \
		"\n  fmt-check    - gofmt -l cmd internal" \
		"\n  vet          - go vet ./..." \
		"\n  tidy         - go mod tidy" \
		"\n  clean        - remove ./$(BIN) and coverage.out"
