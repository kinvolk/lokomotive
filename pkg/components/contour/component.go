package contour

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/assets"
	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/util/walkers"
)

const name = "contour"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	ServiceMonitor bool `hcl:"service_monitor,optional"`

	// TODO: add num of replicas when using "deployment"
}

func newComponent() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}
	if err := gohcl.DecodeBody(*configBody, evalContext, c); err != nil {
		return err
	}
	return nil
}

func (c *component) RenderManifests() (map[string]string, error) {
	ret := make(map[string]string)

	walk := walkers.DumpingWalker(ret, ".yaml")
	if err := assets.Assets.WalkFiles(fmt.Sprintf("/components/%s/%s", name, name), walk); err != nil {
		return nil, errors.Wrap(err, "failed to walk assets")
	}

	// Create service and service monitor for Prometheus to scrape metrics
	if c.ServiceMonitor {
		if err := assets.Assets.WalkFiles(fmt.Sprintf("/components/%s/manifests-metrics", name), walk); err != nil {
			return nil, errors.Wrap(err, "failed to walk assets")
		}
	}

	return ret, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: "projectcontour",
	}
}
