SHELL = bash

# git information
GIT_COMMIT := $(shell git rev-parse --short HEAD)
LATEST_TAG := $(shell git describe --tags --abbrev=7 --always)
STABLE_TAG := $(shell git describe --tags --abbrev=0 --always)

GIT_BRANCH := $(or $(shell git rev-parse --abbrev-ref HEAD))
ifeq ($(GIT_BRANCH),$(LATEST_TAG))
GIT_BRANCH := master
endif

UNCOMMITTED := $(shell [[ `git status --porcelain | wc -l` -eq 0 ]] && echo 0 || echo 1)

FORCE_BUILD := $(shell [[ -n $${OPWIRE_FORCE_BUILD} ]] && echo 1)
OK_FOR_TEST := $(shell [[ $(UNCOMMITTED) -eq 0 ]] && echo 1)
OK_FOR_RELEASE := $(shell [[ $(UNCOMMITTED) -eq 0 && $(LATEST_TAG) = $(STABLE_TAG) ]] && echo 1)

# date/time
built_at := $(shell date +%FT%T%z)
built_by := developers@opwire.org

# go information
GO_VERSION := $(shell go version)

# LDFLAGS
GO_LDFLAGS := $(shell echo "-X main.gitCommit=${GIT_COMMIT} -X main.gitTag=${LATEST_TAG} -X main.builtAt='${built_at}' -X main.builtBy=${built_by}")

# List of OS/ARCH
TARGET_OS_ARCH= \
	darwin/386 \
	darwin/amd64 \
	linux/386 \
	linux/amd64 \
	linux/arm \
	linux/arm64 \
	freebsd/386 \
	freebsd/amd64 \
	netbsd/386 \
	netbsd/amd64 \
	openbsd/386 \
	openbsd/amd64 \
	plan9/386 \
	plan9/amd64 \
	solaris/amd64 \
	windows/386 \
	windows/amd64

build-dev:
	go build -ldflags "${GO_LDFLAGS}"

build-lab:
ifeq ($(filter 1,$(FORCE_BUILD) $(OK_FOR_TEST)),1)
	go build -ldflags "${GO_LDFLAGS}"
else
	@echo "Please commit all of changes before build a LAB edition"
endif

ifeq ($(filter 1,$(FORCE_BUILD) $(OK_FOR_RELEASE)),1)
build-all: build-clean build-mkdir
	for TARGET in ${TARGET_OS_ARCH}; do \
		IFS="/" read -ra OS_ARCH <<< $${TARGET}; \
		GOOS=$${OS_ARCH[0]}; \
		GOARCH=$${OS_ARCH[1]}; \
		ARTIFACT_NAME=opwire-agent-${LATEST_TAG}-$$GOOS-$$GOARCH; \
		[[ "$$GOOS" = "windows" ]] && BIN_EXT=".exe" || BIN_EXT=""; \
		env GOOS=$$GOOS GOARCH=$$GOARCH go build -o ./build/$$ARTIFACT_NAME/opwire-agent$$BIN_EXT -ldflags "${GO_LDFLAGS}" && \
		zip -rjm ./build/$$ARTIFACT_NAME.zip ./build/$$ARTIFACT_NAME/ && \
		rmdir ./build/$$ARTIFACT_NAME/; \
	done
else
build-all:
	@echo "Please commit all of changes and make a tag before build releases"
endif

build-clean:
	rm -rf ./build/

build-mkdir:
	mkdir -p ./build/

clean: build-clean
	go clean ./...
	find . -name \*~ | xargs -r rm -f
	rm -f ./opwire-agent
	rm -f ./opwire-lab

info:
	@echo "GO version: $(GO_VERSION)"
	@echo "Current git branch: $(GIT_BRANCH)"
	@echo "  The stable git Tag: $(STABLE_TAG)"
	@echo "  The latest git Tag: $(LATEST_TAG)"
	@echo "Current git commit: $(GIT_COMMIT)"
ifeq ($(UNCOMMITTED),0)
	@echo "Current change has committed"
else
	@echo "Current change is uncommitted"
endif
