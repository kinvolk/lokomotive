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

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	internaltemplate "github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

const name = "rook-ceph"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Namespace      string              `hcl:"namespace,optional"`
	MonitorCount   int                 `hcl:"monitor_count,optional"`
	NodeAffinity   []util.NodeAffinity `hcl:"node_affinity,block"`
	MetadataDevice string              `hcl:"metadata_device,optional"`
	Tolerations    []util.Toleration   `hcl:"toleration,block"`
	TolerationsRaw string
	StorageClass   *StorageClass `hcl:"storage_class,block"`
}

// StorageClass provides struct to enable it or make it default.
type StorageClass struct {
	Enable  bool `hcl:"enable,optional"`
	Default bool `hcl:"default,optional"`
}

func newComponent() *component {
	return &component{
		Namespace:    "rook",
		MonitorCount: 1,
		StorageClass: &StorageClass{},
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	// Generate YAML for Ceph cluster.
	var err error
	c.TolerationsRaw, err = util.RenderTolerations(c.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("failed to render tolerations: %w", err)
	}

	ret := make(map[string]string)

	// Parse template with values
	for k, v := range template {
		rendered, err := internaltemplate.Render(v, c)
		if err != nil {
			return nil, fmt.Errorf("template rendering failed for %q: %w", k, err)
		}

		ret[k] = rendered
	}

	return ret, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: c.Namespace,
	}
}
