package contour

import (
	"fmt"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/assets"
	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
	"github.com/kinvolk/lokoctl/pkg/util/walkers"
)

const name = "contour"

func init() {
	components.Register(name, &component{})
}

type component struct {
	InstallMode string `hcl:"install_mode,attr"`

	// TODO: add num of replicas when using install_mode "deployment"
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
	if c.InstallMode != "deployment" && c.InstallMode != "daemonset" {
		err := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "install_mode must be either 'deployment' or 'daemonset'",
			Detail:   "Make sure to set install_mode to either 'deployment' or 'daemonset' in lowercase",
		}
		return hcl.Diagnostics{err}
	}
	return nil
}

func (c *component) RenderManifests() (map[string]string, error) {
	ret := make(map[string]string)
	switch c.InstallMode {
	case "deployment", "daemonset":
		break
	default:
		// This should not be possible, it was validated during load
		panic("This is a bug: install_mode was a valid value and it is not a valid value now.")
	}

	walk := walkers.DumpingWalker(ret, ".yaml")
	if err := assets.Assets.WalkFiles(fmt.Sprintf("/components/%s/manifests-%s", name, c.InstallMode), walk); err != nil {
		return nil, errors.Wrap(err, "failed to walk assets")
	}
	return ret, nil
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
