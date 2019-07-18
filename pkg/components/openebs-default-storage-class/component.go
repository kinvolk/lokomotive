package openebsoperator

import (
	"fmt"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/assets"
	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
	"github.com/kinvolk/lokoctl/pkg/util/walkers"
)

const name = "openebs-default-storage-class"

func init() {
	components.Register(name, &component{})
}

type component struct{}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *component) RenderManifests() (map[string]string, error) {
	ret := make(map[string]string)
	walk := walkers.DumpingWalker(ret, ".yaml")
	if err := assets.Assets.WalkFiles(fmt.Sprintf("/components/%s/manifests", name), walk); err != nil {
		return nil, errors.Wrap(err, "failed to walk assets")
	}

	return ret, nil
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
