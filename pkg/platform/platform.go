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

package platform

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v3/pkg/chart"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/helm"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/kinvolk/lokomotive/pkg/version"
)

const (
	// NetworkMTU is the default host network MTU.
	NetworkMTU = 1500

	// ConntrackMaxPerCore is the default conntrack table size per core inherited from upstream kube-proxy.
	ConntrackMaxPerCore = 32768

	// KubernetesChartName is the expected name for the Kubernetes Helm chart.
	KubernetesChartName = "kubernetes"

	// KubeletChartName is the expected name for the Kubelet Helm chart.
	KubeletChartName = "kubelet"
)

// CommonControlPlaneCharts returns a list of control plane Helm charts to be deployed for all
// platforms.
func CommonControlPlaneCharts(includeKubeletChart bool) []helm.LokomotiveChart {
	charts := []helm.LokomotiveChart{
		{
			Name:      "bootstrap-secrets",
			Namespace: "kube-system",
		},
		{
			Name:      "pod-checkpointer",
			Namespace: "kube-system",
		},
		{
			Name:      "kube-apiserver",
			Namespace: "kube-system",
		},
		{
			Name:      KubernetesChartName,
			Namespace: "kube-system",
		},
		{
			Name:      "calico",
			Namespace: "kube-system",
		},
		{
			Name:      "lokomotive",
			Namespace: "lokomotive-system",
		},
	}

	if includeKubeletChart {
		charts = append(charts, helm.LokomotiveChart{
			Name:      KubeletChartName,
			Namespace: "kube-system",
		})
	}

	return charts
}

// ControlPlaneChart is a convenience function which returns a pointer to a chart.Chart
// representing the control plane element named name.
func ControlPlaneChart(name string) (*chart.Chart, error) {
	p := filepath.Join(assets.ControlPlaneSource, name)

	return helm.ChartFromAssets(p)
}

// Platform describes single environment, where cluster can be installed
type Platform interface {
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	Apply(*terraform.Executor) error
	Destroy(*terraform.Executor) error
	Initialize(*terraform.Executor) error
	Meta() Meta
}

// PlatformWithPostApplyHook runs code after Terraform finishes applying. This allows
// running sanity checks on the newly created cluster. Implementing this
// interface is optional for platforms.
type PlatformWithPostApplyHook interface { //nolint:golint
	PostApplyHook(kubeconfig []byte) error
}

// Meta is a generic information format about the platform.
type Meta struct {
	AssetDir           string
	ExpectedNodes      int
	Managed            bool
	ControlplaneCharts []helm.LokomotiveChart
}

// platforms is a collection where all platforms gets automatically registered
var platforms map[string]Platform

// initialize package's global variable when package is imported
func init() {
	platforms = make(map[string]Platform)
}

// Register adds platform into internal map
func Register(name string, p Platform) {
	if _, exists := platforms[name]; exists {
		panic(fmt.Sprintf("platform with name %q registered already", name))
	}
	platforms[name] = p
}

// GetPlatform returns platform based on the name
func GetPlatform(name string) (Platform, error) {
	platform, exists := platforms[name]
	if !exists {
		return nil, fmt.Errorf("no platform with name %q found", name)
	}
	return platform, nil
}

// AppendVersionTag appends the lokoctl-version tag to a given tags map.
func AppendVersionTag(tags *map[string]string) {
	if tags == nil {
		return
	}

	if *tags == nil {
		*tags = make(map[string]string)
	}

	if version.Version != "" {
		(*tags)["lokoctl-version"] = version.Version
	}
}
