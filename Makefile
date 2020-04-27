TAG := `git describe --tags --always`
VERSION :=
# MOD can either be "readonly" or "vendor".
# The default is "vendor" which uses checked out modules for building.
# Use make MOD=readonly to build with sources from the Go module cache instead.
MOD ?= vendor
DOCS_DIR ?= docs/cli

ALL_BUILD_TAGS := "aws,packet,e2e,disruptivee2e,poste2e"

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

LDFLAGS := "-X github.com/kinvolk/lokomotive/pkg/version.Version=$(VERSION) -extldflags '-static'"

.NOTPARALLEL:

.PHONY: build
build: update-assets build-slim

.PHONY: build-in-docker
build-in-docker:
	# increase ulimit to workaround https://github.com/golang/go/issues/37436
	docker run --ulimit memlock=1024000 --rm -ti -v $(shell pwd):/usr/src/lokomotive -w /usr/src/lokomotive golang:1.14 sh -c "make"

.PHONY: build-test
build-test:
	go test -run=nonexistent -mod=$(MOD) -tags=$(ALL_BUILD_TAGS) -covermode=atomic -buildmode=exe -v ./... > /dev/null

.PHONY: all
all: build build-test test lint

.PHONY: update-assets
update-assets:
	GO111MODULE=on go generate -mod=$(MOD) ./...

.PHONY: build-slim
# Once we change CI code to build outside GOPATH, GO111MODULE can be removed, so
# we rely on defaults.
build-slim:
	CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build \
		-mod=$(MOD) \
		-ldflags $(LDFLAGS) \
		-buildmode=exe \
		-o lokoctl \
		github.com/kinvolk/lokomotive/cmd/lokoctl

.PHONY: test
test: run-unit-tests

.PHONY: lint
lint: build-slim build-test
	# Note: Make sure that you run `git config diff.noprefix false` in this repo
	# See this issue for more details: https://github.com/golangci/golangci-lint/issues/948
	golangci-lint run --enable-all --disable=gomnd,godox,gochecknoglobals --max-same-issues=0 --max-issues-per-linter=0 --build-tags $(ALL_BUILD_TAGS) --new-from-rev=$$(git merge-base $$(cat .git/resource/base_sha 2>/dev/null || echo "master") HEAD) --modules-download-mode=$(MOD) --timeout=5m --exclude-use-default=false ./...

.PHONY: lint-docker
lint-docker:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.23.8 make lint

GOFORMAT_FILES := $(shell find . -name '*.go' | grep -v '^./vendor')

.PHONY: run-unit-tests
run-unit-tests:
	go test -mod=$(MOD) -covermode=atomic -buildmode=exe -v ./...

.PHONY: format-go-code
## Formats any go file that differs from gofmt's style
format-go-code:
	@gofmt -s -l -w ${GOFORMAT_FILES}

kubeconfig := $(KUBECONFIG)
## Following kubeconfig path is only valid from CI
ifeq ($(RUN_FROM_CI),"true")
	kubeconfig := "${HOME}/lokoctl-assets/cluster-assets/auth/kubeconfig"
endif

.PHONY: run-e2e-tests
run-e2e-tests:
	KUBECONFIG=${kubeconfig} go test -mod=$(MOD) -tags="$(platform),e2e" -covermode=atomic -buildmode=exe -v ./test/...
	# Test if the metrics are actually being scraped
	KUBECONFIG=${kubeconfig} PLATFORM=${platform} go test -mod=$(MOD) -tags="$(platform),poste2e" -covermode=atomic -buildmode=exe -v ./test/...
	# This is a test that should be run in the end to reduce the disruption to other tests because
	# it will delete a node.
	KUBECONFIG=${kubeconfig} go test -mod=$(MOD) -tags="$(platform),disruptivee2e" -covermode=atomic -buildmode=exe -v ./test/...

.PHONY: all
all: build test

.PHONY: install
install: update-assets install-slim

.PHONY: install-slim
# Once we change CI code to build outside GOPATH, GO111MODULE can be removed,
# so we rely on defaults.
install-slim:
	CGO_ENABLED=0 GOOS=linux GO111MODULE=on go install \
		-mod=$(MOD) \
		-ldflags $(LDFLAGS) \
		-buildmode=exe \
		./cmd/lokoctl

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

.PHONY: docker-build
docker-build:
	docker build -t kinvolk/lokomotive .

.PHONY: docker-vendor
docker-vendor: docker-build
	docker run --rm -ti -v $(shell pwd):/usr/src/lokomotive kinvolk/lokomotive sh -c "make vendor && chown -R $(shell id -u):$(shell id -g) vendor"

.PHONY: docker-update-assets
docker-update-assets: docker-build
	docker run --rm -ti -v $(shell pwd):/usr/src/lokomotive kinvolk/lokomotive sh -c "make update-assets && chown -R $(shell id -u):$(shell id -g) assets"

.PHONY: docker-update-dependencies
docker-update-dependencies: docker-build
	docker run --rm -ti -v $(shell pwd):/usr/src/lokomotive kinvolk/lokomotive sh -c "make update-dependencies && chown $(shell id -u):$(shell id -g) go.mod go.sum"

.PHONY: docs
docs:
	GO111MODULE=on go run -mod=$(MOD) -buildmode=exe cli/cmd/document/main.go $(DOCS_DIR)

.PHONY: build-and-publish-release
build-and-publish-release: SHELL:=/bin/bash
build-and-publish-release:
	goreleaser --release-notes <(./scripts/print-version-changelog.sh)
