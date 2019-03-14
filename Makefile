SHELL = bash

# git information
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(or $(shell git rev-parse --abbrev-ref HEAD))
LATEST_TAG := $(shell git describe --tags --abbrev=7 --always)
ifeq ($(GIT_BRANCH),$(LATEST_TAG))
	GIT_BRANCH := master
endif

# go information
GO_VERSION := $(shell go version)

# date/time
built_at := $(shell date +%FT%T%z)
built_by := developers@opwire.org

build:
	go build -ldflags "-X main.gitCommit=${GIT_COMMIT} -X main.builtAt='${built_at}' -X main.builtBy=${built_by}"

build-clean:
	rm -rf ./build/

build-all: build-clean
	mkdir -p ./build/
	for GOOS in darwin linux windows; do \
		for GOARCH in 386 amd64; do \
			[[ "$$GOOS" = "windows" ]] && BIN_EXT=".exe" || BIN_EXT=""; \
			env GOOS=$$GOOS GOARCH=$$GOARCH go build -o ./build/opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH/opwire-agent$$BIN_EXT -ldflags "-X main.gitCommit=${LATEST_TAG} -X main.builtAt='${built_at}' -X main.builtBy=${built_by}"; \
			zip -rjm ./build/opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH.zip ./build/opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH/ ; \
			rmdir ./build/opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH/; \
		done; \
	done

clean:
	go clean ./...
	find . -name \*~ | xargs -r rm -f
	rm -f ./opwire-agent
	rm -f ./opwire-lab

showinfo:
	@echo "Current go version: $(GO_VERSION)"
	@echo "Current git branch: $(GIT_BRANCH)"
	@echo "  The last git tag: $(LATEST_TAG)"
