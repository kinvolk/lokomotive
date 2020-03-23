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

// +build packet
// +build poste2e

package calico

import (
	"context"
	"testing"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	client "github.com/projectcalico/libcalico-go/lib/clientv3"
	"github.com/projectcalico/libcalico-go/lib/options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kinvolk/lokomotive/test/components/util"
)

func TestHostEndpoints(t *testing.T) {
	// Build Calico client.
	cac := apiconfig.NewCalicoAPIConfig()
	cac.Spec.DatastoreType = apiconfig.Kubernetes
	cac.Spec.Kubeconfig = util.KubeconfigPath(t)

	c, err := client.New(*cac)
	if err != nil {
		t.Fatalf("failed creating Calico client: %v", err)
	}

	// Build list of nodes which has associated HostEndpoint object.
	hostEndpointList, err := c.HostEndpoints().List(context.TODO(), options.ListOptions{})
	if err != nil {
		t.Fatalf("failed getting hostendpoint objects: %v", err)
	}

	endpoints := map[string]struct{}{}

	for _, v := range hostEndpointList.Items {
		endpoints[v.Spec.Node] = struct{}{}
	}

	cs := util.CreateKubeClient(t)

	nodes, err := cs.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("failed getting list of nodes in the cluster: %v", err)
	}

	for _, v := range nodes.Items {
		if _, ok := endpoints[v.Name]; !ok {
			t.Errorf("no HostEndpoint object found for node %q", v.Name)
		}
	}
}
