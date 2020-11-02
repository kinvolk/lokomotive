// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"fmt"

	"github.com/kinvolk/lokomotive/pkg/components"
	awsebscsidriver "github.com/kinvolk/lokomotive/pkg/components/aws-ebs-csi-driver"
	certmanager "github.com/kinvolk/lokomotive/pkg/components/cert-manager"
	clusterautoscaler "github.com/kinvolk/lokomotive/pkg/components/cluster-autoscaler"
	"github.com/kinvolk/lokomotive/pkg/components/contour"
	"github.com/kinvolk/lokomotive/pkg/components/dex"
	externaldns "github.com/kinvolk/lokomotive/pkg/components/external-dns"
	flatcarlinuxupdateoperator "github.com/kinvolk/lokomotive/pkg/components/flatcar-linux-update-operator"
	"github.com/kinvolk/lokomotive/pkg/components/gangway"
	"github.com/kinvolk/lokomotive/pkg/components/httpbin"
	inspektorgadget "github.com/kinvolk/lokomotive/pkg/components/inspektor-gadget"
	istiooperator "github.com/kinvolk/lokomotive/pkg/components/istio-operator"
	"github.com/kinvolk/lokomotive/pkg/components/linkerd"
	"github.com/kinvolk/lokomotive/pkg/components/metallb"
	metricsserver "github.com/kinvolk/lokomotive/pkg/components/metrics-server"
	openebsoperator "github.com/kinvolk/lokomotive/pkg/components/openebs-operator"
	openebsstorageclass "github.com/kinvolk/lokomotive/pkg/components/openebs-storage-class"
	"github.com/kinvolk/lokomotive/pkg/components/prometheus-operator"
	"github.com/kinvolk/lokomotive/pkg/components/rook"
	rookceph "github.com/kinvolk/lokomotive/pkg/components/rook-ceph"
	"github.com/kinvolk/lokomotive/pkg/components/velero"
	webui "github.com/kinvolk/lokomotive/pkg/components/web-ui"
)

func componentsConfigs() map[string]components.Component {
	return map[string]components.Component{
		awsebscsidriver.Name:            awsebscsidriver.NewConfig(),
		certmanager.Name:                certmanager.NewConfig(),
		clusterautoscaler.Name:          clusterautoscaler.NewConfig(),
		contour.Name:                    contour.NewConfig(),
		dex.Name:                        dex.NewConfig(),
		externaldns.Name:                externaldns.NewConfig(),
		flatcarlinuxupdateoperator.Name: flatcarlinuxupdateoperator.NewConfig(),
		gangway.Name:                    gangway.NewConfig(),
		httpbin.Name:                    httpbin.NewConfig(),
		inspektorgadget.Name:            inspektorgadget.NewConfig(),
		istiooperator.Name:              istiooperator.NewConfig(),
		linkerd.Name:                    linkerd.NewConfig(),
		metallb.Name:                    metallb.NewConfig(),
		metricsserver.Name:              metricsserver.NewConfig(),
		openebsoperator.Name:            openebsoperator.NewConfig(),
		openebsstorageclass.Name:        openebsstorageclass.NewConfig(),
		prometheus.Name:                 prometheus.NewConfig(),
		rook.Name:                       rook.NewConfig(),
		rookceph.Name:                   rookceph.NewConfig(),
		velero.Name:                     velero.NewConfig(),
		webui.Name:                      webui.NewConfig(),
	}
}

// AvailableComponents returns list of valid component names.
func AvailableComponents() []string {
	c := []string{}

	for n := range componentsConfigs() {
		c = append(c, n)
	}

	return c
}

func componentConfig(name string) (components.Component, error) {
	c, ok := componentsConfigs()[name]
	if !ok {
		return nil, fmt.Errorf("no component with name %q found", name)
	}

	return c, nil
}
