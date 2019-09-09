TAG := `git describe --tags --always`
VERSION :=
# MOD can either be "readonly" or "vendor".
# The default is "readonly" which fetches the sources to the Go module cache.
# Use make MOD=vendor to build with sources from the vendor directory instead.
MOD ?= readonly

## Adds a '-dirty' suffix to version string if there are uncommitted changes
changes := $(shell git status --porcelain)
ifeq ($(changes),)
	VERSION := $(TAG)
else
	VERSION := $(TAG)-dirty
endif

# Use the Go module mirror (https://blog.golang.org/module-mirror-launch).
# This speeds up build time and protects against disappearing dependencies.
ifeq ($(shell (go env GOPROXY)),)
       export GOPROXY=https://proxy.golang.org
endif


LDFLAGS := "-X github.com/kinvolk/lokoctl/cli/cmd.version=$(VERSION) -extldflags '-static'"

.NOTPARALLEL:

.PHONY: build
build: update-lk-submodule update-assets build-slim

.PHONY: update-lk-submodule
update-lk-submodule:
	git submodule update --init

.PHONY: update-assets
update-assets:
	GO111MODULE=on go generate ./...

.PHONY: build-slim
# Once we change CI code to build outside GOPATH, GO111MODULE can be removed, so
# we rely on defaults.
build-slim:
	CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build \
		-mod=$(MOD) \
		-ldflags $(LDFLAGS) \
		-buildmode=exe \
		github.com/kinvolk/lokoctl

.PHONY: test
test: check-go-format run-unit-tests

GOFORMAT_FILES := $(shell find . -name '*.go' | grep -v '^./vendor')

.PHONY: check-go-format
## Exits with an error if there are files whose formatting differs from gofmt's
check-go-format:
	@./scripts/go-lint ${GOFORMAT_FILES}

.PHONY: run-unit-tests
run-unit-tests:
	go test ./...

.PHONY: format-go-code
## Formats any go file that differs from gofmt's style
format-go-code:
	@gofmt -s -l -w ${GOFORMAT_FILES}

.PHONY: all
all: build test

.PHONY: install
install: update-lk-submodule update-assets install-slim

.PHONY: install-slim
# Once we change CI code to build outside GOPATH, GO111MODULE can be removed,
# so we rely on defaults.
install-slim:
	CGO_ENABLED=0 GOOS=linux GO111MODULE=on go install \
		-mod=$(MOD) \
		-ldflags $(LDFLAGS) \
		-buildmode=exe

.PHONY: install-packr2
install-packr2:
	echo "This target has been removed. This is here only to satisfy CI and will be removed later."

.PHONY: update
update: update-dependencies tidy vendor

.PHONY: update-dependencies
update-dependencies:
	GO111MODULE=on go get -u

.PHONY: tidy
tidy:
	GO111MODULE=on go mod tidy

.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor
