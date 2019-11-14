package openebsstorageclass

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

const (
	name     = "openebs-storage-class"
	poolName = "openebs-storage-pool"
)

func init() {
	components.Register(name, newComponent())
}

type Storageclass struct {
	Name         string   `hcl:"name,label"`
	ReplicaCount int      `hcl:"replica_count,optional"`
	Default      bool     `hcl:"default,optional"`
	Disks        []string `hcl:"disks,optional"`
}
type component struct {
	Storageclasses []Storageclass `hcl:"storage-class,block"`
}

func defaultStorageClass() Storageclass {
	return Storageclass{
		// Name of the storage class
		Name: "openebs-cstor-disk-replica-3",
		// Default replica count value is set to 3
		ReplicaCount: 3,
		// Make the storage class as default
		Default: true,
		// Default disks selection is empty
		Disks: make([]string, 0),
	}
}

func newComponent() *component {
	sc := defaultStorageClass()
	return &component{
		Storageclasses: []Storageclass{sc},
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	if diagnostics := gohcl.DecodeBody(*configBody, evalContext, c); diagnostics != nil {
		return diagnostics
	}
	// if empty config body is provided, default component storage details are still preserved.
	if len(c.Storageclasses) == 0 {
		c.Storageclasses = append(c.Storageclasses, defaultStorageClass())
	}

	if err := c.validateConfig(); err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Validation of the config failed",
				Detail:   fmt.Sprintf("validation failed: %v", err),
			},
		}
	}

	return nil
}

func (c *component) validateConfig() error {
	maxDefaultStorageClass := 0
	for _, sc := range c.Storageclasses {
		if sc.Default == true {
			maxDefaultStorageClass++
		}
		if maxDefaultStorageClass > 1 {
			return errors.New("cannot have more than one default storage class")
		}
	}

	return nil
}

func (c *component) RenderManifests() (map[string]string, error) {

	scTmpl, err := template.New(name).Parse(storageClassTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}

	spTmpl, err := template.New(poolName).Parse(storagePoolTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}

	var manifestsMap = make(map[string]string)

	for _, sc := range c.Storageclasses {
		var scBuffer bytes.Buffer
		var spBuffer bytes.Buffer

		if err := scTmpl.Execute(&scBuffer, sc); err != nil {
			return nil, errors.Wrap(err, "execute template failed")
		}

		filename := fmt.Sprintf("%s-%s.yml", name, sc.Name)
		manifestsMap[filename] = scBuffer.String()

		if err := spTmpl.Execute(&spBuffer, sc); err != nil {
			return nil, errors.Wrap(err, "execute template failed")
		}

		filename = fmt.Sprintf("%s-%s.yml", poolName, sc.Name)
		manifestsMap[filename] = spBuffer.String()
	}

	return manifestsMap, nil
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
