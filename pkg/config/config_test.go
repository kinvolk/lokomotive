//nolint:goconst
package config

import (
	"testing"
)

func TestConfigParseHCLFilesValidLokocfgFile(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}

backend "local" {
 path = "test/backend/path"
}
cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version = "latest"
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
`

	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}
	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid HCL configuration should not produce any errors: %q", diags)
	}

	if config.ClusterConfig == nil {
		t.Fatalf("Config should not be nil after loading valid HCL configuration")
	}
}

func TestConfigParseHCLFilesValidLokocfgFilePathExpand(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}

backend "local" {
 path = "test/backend/path"
}
cluster "packet" {
  asset_dir = pathexpand("~")
  cluster_name = "test-cluster"
  version = "latest"
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
`

	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}
	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if diags.HasErrors() {
		for _, diag := range diags {
			t.Error(diag)
		}

		t.Fatal("Loading valid HCL configuration should not produce any errors:")
	}

	if config.ClusterConfig == nil {
		t.Fatalf("Config should not be nil after loading valid HCL configuration")
	}
}

func TestConfigParseHCLFilesValidLokocfgFileWithFilepath(t *testing.T) {
	hclconfig := `
backend "local" {
 path = "test/backend/path"
}
cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version = "latest"
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = file("./config_test.go")
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
`

	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}
	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid HCL configuration should not produce any errors: %q", diags)
	}

	if config.ClusterConfig == nil {
		t.Fatalf("Config should not be nil after loading valid HCL configuration")
	}
}
func TestConfigParseHCLFilesMultipleLokocfgFiles(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}

cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version = "latest"
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	hclconfig2 := `
backend "local" {
 path = "test/backend/path"
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
`
	hclconfigMap := map[string][]byte{
		"cluster.lokocfg": []byte(hclconfig),
		"backend.lokocfg": []byte(hclconfig2),
	}
	varconfigMap := map[string][]byte{
		"cluster.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Config should load from multiple config files: %s", diags)
	}

	if config.ClusterConfig == nil {
		t.Fatalf("Config should not be nil after loading valid HCL configuration files")
	}
}

func TestConfigParseHCLFilesInvalidLokocfgFile(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}
cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version =
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
`

	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}
	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if !diags.HasErrors() {
		t.Fatalf("Invalid config file: %q", diags)
	}

	if config != nil {
		t.Fatalf("Config should be nil after loading invalid HCL configuration")
	}
}

func TestConfigParseHCLFilesInvalidVarsFile(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}
cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version =
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
novalue=
`

	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}
	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if !diags.HasErrors() {
		t.Fatalf("Invalid vars file: %q", diags)
	}

	if config != nil {
		t.Fatalf("Config should be nil after loading invalid HCL configuration")
	}
}

func TestConfigParseHCLFilesInvalidConfigBlock(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}

test {
  key = "value"
}

cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version =
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
	`

	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}
	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if !diags.HasErrors() {
		t.Fatalf("Invalid block present: %q", diags)
	}

	if config != nil {
		t.Fatalf("Config should be nil after loading invalid HCL configuration")
	}
}

func TestConfigParseHCLFilesEmptyLokocfgFile(t *testing.T) {
	hclconfig := ""
	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}

	varconfigMap := map[string][]byte{}

	_, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Empty config should not have any errors while parsing: %q", diags)
	}
}

func TestConfigParseHCLFilesEmptyVarsFile(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}
cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version = "current"
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	varconfig := ""

	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}

	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	_, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Empty vars file should not have any errors while parsing: %q", diags)
	}
}

func TestConfigParseHCLFilesDuplicateVarsInFile(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}
cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version =
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
`
	varconfig2 := `
ssh_public_keys="test-ssh-key"
`
	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}
	varconfigMap := map[string][]byte{
		"test.vars":  []byte(varconfig),
		"test2.vars": []byte(varconfig2),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if !diags.HasErrors() {
		t.Fatalf("Duplicate variable found: %q", diags)
	}

	if config != nil {
		t.Fatalf("Config should be nil after loading invalid HCL configuration")
	}
}

func TestConfigParseHCLFilesDefaultVarsValue(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {
	default = "default-ssh"
}
cluster "packet" {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  version = "current"
  channel = "stable"
  network_mtu = 1480
  pod_cidr = "10.2.0.0/16"
  service_cidr = "10.3.0.0/16"
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	hclconfigMap := map[string][]byte{
		"test.lokocfg": []byte(hclconfig),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Should use default value of variable and not throw error: %q", diags)
	}

	if config.ClusterConfig == nil {
		t.Fatalf("Config should not be nil after loading invalid HCL configuration")
	}
}
