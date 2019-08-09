package prometheus

import (
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
)

func TestEmptyConfig(t *testing.T) {
	c := newComponent()
	emptyConfig := hcl.EmptyBody()
	evalContext := hcl.EvalContext{}
	diagnostics := c.LoadConfig(&emptyConfig, &evalContext)
	if !diagnostics.HasErrors() {
		t.Fatalf("Empty config should return error")
	}
}

func TestRenderManifest(t *testing.T) {
	config := `
component "prometheus-operator" {
  grafana_admin_password       = "admin"
  prometheus_metrics_retention = "90d"
  namespace                    = "monitoring"
}
`
	hclParser := hclparse.NewParser()

	file, diags := hclParser.ParseHCL([]byte(config), "prometheus-operator.lokocfg")
	if diags.HasErrors() {
		t.Fatalf("Parsing config should succeed, got: %s", diags)
	}

	configBody := hcl.MergeFiles([]*hcl.File{file})

	c := newComponent()
	diagnostics := c.LoadConfig(&configBody, &hcl.EvalContext{})
	if !diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diags)
	}

	m, err := c.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}
	if len(m) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}
