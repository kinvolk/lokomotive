module github.com/kinvolk/lokoctl

go 1.12

require (
	cloud.google.com/go v0.51.0 // indirect
	cloud.google.com/go/storage v1.5.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Microsoft/hcsshim v0.8.7 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/aws/aws-sdk-go v1.28.2 // indirect
	github.com/bmatcuk/doublestar v1.2.2 // indirect
	github.com/containerd/cgroups v0.0.0-20200113070643-7347743e5d1e // indirect
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/containerd/continuity v0.0.0-20200107194136-26c1120b8d41 // indirect
	github.com/deislabs/oras v0.8.0 // indirect
	github.com/docker/cli v0.0.0-20200113155311-34d848623701 // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-openapi/spec v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.6 // indirect
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/hcl/v2 v2.3.0
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80 // indirect
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/terraform v0.12.19
	github.com/hashicorp/terraform-svchost v0.0.0-20191119180714-d2e4933b9136 // indirect
	github.com/hpcloud/tail v1.0.0
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/miekg/dns v1.1.15 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/packethost/packngo v0.2.0
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/posener/complete v1.2.3 // indirect
	github.com/prometheus/alertmanager v0.20.0
	github.com/prometheus/client_golang v1.3.0 // indirect
	github.com/prometheus/common v0.8.0 // indirect
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/shirou/gopsutil v2.19.12+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.6.1
	github.com/ulikunitz/xz v0.5.6 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zclconf/go-cty v1.2.1
	golang.org/x/crypto v0.0.0-20200109152110-61a87790db17 // indirect
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200113162924-86b910548bc1 // indirect
	golang.org/x/tools v0.0.0-20200115044656-831fdb1e1868 // indirect
	google.golang.org/genproto v0.0.0-20200113173426-e1de0a7b01eb // indirect
	gopkg.in/ini.v1 v1.51.1 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect; indirect;
	helm.sh/helm/v3 v3.0.2
	k8s.io/api v0.17.1
	k8s.io/apiextensions-apiserver v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.17.1 // indirect
	k8s.io/utils v0.0.0-20200109141947-94aeca20bf09 // indirect
	sigs.k8s.io/yaml v1.1.0
)

// There is a big confusion how to use Docker in go modules. This points to v19.03.5.
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191113042239-ea84732a7725

// Force upgrade client-go to latest version, otherwise go get complains about incompatible versions.
replace k8s.io/client-go => k8s.io/client-go v0.17.1

// Without this, helm dependency does not build.
replace github.com/deislabs/oras => github.com/deislabs/oras v0.7.0

// Without this, kubectl dependency does not build.
replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2
