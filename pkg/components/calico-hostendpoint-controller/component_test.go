package calicohostendpointcontroller

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokoctl/pkg/components/util"
)

func TestRenderManifest(t *testing.T) {
	configHCL := `
component "calico-hostendpoint-controller" {}
	`

	component := &component{}

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

func TestEmptyConfig(t *testing.T) {
	component := &component{}
	emptyConfig := hcl.EmptyBody()
	evalContext := hcl.EvalContext{}
	diagnostics := component.LoadConfig(&emptyConfig, &evalContext)
	if diagnostics.HasErrors() {
		t.Fatalf("Empty config should not return errors")
	}

	m, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}
	if len(m) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}

func TestRenderManifestBadConfig(t *testing.T) {
	configHCL := `
component "calico-hostendpoint-controller" {
  foo = "bar"
}
  `

	component := &component{}

	body, diagnostics := util.GetComponentBody(configHCL, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
	if !diagnostics.HasErrors() {
		t.Fatalf("Invalid config should return error")
	}
}
