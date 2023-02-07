# --------------------------------------------------------------------------
# Makefile for the Fantom API GraphQL Server
#
# v0.1 (2020/03/09)  - Initial version, base API server build.
# (c) Fantom Foundation, 2020
# --------------------------------------------------------------------------

# project related vars
PROJECT := $(shell pwd)

# go related vars
GO_BASE := $(shell pwd)
GO_BIN := $(CURDIR)/build

# compile time variables will be injected into the app
APP_VERSION := 1.1.0
BUILD_DATE := $(shell date)
BUILD_COMPILER := $(shell go version)
BUILD_COMMIT := $(shell git show --format="%H" --no-patch)
BUILD_COMMIT_TIME := $(shell git show --format="%cD" --no-patch)

## server: Make the API server as build/apiserver
win:
	go build \
	-ldflags="-X 'galaxy-graphql/cmd/apiserver/build.Version=$(APP_VERSION)' -X 'galaxy-graphql/cmd/apiserver/build.Time=$(BUILD_DATE)' -X 'galaxy-graphql/cmd/apiserver/build.Compiler=$(BUILD_COMPILER)' -X 'galaxy-graphql/cmd/apiserver/build.Commit=$(BUILD_COMMIT)' -X 'galaxy-graphql/cmd/apiserver/build.CommitTime=$(BUILD_COMMIT_TIME)'" \
	-o $(GO_BIN)/apiserver.exe \
	./cmd/apiserver

server:
	go build \
	-ldflags="-X 'galaxy-graphql/cmd/apiserver/build.Version=$(APP_VERSION)' -X 'galaxy-graphql/cmd/apiserver/build.Time=$(BUILD_DATE)' -X 'galaxy-graphql/cmd/apiserver/build.Compiler=$(BUILD_COMPILER)' -X 'galaxy-graphql/cmd/apiserver/build.Commit=$(BUILD_COMMIT)' -X 'galaxy-graphql/cmd/apiserver/build.CommitTime=$(BUILD_COMMIT_TIME)'" \
	-o $(GO_BIN)/apiserver \
	./cmd/apiserver

test:
	go test \
	-ldflags="-X 'galaxy-graphql/cmd/apiserver/build.Version=$(APP_VERSION)' -X 'galaxy-graphql/cmd/apiserver/build.Time=$(BUILD_DATE)' -X 'galaxy-graphql/cmd/apiserver/build.Compiler=$(BUILD_COMPILER)' -X 'galaxy-graphql/cmd/apiserver/build.Commit=$(BUILD_COMMIT)' -X 'galaxy-graphql/cmd/apiserver/build.CommitTime=$(BUILD_COMMIT_TIME)'" \
	./...

.PHONY: help test
all: help
help: Makefile
	@echo
	@echo "Choose a make command in "$(PROJECT)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
