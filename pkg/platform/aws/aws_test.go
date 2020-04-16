//nolint:gomnd

package aws

import (
	"testing"

	configpkg "github.com/kinvolk/lokomotive/pkg/cluster/config"
)

func TestConfigNewconfig(t *testing.T) {
	c := newConfig()

	if c.Flatcar.OSName != "flatcar" {
		t.Fatalf("Expected default OSName to be `flatcar`, got: %s", c.Flatcar.OSName)
	}

	if c.Disk.Type != "gp2" {
		t.Fatalf("Expected default disk Type to be `gp2`, got: %s", c.Disk.Type)
	}

	if c.Disk.IOPS != 0 {
		t.Fatalf("Expected default disk IOPS to be `0`, got: %d", c.Disk.IOPS)
	}

	if c.Disk.Size != defaultDiskSize {
		t.Fatalf("Expected default disk Size to be `40`, got: %d", c.Disk.Size)
	}

	if c.Network.HostCIDR != "10.0.0.0/16" {
		t.Fatalf("Expected default network HostCIDR to be `10.0.0.0/16`, got: %s", c.Network.HostCIDR)
	}

	if c.Region != "eu-central-1" {
		t.Fatalf("Expected default Region to be `eu-central-1`, got: %s", c.Region)
	}
}

func TestConfigValidateSuccess(t *testing.T) {
	metadata := &configpkg.Metadata{
		ClusterName: "test-cluster",
		AssetDir:    "test-dir",
	}

	aws := &config{
		DNSZone:   "test-zone",
		DNSZoneID: "test-zone-id",
		Region:    "eu-central-1",
		Flatcar: &flatcar{
			OSName: "flatcar",
		},
		Network: &network{
			HostCIDR: "10.0.0.0/16",
		},
		Controller: &controller{
			Type: "t3_medium",
		},
		WorkerPools: []workerPool{
			{},
		},
	}

	aws.SetMetadata(metadata)

	diags := aws.Validate()
	if diags.HasErrors() {
		t.Fatalf("Expected no errors in validating configuration, got: %s", diags.Error())
	}
}

func TestConfigValidateFailInvalidCIDR(t *testing.T) {
	metadata := &configpkg.Metadata{
		ClusterName: "test-cluster",
		AssetDir:    "test-dir",
	}

	aws := &config{
		DNSZone:   "test-zone",
		DNSZoneID: "test-zone-id",
		Region:    "eu-central-1",
		Flatcar: &flatcar{
			OSName: "flatcar",
		},
		Network: &network{
			HostCIDR: "x.0.0.0/16",
		},
		Controller: &controller{
			Type: "t3_medium",
		},
		WorkerPools: []workerPool{
			{},
		},
	}

	aws.SetMetadata(metadata)

	diags := aws.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected no errors in validating configuration, got: %s", diags.Error())
	}
}

func TestConfigValidateFailLongClusterName(t *testing.T) {
	metadata := &configpkg.Metadata{
		ClusterName: "a-very-long-test-cluster-name-to-check-the-max-length",
		AssetDir:    "test-dir",
	}

	aws := &config{
		DNSZone:   "test-zone",
		DNSZoneID: "test-zone-id",
		Region:    "eu-central-1",
		Flatcar: &flatcar{
			OSName: "flatcar",
		},
		Network: &network{
			HostCIDR: "10.0.0.0/16",
		},
		Controller: &controller{
			Type: "t3_medium",
		},
		WorkerPools: []workerPool{
			{},
		},
	}

	aws.SetMetadata(metadata)

	diags := aws.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected name length error in validating configuration")
	}
}

func TestConfigValidateFailDuplicateWorkerPool(t *testing.T) {
	metadata := &configpkg.Metadata{
		ClusterName: "test-cluster",
		AssetDir:    "test-dir",
	}

	aws := &config{
		DNSZone:   "test-zone",
		DNSZoneID: "test-zone-id",
		Region:    "eu-central-1",
		Flatcar: &flatcar{
			OSName: "flatcar",
		},
		Network: &network{
			HostCIDR: "10.0.0.0/16",
		},
		Controller: &controller{
			Type: "t3_medium",
		},
		WorkerPools: []workerPool{
			{
				Name: "pool-name",
			},
			{
				Name: "pool-name",
			},
		},
	}

	aws.SetMetadata(metadata)

	diags := aws.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected unique worker pool name error in validating configuration")
	}
}

func TestRenderSuccess(t *testing.T) {
	metadata := &configpkg.Metadata{
		AssetDir:    "/path/to/assetdir",
		ClusterName: "test-cluster",
	}

	lokoCfg := &configpkg.LokomotiveConfig{
		Metadata:   metadata,
		Cluster:    configpkg.DefaultClusterConfig(),
		Flatcar:    configpkg.DefaultFlatcarConfig(),
		Network:    configpkg.DefaultNetworkConfig(),
		Controller: configpkg.DefaultControllerConfig(),
	}

	a := &config{
		DNSZone:         "test-zone",
		DNSZoneID:       "test-zone-id",
		ExposeNodePorts: true,
		Flatcar: &flatcar{
			OSName: "flatcar",
		},
		Network: &network{
			HostCIDR: "10.0.0.0/16",
		},
		Controller: &controller{
			Type: "t3_medium",
		},
		Disk: &disk{
			//nolint:gomnd
			Size: 80,
			Type: "eb2",
			//nolint:gomnd
			IOPS: 100,
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

	renderedTemplate, err := a.Render(lokoCfg)

	if err != nil {
		t.Fatalf("Expected render to succeed, got: %v", err)
	}

	if renderedTemplate == "" {
		t.Fatalf("Expected rendered string to be non-empty")
	}
}
