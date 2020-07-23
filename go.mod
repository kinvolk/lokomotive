module github.com/kinvolk/lokomotive

go 1.12

require (
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/containerd/cgroups v0.0.0-20200308110149-6c3dec43a103 // indirect
	github.com/containerd/containerd v1.3.3 // indirect
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c // indirect
	github.com/docker/cli v0.0.0-20200312141509-ef2f64abbd37 // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/emicklei/go-restful v2.12.0+incompatible // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-openapi/spec v0.19.7 // indirect
	github.com/go-openapi/swag v0.19.8 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl/v2 v2.3.0
	github.com/hpcloud/tail v1.0.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/linkerd/linkerd2 v0.5.1-0.20200623171206-83ae0ccf0f1a
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/packethost/packngo v0.2.0
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/alertmanager v0.20.0
	github.com/prometheus/client_golang v1.5.0
	github.com/prometheus/procfs v0.0.10 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/shirou/gopsutil v2.20.2+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.6.2
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.3.1
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/tools v0.0.0-20200129045341-207d3de1faaf // indirect
	google.golang.org/genproto v0.0.0-20200312145019-da6875a35672 // indirect
	google.golang.org/grpc v1.28.0 // indirect
	gopkg.in/ini.v1 v1.54.0 // indirect
	gotest.tools/v3 v3.0.2 // indirect
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.18.0
	k8s.io/apimachinery v0.18.0
	k8s.io/client-go v0.18.0
	k8s.io/kube-openapi v0.0.0-20200204173128-addea2498afe // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)

// There is a big confusion how to use Docker in go modules. This points to v19.03.5.
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725

// Without this, kubectl dependency does not build.
replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2

// Use patched version of helm until upstream PRs get merged:
//
// https://github.com/helm/helm/pull/7405
// https://github.com/helm/helm/pull/7431
replace helm.sh/helm/v3 => github.com/kinvolk/helm/v3 v3.2.2-0.20200526121938-305e0b796fc9

// This module has been renamed, which causes confusion now.
// https://github.com/kubernetes/kubernetes/issues/88183
// https://github.com/googleapis/gnostic/commit/896953e6749863beec38e27029c804e88c3144b8
replace github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0

// To address https://github.com/etcd-io/etcd/issues/11563. Required until new version of etcd is released with the fix.
replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

// Taken from Helm upstream
replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible
