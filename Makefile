TAG := `git describe --tags --always`
VERSION :=
# MOD can either be "readonly" or "vendor".
# The default is "vendor" which uses checked out modules for building.
# Use make MOD=readonly to build with sources from the Go module cache instead.
MOD ?= vendor
DOCS_DIR ?= docs/cli

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

.PHONY: build-test
build-test:
	go test -run=nonexistent -mod=$(MOD) -tags="aws,packet,e2e,disruptive-e2e" -covermode=atomic -buildmode=exe -v ./...

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
lint:
	golangci-lint run --enable-all --disable=godox --max-same-issues=0 --max-issues-per-linter=0 --build-tags aws,packet,e2e,disruptive-e2e --new-from-rev=$$(git merge-base $$(cat .git/resource/base_sha 2>/dev/null || echo "master") HEAD) --modules-download-mode=$(MOD) --timeout=5m --exclude-use-default=false ./...

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
	KUBECONFIG=${kubeconfig} go test -mod=$(MOD) -tags="$(platform),e2e" -covermode=atomic -buildmode=exe -v ./...
	# This is a test that should be run in the end to reduce the disruption to other tests because
	# it will delete a node.
	KUBECONFIG=${kubeconfig} go test -mod=$(MOD) -tags="$(platform),disruptive-e2e" -covermode=atomic -buildmode=exe -v ./...

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

.PHONY: docker-build
docker-build:
	docker build -t kinvolk/lokomotive .

.PHONY: docker-vendor
docker-vendor: docker-build
	docker run --rm -ti -v $(pwd):/usr/src/lokoctl kinvolk/lokomotive sh -c "make vendor && chown -R $(shell id -u):$(shell id -g) vendor"

.PHONY: docker-update-assets
docker-update-assets: docker-build
	docker run --rm -ti -v $(pwd):/usr/src/lokoctl kinvolk/lokomotive sh -c "make update-assets && chown -R $(shell id -u):$(shell id -g) assets"

.PHONY: docker-update-dependencies
docker-update-dependencies: docker-build
	docker run --rm -ti -v $(pwd):/usr/src/lokoctl kinvolk/lokomotive sh -c "make update-dependencies && chown $(shell id -u):$(shell id -g) go.mod go.sum"

.PHONY: docs
docs:
	GO111MODULE=on go run -mod=$(MOD) -buildmode=exe cli/cmd/document/main.go $(DOCS_DIR)
