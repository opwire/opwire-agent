SHELL = bash

# git information
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(or $(shell git rev-parse --abbrev-ref HEAD))
LATEST_TAG := $(shell git describe --tags --abbrev=0 --always)
ifeq ($(GIT_BRANCH),$(LATEST_TAG))
	GIT_BRANCH := master
endif

# go information
GO_VERSION := $(shell go version)

# date/time
built_at := $(shell date +%FT%T%z)
built_by := devteam@opwire.org

build:
	go build -ldflags "-X main.gitCommit=${GIT_COMMIT} -X main.builtAt='${built_at}' -X main.builtBy=${built_by}"

clean:
	go clean ./...
	find . -name \*~ | xargs -r rm -f
	rm -f opwire-agent
	rm -f opwire-lab

showinfo:
	@echo "Current go version: $(GO_VERSION)"
	@echo "Current git branch: $(GIT_BRANCH)"
	@echo "  The last git tag: $(LATEST_TAG)"
