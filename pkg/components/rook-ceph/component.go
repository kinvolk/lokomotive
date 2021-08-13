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

package rookceph

import (
	"fmt"

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
	// Name represents Rook Ceph component name as it should be referenced in function calls
	// and in configuration.
	Name = "rook-ceph"
)

type component struct {
	Namespace       string              `hcl:"namespace,optional"`
	MonitorCount    int                 `hcl:"monitor_count,optional"`
	NodeAffinity    []util.NodeAffinity `hcl:"node_affinity,block"`
	NodeAffinityRaw string
	MetadataDevice  string            `hcl:"metadata_device,optional"`
	Tolerations     []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw  string
	StorageClass    *StorageClass `hcl:"storage_class,block"`
	EnableToolbox   bool          `hcl:"enable_toolbox,optional"`

	Resources *Resources `hcl:"resources,block"`
}

// Resources struct allows user to specify resource request and limits on the rook-ceph
// sub-components.
type Resources struct {
	MON               *util.ResourceRequirements `hcl:"mon,block"`
	MONRaw            string
	MGR               *util.ResourceRequirements `hcl:"mgr,block"`
	MGRRaw            string
	OSD               *util.ResourceRequirements `hcl:"osd,block"`
	OSDRaw            string
	MDS               *util.ResourceRequirements `hcl:"mds,block"`
	MDSRaw            string
	PrepareOSD        *util.ResourceRequirements `hcl:"prepareosd,block"`
	PrepareOSDRaw     string
	CrashCollector    *util.ResourceRequirements `hcl:"crashcollector,block"`
	CrashCollectorRaw string
	MGRSidecar        *util.ResourceRequirements `hcl:"mgr_sidecar,block"`
	MGRSidecarRaw     string
}

// StorageClass provides struct to enable it or make it default.
type StorageClass struct {
	Enable        bool   `hcl:"enable,optional"`
	Default       bool   `hcl:"default,optional"`
	ReclaimPolicy string `hcl:"reclaim_policy,optional"`
}

// NewConfig returns new Rook Ceph component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{
		Namespace:    "rook",
		MonitorCount: 1,
		StorageClass: &StorageClass{ReclaimPolicy: "Retain"},
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) addResourceRequirements() error {
	if c.Resources == nil {
		return nil
	}

	var err error

	if c.Resources.MONRaw, err = util.RenderResourceRequirements(c.Resources.MON); err != nil {
		return fmt.Errorf("rendering resources.mon: %w", err)
	}

	if c.Resources.MGRRaw, err = util.RenderResourceRequirements(c.Resources.MGR); err != nil {
		return fmt.Errorf("rendering resources.mgr: %w", err)
	}

	if c.Resources.OSDRaw, err = util.RenderResourceRequirements(c.Resources.OSD); err != nil {
		return fmt.Errorf("rendering resources.osd: %w", err)
	}

	if c.Resources.MDSRaw, err = util.RenderResourceRequirements(c.Resources.MDS); err != nil {
		return fmt.Errorf("rendering resources.mds: %w", err)
	}

	if c.Resources.PrepareOSDRaw, err = util.RenderResourceRequirements(c.Resources.PrepareOSD); err != nil {
		return fmt.Errorf("rendering resources.prepareosd: %w", err)
	}

	if c.Resources.CrashCollectorRaw, err = util.RenderResourceRequirements(c.Resources.CrashCollector); err != nil {
		return fmt.Errorf("rendering resources.crashcollector: %w", err)
	}

	if c.Resources.MGRSidecarRaw, err = util.RenderResourceRequirements(c.Resources.MGRSidecar); err != nil {
		return fmt.Errorf("rendering resources.mgr_sidecar: %w", err)
	}

	return nil
}

func (c *component) generateValues() (string, error) {
	var err error

	// Generate YAML for Ceph cluster.
	c.TolerationsRaw, err = util.RenderTolerations(c.Tolerations)
	if err != nil {
		return "", fmt.Errorf("rendering tolerations: %w", err)
	}

	c.NodeAffinityRaw, err = util.RenderNodeAffinity(c.NodeAffinity)
	if err != nil {
		return "", fmt.Errorf("rendering node affinity: %w", err)
	}

	if err := c.addResourceRequirements(); err != nil {
		return "", fmt.Errorf("rendering resources field: %w", err)
	}

	return template.Render(chartValuesTmpl, c)
}

// TODO: Convert to Helm chart.
func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	values, err := c.generateValues()
	if err != nil {
		return nil, fmt.Errorf("rendering values template: %w", err)
	}

	renderedFiles, err := util.RenderChart(helmChart, Name, c.Metadata().Namespace.Name, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
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
			TargetNamespace: c.Namespace,
			Values: &apiextensionsv1.JSON{
				Raw: values,
			},
		},
	}, nil
}
