PACKAGE=github.com/microcks/microcks-cli
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/build/dist
CLI_NAME=microcks
BIN_NAME=microcks

HOST_OS=$(shell go env GOOS)
HOST_ARCH=$(shell go env GOARCH)

.PHONY: build-local
build-local:
	go build -o ${DIST_DIR}/${BIN_NAME}

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

.PHONY: build-watcher
build-watcher:
	go build -o ${DIST_DIR}/${BIN_NAME}-${WATCHER_NAME} ${PACKAGE}/pkg/importer