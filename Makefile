.PHONY: build
build:
	go build -o lokoctl github.com/kinvolk/lokoctl/cli

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

.PHONY: bindata-ingressnginx
bindata-ingressnginx:
	go-bindata -pkg ingressnginx -o pkg/component/ingressnginx/bindata.go manifests/nginx-ingress manifests/nginx-ingress/rbac

.PHONY: bindata-networkpolicy
bindata-networkpolicy:
	go-bindata -pkg networkpolicy -o pkg/component/networkpolicy/bindata.go manifests/default-network-policies/deny-metadata-access.yaml

.PHONY: bindata-aws
bindata-aws:
	./scripts/bindata-aws

.PHONY: bindata
# make sure that `format-go-code` target is always the last one to run
bindata: | bindata-ingressnginx bindata-networkpolicy bindata-aws format-go-code

.PHONY: all
all: getbindata bindata build test
