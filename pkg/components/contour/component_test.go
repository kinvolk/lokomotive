package contour

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
		t.Fatalf("Empty config should return errors")
	}
}

func testRenderManifest(t *testing.T, configHCL string) {
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

func TestRenderManifest_WithInstallModeDeployment(t *testing.T) {
	configHCL := `
component "contour" {
  install_mode = "deployment"
}
`
	testRenderManifest(t, configHCL)
}

func TestRenderManifest_WithInstallModeDaemonSet(t *testing.T) {
	configHCL := `
component "contour" {
  install_mode = "daemonset"
}
`
	testRenderManifest(t, configHCL)
}

func TestRenderManifestWithInstallModeDeploymentAndServiceMonitor(t *testing.T) {
	configHCL := `
component "contour" {
  install_mode = "deployment"
  service_monitor = true
}
`
	testRenderManifest(t, configHCL)
}

func TestRenderManifestWithInstallModeDaemonSetAndServiceMonitor(t *testing.T) {
	configHCL := `
component "contour" {
  install_mode = "daemonset"
  service_monitor = true
}
`
	testRenderManifest(t, configHCL)
}
