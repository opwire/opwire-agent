SHELL = bash

# git information
GIT_COMMIT := $(shell git rev-parse --short HEAD)
LATEST_TAG := $(shell git describe --tags --abbrev=7 --always)
STABLE_TAG := $(shell git describe --tags --abbrev=0 --always)

UNCOMMITTED := 0
ifneq ($(shell git status --porcelain | wc -c), 0)
UNCOMMITTED := 1
endif

GIT_BRANCH := $(or $(shell git rev-parse --abbrev-ref HEAD))
ifeq ($(GIT_BRANCH),$(LATEST_TAG))
GIT_BRANCH := master
endif

# go information
GO_VERSION := $(shell go version)

# date/time
built_at := $(shell date +%FT%T%z)
built_by := developers@opwire.org

build-dev:
	go build -ldflags "-X main.gitCommit=${GIT_COMMIT} -X main.gitTag=${LATEST_TAG} -X main.builtAt='${built_at}' -X main.builtBy=${built_by}"

build-lab:
ifeq ($(UNCOMMITTED),0)
	go build -ldflags "-X main.gitCommit=${GIT_COMMIT} -X main.gitTag=${LATEST_TAG} -X main.builtAt='${built_at}' -X main.builtBy=${built_by}"
else
	@echo "The current code is uncommitted"
endif

ifeq ($(LATEST_TAG),$(STABLE_TAG) && $(UNCOMMITTED),0)
build-all: build-clean
	mkdir -p ./build/
	for GOOS in darwin linux windows; do \
		for GOARCH in 386 amd64; do \
			[[ "$$GOOS" = "windows" ]] && BIN_EXT=".exe" || BIN_EXT=""; \
			env GOOS=$$GOOS GOARCH=$$GOARCH go build -o ./build/opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH/opwire-agent$$BIN_EXT -ldflags "-X main.gitCommit=${GIT_COMMIT} -X main.gitTag=${LATEST_TAG} -X main.builtAt='${built_at}' -X main.builtBy=${built_by}"; \
			zip -rjm ./build/opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH.zip ./build/opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH/ ; \
			rmdir ./build/opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH/; \
		done; \
	done
endif

build-clean:
	rm -rf ./build/

clean:
	go clean ./...
	find . -name \*~ | xargs -r rm -f
	rm -f ./opwire-agent
	rm -f ./opwire-lab

info:
	@echo "Current go version: $(GO_VERSION)"
	@echo "Current git branch: $(GIT_BRANCH)"
	@echo "Current git commit: $(GIT_COMMIT)"
	@echo "  The stable git Tag: $(STABLE_TAG)"
	@echo "  The latest git Tag: $(LATEST_TAG)"
ifeq ($(UNCOMMITTED),0)
	@echo "Current git has been uncommitted"
else
	@echo "Current git is uncommitted"
endif