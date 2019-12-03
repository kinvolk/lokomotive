package flatcarlinuxupdateoperator

import (
	"testing"

	"github.com/hashicorp/hcl2/hcl"

	"github.com/kinvolk/lokoctl/pkg/components/util"
)

func TestRenderManifest(t *testing.T) {
	configHCL := `
component "flatcar-linux-update-operator" {}
	`

	component := &component{}

	body, diagnostics := util.GetComponentBody(configHCL, componentName)
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
