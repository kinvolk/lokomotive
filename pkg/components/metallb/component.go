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

package metallb

import (
	"fmt"
	"time"

	api "github.com/fluxcd/helm-controller/api/v2beta1"
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
	// Name represents MetalLB component name as it should be referenced in function calls
	// and in configuration.
	Name = "metallb"

	namespace = "metallb-system"
)

type component struct {
	AddressPools            map[string][]string `hcl:"address_pools"`
	ControllerNodeSelectors map[string]string   `hcl:"controller_node_selectors,optional"`
	SpeakerNodeSelectors    map[string]string   `hcl:"speaker_node_selectors,optional"`
	ControllerTolerations   []util.Toleration   `hcl:"controller_toleration,block"`
	SpeakerTolerations      []util.Toleration   `hcl:"speaker_toleration,block"`
	ServiceMonitor          bool                `hcl:"service_monitor,optional"`

	ControllerTolerationsJSON string
	SpeakerTolerationsJSON    string
}

// NewConfig returns new MetalLB component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) generateValues() (string, error) {
	// Here are `nodeSelectors` and `tolerations` that are set by upstream. To make sure that we
	// don't miss them out we set them manually here. We cannot make these changes in the template
	// because we have parameterized these fields.
	if c.SpeakerNodeSelectors == nil {
		c.SpeakerNodeSelectors = map[string]string{}
	}
	// MetalLB only supports Linux, so force this selector, even if it's already specified by the
	// user.
	c.SpeakerNodeSelectors["beta.kubernetes.io/os"] = "linux"

	if c.ControllerNodeSelectors == nil {
		c.ControllerNodeSelectors = map[string]string{}
	}
	c.ControllerNodeSelectors["beta.kubernetes.io/os"] = "linux"
	c.ControllerNodeSelectors["node.kubernetes.io/master"] = ""

	c.ControllerTolerations = append(c.ControllerTolerations, util.Toleration{
		Effect: "NoSchedule",
		Key:    "node-role.kubernetes.io/master",
	})

	t, err := util.RenderTolerations(c.SpeakerTolerations)
	if err != nil {
		return "", fmt.Errorf("marshaling speaker tolerations: %w", err)
	}
	c.SpeakerTolerationsJSON = t

	t, err = util.RenderTolerations(c.ControllerTolerations)
	if err != nil {
		return "", fmt.Errorf("rendering controller tolerations: %w", err)
	}
	c.ControllerTolerationsJSON = t

	return template.Render(chartValuesTmpl, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	values, err := c.generateValues()
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	renderedFiles, err := util.RenderChart(helmChart, Name, c.Metadata().Namespace.Name, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) GenerateHelmRelease() (*api.HelmRelease, error) {
	valuesYaml, err := c.generateValues()
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

	values, err := k8syaml.YAMLToJSON([]byte(valuesYaml))
	if err != nil {
		return nil, fmt.Errorf("converting YAML to JSON: %w", err)
	}

	interval := time.Minute * 10

	return &api.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: "flux-system",
		},
		Spec: api.HelmReleaseSpec{
			Chart: api.HelmChartTemplate{
				Spec: api.HelmChartTemplateSpec{
					Chart: "./assets/charts/components/metallb/",
					SourceRef: api.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: "lokomotive-" + version.Version,
					},
				},
			},
			ReleaseName: Name,
			Install: &api.Install{
				CRDs:            api.CreateReplace,
				CreateNamespace: true,
			},
			Upgrade: &api.Upgrade{
				Force: true,
				CRDs:  api.CreateReplace,
			},
			Interval:        metav1.Duration{time.Minute},
			Timeout:         &metav1.Duration{interval},
			TargetNamespace: namespace,
			Values: &apiextensionsv1.JSON{
				Raw: values,
			},
		},
	}, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: namespace,
		},
	}
}
