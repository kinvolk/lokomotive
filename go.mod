module github.com/kinvolk/lokomotive

go 1.15

require (
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20210801061803-8e322dfb79c4 // indirect
	github.com/fluxcd/helm-controller/api v0.11.1
	github.com/fluxcd/source-controller/api v0.15.3
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-openapi/spec v0.19.7 // indirect
	github.com/go-openapi/swag v0.19.8 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.5.5
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl/v2 v2.7.2
	github.com/hpcloud/tail v1.0.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/linkerd/linkerd2 v0.5.1-0.20210517230931-5535e9c4edda
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/packethost/packngo v0.2.0
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/prometheus/alertmanager v0.20.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/shirou/gopsutil v2.20.2+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/zclconf/go-cty v1.7.0
	gopkg.in/ini.v1 v1.54.0 // indirect
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.0
	sigs.k8s.io/yaml v1.2.0
)

// There is a big confusion how to use Docker in go modules. This points to v19.03.5.
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725

// With v0.2.0 package has been renames, so until all dependencies are updated to use new import name,
// we need to use older version.
//
// See: https://github.com/moby/spdystream/releases/tag/v0.2.0
replace github.com/docker/spdystream => github.com/moby/spdystream v0.1.0

// Borrowed from Helm.
replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d

//replace github.com/deislabs/oras => github.com/deislabs/oras v0.8.1

// Without this, kubectl dependency does not build.
replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2

// Use patched version of helm until upstream PRs get merged:
//
// https://github.com/helm/helm/pull/7405
// https://github.com/helm/helm/pull/7431
replace helm.sh/helm/v3 => github.com/kinvolk/helm/v3 v3.6.3-patched

// With v0.19.9+ kustomize no longer builds.
replace github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.8

// Force latest version of client-go, so 'v11.0.0+incompatible' does not get pulled on update.
replace k8s.io/client-go => k8s.io/client-go v0.21.3
