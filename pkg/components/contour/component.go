package contour

import (
	"path"

	packr "github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/packr/v2/file"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
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

	// XXX: To use dynamic resolution with packr2 path var, we should use
	// something like this:
	// https://github.com/gobuffalo/packr/tree/master/v2#dynamic-box-paths
	//
	// But that fails using the manifest path as a string variable, for some
	// reason (need to create a mini reproduction case and report the bug).
	// However, I found a work around: we can just declare the box pointer
	// and assign the pointer to a new box on each path (using a different
	// name for the box). That avoids all issues with "packr2 build" and the
	// right assets are used on each case.
	var box *packr.Box

	if c.InstallMode == "deployment" {
		box = packr.New("contour-deployment", "../../../assets/components/contour/manifests-deployment/")
	} else if c.InstallMode == "daemonset" {
		box = packr.New("contour-daemonset", "../../../assets/components/contour/manifests-daemonset/")
	} else {
		// This should not be possible, it was validated during load
		panic("This is a bug: install_mode was a valid value and it is not a valid value now.")
	}

	box.Walk(func(f string, content file.File) error {
		if path.Ext(f) != ".yaml" {
			return nil
		}

		ret[f] = content.String()
		return nil
	})
	return ret, nil
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
