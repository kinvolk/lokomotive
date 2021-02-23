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

package awsebscsidriver

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents AWS EBS CSI driver component name as it should be referenced in function calls
	// and in configuration.
	Name = "aws-ebs-csi-driver"

	chartValuesTmpl = `
enableDefaultStorageClass: {{ .EnableDefaultStorageClass }}
# Enable volume scheduling for dynamic volume provisioning.
enableVolumeScheduling: {{ .EnableVolumeScheduling }}
# Enable volume resizing.
enableVolumeResizing: {{ .EnableVolumeResizing }}
# Enable volume snapshot.
enableVolumeSnapshot: {{ .EnableVolumeSnapshot }}

storageClasses:
- name: ebs-sc
  {{ if .EnableDefaultStorageClass }}
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
  {{ end }}
  volumeBindingMode: WaitForFirstConsumer
  reclaimPolicy: Retain

{{- if .Tolerations }}
tolerateAllTaints: false
tolerations: {{ .TolerationsRaw }}
node:
  tolerateAllTaints: false
  tolerations: {{ .TolerationsRaw }}
{{- end }}

{{- if .NodeAffinity }}
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions: {{ .NodeAffinityRaw }}
{{- end}}
`
)

type component struct {
	EnableDefaultStorageClass bool                `hcl:"enable_default_storage_class,optional"`
	EnableVolumeScheduling    bool                `hcl:"enable_volume_scheduling,optional"`
	EnableVolumeResizing      bool                `hcl:"enable_volume_resizing,optional"`
	EnableVolumeSnapshot      bool                `hcl:"enable_volume_snapshot,optional"`
	Tolerations               []util.Toleration   `hcl:"tolerations,block"`
	NodeAffinity              []util.NodeAffinity `hcl:"node_affinity,block"`

	TolerationsRaw  string
	NodeAffinityRaw string
}

// NewConfig returns new AWS EBS CSI driver component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{
		EnableDefaultStorageClass: true,
		EnableVolumeScheduling:    true,
		EnableVolumeResizing:      true,
		EnableVolumeSnapshot:      true,
	}
}

// LoadConfig loads the component config.
func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	return gohcl.DecodeBody(*configBody, evalContext, c)
}

// RenderManifests renders the Helm chart templates with values provided.
func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	c.TolerationsRaw, err = util.RenderTolerations(c.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering tolerations failed: %w", err)
	}

	c.NodeAffinityRaw, err = util.RenderNodeAffinity(c.NodeAffinity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node affinity: %w", err)
	}

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering chart values template failed: %w", err)
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
			Name: "kube-system",
		},
	}
}
