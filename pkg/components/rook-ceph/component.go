package rookceph

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
	"github.com/pkg/errors"
)

const name = "rook-ceph"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Namespace      string              `hcl:"namespace,optional"`
	MonitorCount   int                 `hcl:"monitor_count,optional"`
	NodeSelectors  []util.NodeSelector `hcl:"node_selector,block"`
	MetadataDevice string              `hcl:"metadata_device,optional"`
	Tolerations    []util.Toleration   `hcl:"toleration,block"`
	TolerationsRaw string
}

func newComponent() *component {
	return &component{
		Namespace:    "rook",
		MonitorCount: 1,
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
		return nil, errors.Wrap(err, "failed to render tolerations")
	}

	cephClusterStr, err := util.RenderTemplate(cephCluster, c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render template")
	}

	return map[string]string{
		"ceph-cluster.yaml": cephClusterStr,
	}, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: c.Namespace,
	}
}
