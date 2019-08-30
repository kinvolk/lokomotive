package metallb

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
	if diagnostics.HasErrors() {
		t.Fatalf("Empty config should not return errors")
	}
}

func TestRenderManifest(t *testing.T) {
	configHCL := `
component "metallb" {
  speaker_toleration {
    key = "speaker_key1"
    operator = "Equal"
    value = "value1"
  }
  speaker_toleration {
    key = "speaker_key2"
    operator = "Equal"
    value = "value2"
  }

  controller_toleration {
    key = "controller_key1"
    operator = "Equal"
    value = "value1"
  }
  controller_toleration {
    key = "controller_key2"
    operator = "Equal"
    value = "value2"
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
