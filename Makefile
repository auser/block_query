BINARY := block_query
VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BIN_DIR := $(shell pwd)/build
CURR_DIR := $(shell pwd)
ANTLR := $(shell which antlr)

PKGS := $(shell go list ./... | grep -v vendor)

COMMIT = $(shell git rev-parse HEAD | cut -c 1-6)
BUILD_TIME = $(shell date -u '+%Y-%m-%dT%I:%M:%S%p')
MAKEFLAGS = -s

PLATFORMS := linux darwin
os = $(word 1, $@)

LDFLAGS =-ldflags "-X github.com/auser/block_query/cmd.AppName=$(BINARY) -X github.com/auser/block_query/cmd.Branch=$(BRANCH) -X github.com/auser/block_query/cmd.Version=$(VERSION) -X github.com/auser/block_query/cmd.Commit=$(COMMIT) -X github.com/auser/block_query/cmd.BuildTime=$(BUILD_TIME)"

.PHONY: parser build test

deps:
	go get -u github.com/pointlander/peg
	dep ensure

block_query.go: grammar/block_query.y
	goyacc -v grammar/y.output -o grammar/block_query.go grammar/block_query.y
	gofmt -w grammar/block_query.go

json.go:
	pigeon -o backends/json_backend/json.go  backends/json_backend/json_backend.peg
	gofmt -w backends/json_backend/json.go

clean:
	rm -f grammar/y.output grammar/block_query.go scanner/json_scanner/json.go scanner/json_scanner/y.output

build:
	$(GOPATH)/bin/peg grammar/block_query.peg
	@go build ${LDFLAGS} -o $(CURR_DIR)/build/bin/$(BINARY)

test:
	go test ./...
