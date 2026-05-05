.PHONY: build test test-race vet tidy clean install help

BIN := bin/buckle
PKG := github.com/WagnerJust/buckle
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/buckle

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/buckle

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/

help:
	@echo "make build      - build the buckle binary"
	@echo "make install    - go install ./cmd/buckle into \$$GOPATH/bin"
	@echo "make test       - run all tests"
	@echo "make test-race  - run all tests with the race detector"
	@echo "make vet        - go vet"
	@echo "make tidy       - go mod tidy"
	@echo "make clean      - remove the built binary"
