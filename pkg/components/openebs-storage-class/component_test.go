package openebsstorageclass

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
		t.Fatal("Empty config should not return errors")
	}
}

func TestDefaultValues(t *testing.T) {
	c := newComponent()

	if len(c.Storageclasses) != 1 {
		t.Fatal("Default should contain only 1 storage class")
	}
	if c.Storageclasses[0].ReplicaCount != 3 {
		t.Fatal("Default value of replica count should be 3")
	}
	if !c.Storageclasses[0].Default {
		t.Fatal("Default value should be true")
	}
	if len(c.Storageclasses[0].Disks) != 0 {
		t.Fatal("Default list of disks should be empty")
	}
}

func TestUserInputValues(t *testing.T) {

	storageClasses := `
	component "openebs-storage-class" {
		storage-class "replica1-no-disk-selected" {
			replica_count = 1
		}
		storage-class "replica1" {
			disks = ["disk1"]
			replica_count = 1
		}
		storage-class "replica3" {
			replica_count = 3
			default = true
			disks = ["disk2","disk3","disk4"]
		}
	}
	`
	testRenderManifest(t, storageClasses)
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
