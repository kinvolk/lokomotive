// Copyright 2021 The Lokomotive Authors
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

package platform_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/platform"
)

func TestAppendVersionTagUninitializedMap(t *testing.T) {
	var f map[string]string

	platform.AppendVersionTag(&f)

	if len(f) == 0 {
		t.Fatalf("should append version tag to uninitialized map")
	}
}

func TestAppendVersionTagIgnoreNil(t *testing.T) {
	platform.AppendVersionTag(nil)
}

func TestAppendVersionTag(t *testing.T) {
	f := map[string]string{
		"foo": "bar",
	}

	platform.AppendVersionTag(&f)

	if len(f) != 2 {
		t.Fatalf("should append version tag to existing map")
	}
}

func TestAtLeastOneCommonControlplaneChartIsDefined(t *testing.T) {
	if len(platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: false})) == 0 {
		t.Fatalf("There should be at least one common controlplane chart defined")
	}
}

func TestBootstrapSecretsAreUpdatedFirst(t *testing.T) {
	if platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: true})[0].Name != "bootstrap-secrets" {
		t.Fatalf("Bootstrap-secrets should be updated first to allow nodes to proceed with bootstrapping process")
	}
}

func TestPodCheckpointerIsUpdatedBeforeKubeApiserver(t *testing.T) {
	podCheckpointerFound := false

	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: true}) {
		if c.Name == "pod-checkpointer" {
			podCheckpointerFound = true
		}

		if c.Name == "kube-apiserver" && !podCheckpointerFound {
			t.Fatalf("Pod-checkpointer should be updated before kube-apiserver to ensure API availability if " +
				"upgrade process of kube-apiserver fails for some reason.")
		}
	}
}

func TestKubeApiserverIsUpdatedBeforeOtherKubernetesComponents(t *testing.T) {
	kubeAPIServerFound := false

	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: true}) {
		if c.Name == "kube-apiserver" {
			kubeAPIServerFound = true
		}

		if c.Name == platform.KubernetesChartName && !kubeAPIServerFound {
			t.Fatalf("Kube-apiserver must be updated before other Kubernetes components to ensure " +
				"Kubernetes version skew support policy")
		}
	}
}

func TestCalicoIsUpdatedAfterKubernetesComponents(t *testing.T) {
	kubernetesFound := false

	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: true}) {
		if c.Name == platform.KubernetesChartName {
			kubernetesFound = true
		}

		if c.Name == "calico" && !kubernetesFound {
			t.Fatalf("Calico must be updated after Kubernetes component, as it requires functional kube-proxy " +
				"to converge")
		}
	}
}

func TestKubeletIsUpdatedAfterOtherKubernetesComponents(t *testing.T) {
	kubernetesFound := false

	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: true}) {
		if c.Name == platform.KubernetesChartName {
			kubernetesFound = true
		}

		if c.Name == platform.KubeletChartName && !kubernetesFound {
			t.Fatalf("kubelet must be updated after Kubernetes component to ensure Kubernetes version skew support policy")
		}
	}
}

func TestLokomotiveIsUpdatedAfterCalico(t *testing.T) {
	calicoFound := false

	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: true}) {
		if c.Name == "calico" {
			calicoFound = true
		}

		if c.Name == "lokomotive" && !calicoFound {
			t.Fatalf("Lokomotive must be updated after Calico, as it requires Pod networking to converge")
		}
	}
}

func TestKubeletIsExcludedFromUpdatesWhenNotRequested(t *testing.T) {
	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: false}) {
		if c.Name == platform.KubeletChartName {
			t.Fatalf("Kubelet should not be included in charts list when not requested")
		}
	}
}

func TestKubeletIsIncludedInCommonChartsWhenRequested(t *testing.T) {
	kubeletFound := false

	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{Kubelet: true}) {
		if c.Name == platform.KubeletChartName {
			kubeletFound = true

			break
		}
	}

	if !kubeletFound {
		t.Fatalf("Kubelet should be included in charts list when requested")
	}
}

func TestNodeLocalDNSIsExcludedFromUpdatesWhenNotRequested(t *testing.T) {
	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{NodeLocalDNS: false}) {
		if c.Name == platform.NodeLocalDNSChartName {
			t.Fatalf("NodeLocalDNS should not be included in charts list when not requested")
		}
	}
}

func TestNodeLocalDNSIsIncludedInCommonChartsWhenRequested(t *testing.T) {
	nodeLocalDNSFound := false

	for _, c := range platform.CommonControlPlaneCharts(platform.ControlPlanCharts{NodeLocalDNS: true}) {
		if c.Name == platform.NodeLocalDNSChartName {
			nodeLocalDNSFound = true

			break
		}
	}

	if !nodeLocalDNSFound {
		t.Fatalf("NodeLocalDNS should be included in charts list when requested")
	}
}
