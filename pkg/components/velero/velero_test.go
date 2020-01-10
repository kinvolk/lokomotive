package velero

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokoctl/pkg/components/util"
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

	component := newComponent()

	body, diagnostics := util.GetComponentBody(configHCL, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
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
