.PHONY: build
build:
	go build -o lokoctl github.com/kinvolk/lokoctl/cli

.PHONY: test
test: check-go-format

GOFORMAT_FILES := $(shell find . -name '*.go' | grep -v vendor)

.PHONY: check-go-format
## Exits with an error if there are files whose formatting differs from gofmt's
check-go-format:
	@gofmt -s -l ${GOFORMAT_FILES} 2>&1 \
		| tee /tmp/gofmt-errors \
		| read \
	&& echo "ERROR: These files differ from gofmt's style (run 'make format-go-code' to fix this):" \
	&& cat /tmp/gofmt-errors \
	&& exit 1 \
	|| true

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

.PHONY: bindata
bindata: bindata-ingressnginx

.PHONY: all
all: getbindata bindata build test
