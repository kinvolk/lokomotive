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

package flatcarlinuxupdateoperator

import (
	"fmt"

	helmcontrollerapi "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/version"
)

const (
	// Name represents Flatcar Linux Update Operator component name as it should be referenced in function calls
	// and in configuration.
	Name = "flatcar-linux-update-operator"

	namespace = "reboot-coordinator"
)

// NewConfig returns new Flatcar Linux Update Operator component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{}
}

type component struct{}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		// This component has no configuration, so don't complain when there is no configuration defined.
		return nil
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("loading chart from assets: %w", err)
	}

	return util.RenderChart(helmChart, Name, c.Metadata().Namespace.Name, "")
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: namespace,
		},
	}
}

func (c *component) GenerateHelmRelease() (*helmcontrollerapi.HelmRelease, error) {
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
				CreateNamespace: false,
				Remediation: &helmcontrollerapi.InstallRemediation{
					Retries: -1,
				},
			},
			Upgrade: &helmcontrollerapi.Upgrade{
				CRDs: helmcontrollerapi.CreateReplace,
			},
			Interval:        components.FluxInstallInterval,
			Timeout:         &components.FluxInstallTimeout,
			TargetNamespace: namespace,
		},
	}, nil
}
