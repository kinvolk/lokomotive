//nolint:goconst
package config

import (
	"testing"
)

func TestConfigParseHCLFilesValidLokocfgFile(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}

metadata {
	asset_dir = "test/path"
	cluster_name = "test-cluster"
}
backend "local" {
 path = "test/backend/path"
}

flatcar {
	version = "latest"
	channel = "stable"
}

network {
 network_mtu = 1480
 pod_cidr = "10.2.0.0/16"
 service_cidr = "10.3.0.0/16"
}

controller {
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

	if config.Config == nil {
		t.Fatalf("ClusterConfig should not be nil after loading valid HCL configuration")
	}
}

func TestConfigParseHCLFilesMultipleLokocfgFiles(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}

controller {
  count = 1
  ssh_pubkeys = var.ssh_public_keys
  tags = {
    "key1"= "test1"
    "key2" = "test2"
  }
}
`
	hclconfig2 := `
metadata {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
}
`
	varconfig := `
ssh_public_keys="test-ssh-key"
`
	hclconfigMap := map[string][]byte{
		"test.lokocfg":  []byte(hclconfig),
		"test2.lokocfg": []byte(hclconfig2),
	}
	varconfigMap := map[string][]byte{
		"test.vars": []byte(varconfig),
	}

	config, diags := ParseHCLFiles(hclconfigMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Config should load from multiple config files: %s", diags)
	}

	if config.Config == nil {
		t.Fatalf("ClusterConfig should not be nil after loading valid HCL configuration files")
	}
}

func TestConfigParseHCLFilesInvalidLokocfgFile(t *testing.T) {
	hclconfig := `
variable "ssh_public_keys" {}
metadata {
  key = 
  key2 = "test2"
}
controller {
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
metadata {
  key = "test1" 
  key2 = "test2"
}
controller {
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
  key = "test"
  key2 = "test2"
}
controller {
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
		t.Fatalf("Unsupported block present: %q", diags)
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
controller {
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
metadata {
  key = "test"
  key2 = "test2"
}
controller {
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
metadata {
  key = "test"
  key2 = "test2"
}
controller {
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

	if config.Config == nil {
		t.Fatalf("Config should not be nil after loading invalid HCL configuration")
	}
}
