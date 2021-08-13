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

package rook

import (
	"fmt"
	"sort"

	helmcontrollerapi "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/version"
)

const (
	// Name represents Rook component name as it should be referenced in function calls
	// and in configuration.
	Name = "rook"
)

type component struct {
	Namespace                string            `hcl:"namespace,optional"`
	NodeSelector             util.NodeSelector `hcl:"node_selector,optional"`
	NodeSelectorRaw          string
	RookNodeAffinity         string
	Tolerations              []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw           string
	AgentTolerationKey       string            `hcl:"agent_toleration_key,optional"`
	AgentTolerationEffect    string            `hcl:"agent_toleration_effect,optional"`
	DiscoverTolerationKey    string            `hcl:"discover_toleration_key,optional"`
	DiscoverTolerationEffect string            `hcl:"discover_toleration_effect,optional"`
	EnableMonitoring         bool              `hcl:"enable_monitoring,optional"`
	CSIPluginNodeSelector    util.NodeSelector `hcl:"csi_plugin_node_selector,optional"`
	CSIPluginNodeAffinity    string
	CSIPluginTolerations     []util.Toleration `hcl:"csi_plugin_toleration,block"`
	CSIPluginTolerationsRaw  string
}

// NewConfig returns new Rook component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{
		Namespace: "rook",
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) generateValues() (string, error) {
	var err error

	c.TolerationsRaw, err = util.RenderTolerations(c.Tolerations)
	if err != nil {
		return "", fmt.Errorf("rendering tolerations: %w", err)
	}

	c.CSIPluginTolerationsRaw, err = util.RenderTolerations(c.CSIPluginTolerations)
	if err != nil {
		return "", fmt.Errorf("rendering CSI tolerations: %w", err)
	}

	c.NodeSelectorRaw, err = c.NodeSelector.Render()
	if err != nil {
		return "", fmt.Errorf("rendering node selector: %w", err)
	}

	c.RookNodeAffinity = convertNodeSelector(c.NodeSelector)
	c.CSIPluginNodeAffinity = convertNodeSelector(c.CSIPluginNodeSelector)

	return template.Render(chartValuesTmpl, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	values, err := c.generateValues()
	if err != nil {
		return nil, fmt.Errorf("rendering values template: %w", err)
	}

	// Generate YAML for the Rook operator deployment.
	renderedFiles, err := util.RenderChart(helmChart, Name, c.Metadata().Namespace.Name, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: c.Namespace,
		},
	}
}

// convertNodeSelector converts the key value pair in the map to the format:
// key1=value1; key2=value2;
func convertNodeSelector(m map[string]string) string {
	var ret string

	// Convert map to slice and sort them
	// by keys.
	keys := []string{}

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		ret += fmt.Sprintf("%s=%s; ", k, m[k])
	}

	return ret
}

func (c *component) GenerateHelmRelease() (*helmcontrollerapi.HelmRelease, error) {
	valuesYaml, err := c.generateValues()
	if err != nil {
		return nil, fmt.Errorf("rendering values template: %w", err)
	}

	values, err := k8syaml.YAMLToJSON([]byte(valuesYaml))
	if err != nil {
		return nil, fmt.Errorf("converting YAML to JSON: %w", err)
	}

	return &helmcontrollerapi.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: "flux-system",
		},
		Spec: helmcontrollerapi.HelmReleaseSpec{
			Chart: helmcontrollerapi.HelmChartTemplate{
				Spec: helmcontrollerapi.HelmChartTemplateSpec{
					Chart: components.ComponentsPath + Name,
					SourceRef: helmcontrollerapi.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: "lokomotive-" + version.Version,
					},
				},
			},
			ReleaseName: Name,
			Install: &helmcontrollerapi.Install{
				CRDs:            helmcontrollerapi.CreateReplace,
				CreateNamespace: true,
				Remediation: &helmcontrollerapi.InstallRemediation{
					Retries: -1,
				},
			},
			Upgrade: &helmcontrollerapi.Upgrade{
				CRDs: helmcontrollerapi.CreateReplace,
			},
			Interval:        components.FluxInstallInterval,
			Timeout:         &components.FluxInstallTimeout,
			TargetNamespace: c.Namespace,
			Values: &apiextensionsv1.JSON{
				Raw: values,
			},
		},
	}, nil
}
