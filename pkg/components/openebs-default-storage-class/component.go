package openebsoperator

import (
	"path"

	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/packr/v2/file"
	"github.com/hashicorp/hcl2/hcl"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
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
	box := packr.New(name, "../../../assets/components/openebs-default-storage-class/manifests/")

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
