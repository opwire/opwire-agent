SHELL = bash

# git information
BRANCH := $(or $(shell git rev-parse --abbrev-ref HEAD))
LAST_TAG := $(shell git describe --tags --abbrev=0 --always)
ifeq ($(BRANCH),$(LAST_TAG))
	BRANCH := master
endif

# go information
GO_VERSION := $(shell go version)

clean:
	go clean ./...
	find . -name \*~ | xargs -r rm -f
	rm -f opwire-agent

showinfo:
	@echo "Current go version: $(GO_VERSION);"
	@echo "Current git branch: $(BRANCH);"
	@echo "  The last git tag: $(LAST_TAG)"
