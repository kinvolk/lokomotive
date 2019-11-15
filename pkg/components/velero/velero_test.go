package velero

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"

	"github.com/kinvolk/lokoctl/pkg/config"
)

func TestEmptyConfig(t *testing.T) {
	c := newComponent()
	emptyConfig := hcl.EmptyBody()
	evalContext := hcl.EvalContext{}
	diagnostics := c.LoadConfig(&emptyConfig, &evalContext)
	if !diagnostics.HasErrors() {
		t.Errorf("Empty config should return error")
	}
}

func TestRenderManifestAzure(t *testing.T) {
	configHCL := `
component "velero" {
  azure {
    subscription_id  = "foo"
    tenant_id        = "foo"
    client_id        = "foo"
    client_secret    = "foo"
    resource_group   = "foo"

    backup_storage_location {
      resource_group  = "foo"
      storage_account = "foo"
      bucket          = "foo"
    }
  }
}
`

	hclParser := hclparse.NewParser()

	file, diags := hclParser.ParseHCL([]byte(configHCL), fmt.Sprintf("%s.lokocfg", name))
	if diags.HasErrors() {
		t.Fatalf("Parsing config should succeed, got: %s", diags)
	}

	configBody := hcl.MergeFiles([]*hcl.File{file})

	var rootConfig config.RootConfig

	diagnostics := gohcl.DecodeBody(configBody, nil, &rootConfig)
	if diags.HasErrors() {
		t.Fatalf("Valid root config should not return error, got: %s", diagnostics)
	}

	c := &config.Config{
		RootConfig: &rootConfig,
	}

	component := newComponent()
	diagnostics = component.LoadConfig(c.LoadComponentConfigBody(name), &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}

	m, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}
	if len(m) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}
