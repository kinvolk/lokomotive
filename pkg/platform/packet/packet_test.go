//nolint:gomnd

package packet

import (
	"testing"

	configpkg "github.com/kinvolk/lokomotive/pkg/cluster/config"
	"github.com/kinvolk/lokomotive/pkg/dns"
)

func TestConfigNewconfig(t *testing.T) {
	c := newConfig()
	if c.Flatcar.Arch != "amd64" {
		t.Fatalf("Expected default Arch to be `amd64`, got: %s", c.Flatcar.Arch)
	}

	if c.Controller.Type != "baremetal_0" {
		t.Fatalf("Expected default controller Type to be `baremetal_0`, got: %s", c.Controller.Type)
	}
}

func TestConfigValidateSuccess(t *testing.T) {
	p := &config{
		AuthToken: "test-token",
		ProjectID: "test-project-id",
		Facility:  "ams1",
		Flatcar: &flatcar{
			Arch: "amd64",
		},
		Network: &network{
			NodePrivateCIDR: "10.0.0.0/16",
			ManagementCIDRs: []string{"0.0.0.0/16"},
		},
		Controller: &controller{
			Type: "baremetal_0",
		},
		WorkerPools: []workerPool{
			{},
		},
	}

	diags := p.Validate()
	if diags.HasErrors() {
		t.Fatalf("Expected no errors in validating configuration, got: %s", diags.Error())
	}
}

func TestConfigValidateFail(t *testing.T) {
	p := &config{
		AuthToken: "test-token",
		ProjectID: "test-project-id",
		Facility:  "sjc1",
		Flatcar: &flatcar{
			Arch: "amd64",
		},
		Network: &network{
			NodePrivateCIDR: "x.0.0.0/16",
			ManagementCIDRs: []string{"0.0.0.0/16"},
		},
		Controller: &controller{
			Type: "baremetal_0",
		},
		WorkerPools: []workerPool{
			{},
		},
	}

	diags := p.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected no errors in validating configuration, got: %s", diags.Error())
	}
}

func TestRenderSuccess(t *testing.T) {
	metadata := &configpkg.Metadata{
		AssetDir:    "/path/to/assetdir",
		ClusterName: "test-cluster",
	}

	lokoCfg := &configpkg.LokomotiveConfig{
		Metadata: metadata,
		Cluster:  configpkg.DefaultClusterConfig(),
		Flatcar:  configpkg.DefaultFlatcarConfig(),
		Network:  configpkg.DefaultNetworkConfig(),
		Controller: &configpkg.ControllerConfig{
			//nolint:gomnd
			Count:      3,
			SSHPubKeys: []string{"test-ssh-key"},
		},
	}

	p := &config{
		AuthToken: "test-token",
		ProjectID: "test-project-id",
		Facility:  "sjc1",
		Flatcar: &flatcar{
			Arch:          "arm64",
			IPXEScriptURL: "https://this.is.a.test.url",
		},
		DNS: dns.Config{
			Zone: "test-zone",
		},
		Network: &network{
			NodePrivateCIDR: "10.0.0.0/16",
			ManagementCIDRs: []string{"0.0.0.0/16"},
		},
		Controller: &controller{
			Type: "baremetal_0",
			Tags: map[string]string{
				"key1": "value1",
			},
		},
		WorkerPools: []workerPool{
			{
				Name: "pool1",
				//nolint:gomnd
				Count: 2,
			},
		},
		Metadata: &configpkg.Metadata{
			AssetDir:    "/path/to/assetdir",
			ClusterName: "test-cluster",
		},
	}

	renderedTemplate, err := p.Render(lokoCfg)
	if err != nil {
		t.Fatalf("Expected render to succeed, got: %v", err)
	}

	if renderedTemplate == "" {
		t.Fatalf("Expected rendered string to be non-empty")
	}
}
