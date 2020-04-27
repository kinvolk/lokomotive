package config

import (
	"testing"

	// Register platforms by adding an anonymous import.
	_ "github.com/kinvolk/lokomotive/pkg/platform/aws"
	_ "github.com/kinvolk/lokomotive/pkg/platform/baremetal"
	_ "github.com/kinvolk/lokomotive/pkg/platform/packet"

	// Register backends by adding an anonymous import.
	_ "github.com/kinvolk/lokomotive/pkg/backend/local"
	_ "github.com/kinvolk/lokomotive/pkg/backend/s3"
)

func TestHCLLoaderParsePacketValidLokomotiveConfig(t *testing.T) {
	cfg := `
variable "ssh_public_keys" {}

backend "local" {
 path = "test/backend/path"
}
cluster "packet" {
	controller_count = 3
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  enable_aggregation = false
  cluster_domain_suffix = "test.local"
  certs_validity_period_hours = 100
  os_version = "latest"
  os_channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  enable_reporting = true
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
  project_id = "test-project-id"
  facility = "ams1"
  management_cidrs = ["0.0.0.0/0"]
  node_private_cidr = "10.0.0.0/8"
  dns {
    zone = "test.zone"
    provider {
      route53 {
        zone_id = "TESTZONE"
      }
    }
  }
}
`

	varconfig := `
ssh_public_keys=["test-ssh-key"]
`

	cfgMap := map[string][]byte{
		"cluster.lokocfg": []byte(cfg),
	}

	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	lokocfg, diags := ParseToLokomotiveConfig(config)
	if diags.HasErrors() {
		for _, diag := range diags {
			t.Error(diag)
		}

		t.Fatal("Error parsing cluster config to LokomotiveConfig:")
	}

	if lokocfg.Platform == nil {
		t.Fatalf("expected 'platform' to not be nil")
	}

	if lokocfg.Backend == nil {
		t.Fatalf("expected 'backend' to not be nil")
	}
}

func TestHCLLoaderPacketRequiredFieldMissing(t *testing.T) {
	cfg := `
variable "ssh_public_keys" {}

cluster "packet" {
  enable_aggregation = false
  cluster_domain_suffix = "test.local"
  certs_validity_period_hours = 100
  os_version = "latest"
  os_channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  enable_reporting = true
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
  dns {
    zone = "test.zone"
    provider {
      route53 {
        zone_id = "TESTZONE"
      }
    }
  }
}
`

	varconfig := `
ssh_public_keys=["test-ssh-key"]
`

	cfgMap := map[string][]byte{
		"cluster.lokocfg": []byte(cfg),
	}

	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	platform, diags := loadPlatformConfiguration(config)
	if !diags.HasErrors() {
		t.Fatal("Expected missing required argument errors")
	}

	if platform != nil {
		t.Fatalf("expected 'platform' to be nil, got: %s", platform)
	}
}
