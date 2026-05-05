PACKAGE=github.com/microcks/microcks-cli
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/build/dist
CLI_NAME=microcks
BIN_NAME=microcks
WATCHER_NAME=watcher

HOST_OS=$(shell go env GOOS)
HOST_ARCH=$(shell go env GOARCH)
GO ?= go
BUILD_FLAGS ?=
VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo unknown)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS=-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.Commit=${COMMIT}

.PHONY: build-local
build-local:
	$(GO) build $(BUILD_FLAGS) -ldflags "${LDFLAGS}" -o ${DIST_DIR}/${BIN_NAME}

.PHONY: clean
clean:
	rm -rf ${CURRENT_DIR}/build/dist

.PHONY: build-binaries
build-binaries: 
	make BIN_NAME=${CLI_NAME}-linux-amd64 GOOS=linux build-local
	make BIN_NAME=${CLI_NAME}-linux-arm64 GOOS=linux GOARCH=arm64 build-local
	make BIN_NAME=${CLI_NAME}-darwin-amd64 GOOS=darwin build-local
	make BIN_NAME=${CLI_NAME}-darwin-arm64 GOOS=darwin GOARCH=arm64 build-local
	make BIN_NAME=${CLI_NAME}-windows-amd64.exe GOOS=windows build-local
	make BIN_NAME=${CLI_NAME}-windows-386.exe GOOS=windows GOARCH=386 build-local

.PHONY: build-release
build-release:
	$(eval RELEASE_VERSION := $(shell git describe --tags --exact-match 2>/dev/null))
	@test -n "${RELEASE_VERSION}" || (echo "build-release must be run from a git tag" && exit 1)
	make VERSION=${RELEASE_VERSION} build-binaries

.PHONY: build-watcher
build-watcher:
	$(GO) build $(BUILD_FLAGS) -o ${DIST_DIR}/${BIN_NAME}-${WATCHER_NAME} ${PACKAGE}/watcher
