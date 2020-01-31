package contour

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokoctl/pkg/components/util"
)

func testRenderManifest(t *testing.T, configHCL string) {
	body, diagnostics := util.GetComponentBody(configHCL, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	component := newComponent()
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

func TestRenderManifestWithIngressHosts(t *testing.T) {
	configHCL := `
component "contour" {
  ingress_hosts = ["test.domain.com"]
}
`
	testRenderManifest(t, configHCL)
}

func TestRenderManifestWithServiceMonitor(t *testing.T) {
	configHCL := `
component "contour" {
  service_monitor = true
}
`
	testRenderManifest(t, configHCL)
}
