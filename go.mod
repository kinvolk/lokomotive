module github.com/kinvolk/lokomotive

go 1.13

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/hpcloud/tail v1.0.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/packethost/packngo v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/alertmanager v0.21.0
	github.com/prometheus/client_golang v1.7.1
	github.com/shirou/gopsutil v2.20.7+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20200627165143-92b8a710ab6c
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/zclconf/go-cty v1.5.1
	helm.sh/helm/v3 v3.2.4
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)

// Use patched version of helm until upstream PRs get merged:
//
// https://github.com/helm/helm/pull/7405
// https://github.com/helm/helm/pull/7431
replace helm.sh/helm/v3 => github.com/kinvolk/helm/v3 v3.2.2-0.20200526121938-305e0b796fc9
