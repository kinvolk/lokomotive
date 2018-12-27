TAG := `git describe --tags --always`
VERSION :=

## Adds a '-dirty' suffix to version string if there are uncommitted changes
changes := $(shell git status --porcelain)
ifeq ($(changes),)
	VERSION := $(TAG)
else
	VERSION := $(TAG)-dirty
endif

LDFLAGS := "-X github.com/kinvolk/lokoctl/cli/cmd.version=$(VERSION) -extldflags '-static'"

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build \
		-ldflags $(LDFLAGS) \
		-o lokoctl \
		github.com/kinvolk/lokoctl/cli

.PHONY: packr2
packr2:
	cd cli && packr2

.PHONY: build-fat
build-fat: packr2 build
	cd cli && packr2 clean

.PHONY: test
test: check-go-format

GOFORMAT_FILES := $(shell find . -name '*.go' | grep -v vendor)

.PHONY: check-go-format
## Exits with an error if there are files whose formatting differs from gofmt's
check-go-format:
	@./scripts/go-lint ${GOFORMAT_FILES}

.PHONY: format-go-code
## Formats any go file that differs from gofmt's style
format-go-code:
	@gofmt -s -l -w ${GOFORMAT_FILES}

.PHONY: getbindata
getbindata:
	go get -u github.com/twitter/go-bindata/...

.PHONY: bindata-installer
bindata-installer:
	./scripts/bindata-installer

.PHONY: bindata-components
bindata-components:
	./scripts/bindata-components

.PHONY: bindata
# make sure that `format-go-code` target is always the last one to run
bindata: | bindata-installer bindata-components format-go-code

.PHONY: all
all: getbindata bindata build test
