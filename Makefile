NAME		:= burrows
DEST		:= ./dist
OUTPUT_BIN	:= $(DEST)/$(NAME)
PACKAGE		:= $(shell grep -E "^module" go.mod | cut -d ' ' -f 2)
GIT_REV		?= $(shell git rev-parse --short HEAD)
SOURCE_DATE_EPOCH ?= $(shell date +%s)
ifeq ($(shell uname), Darwin)
DATE       ?= $(shell TZ=UTC date -j -f "%s" ${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
else
DATE       ?= $(shell date -u -d @${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
endif
VERSION    ?= v0.1

default: help

build:	## Build binaries
	@echo '*********** BUILD EXEC ***************'
	@echo Destination: ${DEST}
	@mkdir -p ${DEST}
	@go build \
	-ldflags "-w -s -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT_REV} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -o $(OUTPUT_BIN) ./main.go

test:  ## Run the tests
	@echo '*********** TESTING ***************'
	@go test -v ./...
	
cover:  ## Produce coverage reports
	@echo '*********** CODE COVERAGE ***************'
	@go test -v -cover -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o=coverge.html

clean:  ## Remove the binaries directory
	@echo '*********** CLEANUP ***************'
	@rm -r $(DEST)

help:  ## Print help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'