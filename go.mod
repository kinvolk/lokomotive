module github.com/kinvolk/lokomotive

go 1.15

require (
	github.com/elazarl/goproxy/ext v0.0.0-20210801061803-8e322dfb79c4 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.5.4
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/hcl/v2 v2.7.2
	github.com/hpcloud/tail v1.0.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/linkerd/linkerd2 v0.5.1-0.20210517230931-5535e9c4edda
	github.com/mitchellh/go-homedir v1.1.0
	github.com/packethost/packngo v0.2.0
	github.com/prometheus/alertmanager v0.20.0
	github.com/prometheus/client_golang v1.7.1
	github.com/shirou/gopsutil/v3 v3.21.7
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/zclconf/go-cty v1.7.0
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.4
	k8s.io/apiextensions-apiserver v0.21.4 // indirect
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v0.21.4
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
replace k8s.io/client-go => k8s.io/client-go v0.21.4
