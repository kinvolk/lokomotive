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

.NOTPARALLEL:

.PHONY: build
build: update-lk-submodule packr2 build-slim packr2-clean

.PHONY: update-lk-submodule
update-lk-submodule:
	git submodule update --init

.PHONY: packr2
packr2:
	cd pkg/components && packr2
	cd pkg/install && packr2

.PHONY: packr2-clean
packr2-clean:
	cd pkg/components && packr2 clean
	cd pkg/install && packr2 clean

.PHONY: build-slim
# Once we change CI code to build outside GOPATH, GO111MODULE can be removed, so
# we rely on defaults.
build-slim:
	CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build \
		-ldflags $(LDFLAGS) \
		-buildmode=exe \
		github.com/kinvolk/lokoctl

.PHONY: test
test: check-go-format run-unit-tests

GOFORMAT_FILES := $(shell find . -name '*.go')

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

.PHONY: install-packr2
install-packr2:
	# Once we change CI code to build outside GOPATH, GO111MODULE can be removed,
	# so we rely on defaults.
	GO111MODULE=on go get github.com/gobuffalo/packr/v2/packr2
