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

func TestHCLLoaderParseValidClusterConfig(t *testing.T) {
	cfg := `
cluster {
  enable_aggregation = false
	cluster_domain_suffix = "test.local"
	certs_validity_period_hours = 100
}
`
	cfgMap := map[string][]byte{
		"cluster.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid cluster configuration should not produce any errors: %s", diags)
	}

	cluster, diags := loadClusterConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing cluster config to LokomotiveConfig: %s", diags)
	}

	if cluster.EnableAggregation {
		t.Fatalf("enable_aggregation should be false', got: %t", cluster.EnableAggregation)
	}

	if cluster.ClusterDomainSuffix != "test.local" {
		t.Fatalf("cluster_domain_suffix should be equal to 'test.local', got: %s", cluster.ClusterDomainSuffix)
	}
	//nolint:gomnd
	if cluster.CertsValidityPeriodHours != 100 {
		t.Fatalf("certs_validity_period_hours should be equal to 100, got: %d", cluster.CertsValidityPeriodHours)
	}
}

func TestHCLLoaderParseClusterConfigUnsupportedArgument(t *testing.T) {
	cfg := `
cluster {
	enable_aggregation = false
	cluster_domain_suffix = "test.local"
	certs_validity_period_hours = 100
  test_field = "should not exist"
}
`
	cfgMap := map[string][]byte{
		"metadata.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid cluster configuration should not produce any errors: %s", diags)
	}

	cluster, diags := loadClusterConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("expected unsupported argument, got: %s", diags)
	}

	if cluster != nil {
		t.Fatalf("expected cluster to be nil, got: %v", cluster)
	}
}

func TestHCLLoaderParseValidMetadataConfig(t *testing.T) {
	cfg := `
metadata {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
}
`
	cfgMap := map[string][]byte{
		"metadata.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid metadata configuration should not produce any errors: %s", diags)
	}

	metadata, diags := loadMetadataConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing metadata config to LokomotiveConfig: %s", diags)
	}

	if metadata.AssetDir != "test/path" {
		t.Fatalf("asset_dir should be equal to 'test/path', got: %s", metadata.AssetDir)
	}

	if metadata.ClusterName != "test-cluster" {
		t.Fatalf("cluster_name should be equal to 'test-cluster', got: %s", metadata.ClusterName)
	}
}

func TestHCLLoaderParseMetadataConfigRequiredArgumentMissing(t *testing.T) {
	cfg := `
metadata {
  asset_dir = "test/path"
}
`
	cfgMap := map[string][]byte{
		"metadata.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid metadata configuration should not produce any errors: %s", diags)
	}

	metadata, diags := loadMetadataConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("expected missing required argument, got: %s", diags)
	}

	if metadata != nil {
		t.Fatalf("expected metadata to be nil, got: %v", metadata)
	}
}

func TestHCLLoaderParseMetadataConfigUnsupportedArgument(t *testing.T) {
	cfg := `
metadata {
  asset_dir = "test/path"
  cluster_name = "test-cluster"
  test_field = "should not exist"
}
`
	cfgMap := map[string][]byte{
		"metadata.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid metadata configuration should not produce any errors: %s", diags)
	}

	metadata, diags := loadMetadataConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("expected unsupported argument, got: %s", diags)
	}

	if metadata != nil {
		t.Fatalf("expected metadata to be nil, got: %v", metadata)
	}
}

func TestHCLLoaderParseValidControllerConfig(t *testing.T) {
	cfg := `
controller {
  count = 3
	ssh_pubkeys = ["test-ssh-key"]
}
`
	cfgMap := map[string][]byte{
		"controller.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid controller configuration should not produce any errors: %s", diags)
	}

	controller, diags := loadControllerConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing metadata config to LokomotiveConfig: %s", diags)
	}

	//nolint:gomnd
	if controller.Count != 3 {
		t.Fatalf("count should be equal to 3, got: %d", controller.Count)
	}

	//nolint:gomnd
	//if len(controller.Tags) != 1 {
	//	t.Fatalf("number of 'tags' should be equal to 1, got: %d", len(controller.Tags))
	//}

	//if controller.Tags["key1"] != "test1" {
	//	t.Fatalf("value of the tag key 'key1' should be equal to 'test1', got: %s", controller.Tags["key1"])
	//}

	//nolint:gomnd
	if len(controller.SSHPubKeys) != 1 {
		t.Fatalf("number of 'ssh_pubkeys' should be equal to 1, got: %d", len(controller.SSHPubKeys))
	}

	if controller.SSHPubKeys[0] != "test-ssh-key" {
		t.Fatalf("ssh_pub_keys should be equal to 'test-ssh-key', got: %s", controller.SSHPubKeys[0])
	}
}

func TestHCLLoaderParseControllerConfigRequiredArgumentMissing(t *testing.T) {
	cfg := `
controller {
	tags = {
		"key1" = "test1"
	}
}
`
	cfgMap := map[string][]byte{
		"controller.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	controller, diags := loadControllerConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("expected required argument missing, got: %s", diags)
	}

	if controller != nil {
		t.Fatalf("expected controller to be nil, got: %v", controller)
	}
}

func TestHCLLoaderParseControllerConfigUnsupportedArgument(t *testing.T) {
	cfg := `
controller {
  count = 1
  ssh_pubkeys = ["test-ssh-key"]
	tags = {
		"key1" = "test1"
	}
  test_field = "test-field"
}
`
	cfgMap := map[string][]byte{
		"controller.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	controller, diags := loadControllerConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("expected unsupported argument, got: %s", diags)
	}

	if controller != nil {
		t.Fatalf("expected controller to be nil, got: %v", controller)
	}
}

func TestHCLLoaderParseValidNetworkConfig(t *testing.T) {
	cfg := `
network {
  network_mtu = 1000
  pod_cidr = "10.10.0.0/16"
	service_cidr = "10.11.0.0/16"
  enable_reporting = true
}
`
	cfgMap := map[string][]byte{
		"network.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid network configuration should not produce any errors: %s", diags)
	}

	network, diags := loadNetworkConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing network config to NetworkConfig: %s", diags)
	}

	//nolint:gomnd
	if network.NetworkMTU != 1000 {
		t.Fatalf("expected 'network_mtu' to be 1000, got: %d", network.NetworkMTU)
	}

	if network.PodCIDR != "10.10.0.0/16" {
		t.Fatalf("expected 'pod_cidr' to be '10.10.0.0/16', got: %s", network.PodCIDR)
	}

	if network.ServiceCIDR != "10.11.0.0/16" {
		t.Fatalf("expected 'service_cidr' to be '10.11.0.0/16', got: %s", network.ServiceCIDR)
	}

	if !network.EnableReporting {
		t.Fatalf("expected 'enable_reporting' to be true, got: %t", network.EnableReporting)
	}
}

func TestHCLLoaderParseNetworkConfigDefaultConfig(t *testing.T) {
	cfg := `
network {
}
`
	cfgMap := map[string][]byte{
		"network.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid network configuration should not produce any errors: %s", diags)
	}

	network, diags := loadNetworkConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing network config to NetworkConfig: %s", diags)
	}

	//nolint:gomnd
	if network.NetworkMTU != 1480 {
		t.Fatalf("expected 'network_mtu' to be 1480, got: %d", network.NetworkMTU)
	}

	if network.PodCIDR != "10.2.0.0/16" {
		t.Fatalf("expected 'pod_cidr' to be '10.2.0.0/16', got: %s", network.PodCIDR)
	}

	if network.ServiceCIDR != "10.3.0.0/16" {
		t.Fatalf("expected 'service_cidr' to be '10.3.0.0/16', got: %s", network.ServiceCIDR)
	}

	if network.EnableReporting {
		t.Fatalf("expected 'enable_reporting' to be false, got: %t", network.EnableReporting)
	}
}

func TestHCLLoaderParseNetworkConfigUnexpectedArgument(t *testing.T) {
	cfg := `
network {
test_field = "test-field"
}
`
	cfgMap := map[string][]byte{
		"network.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	network, diags := loadNetworkConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("Error parsing network config to NetworkConfig: %s", diags)
	}

	if network != nil {
		t.Fatalf("expected network to be nil, got: %v", network)
	}
}

func TestHCLLoaderParseValidFlatcarConfig(t *testing.T) {
	cfg := `
flatcar {
  channel = "beta"
	version = "2303.1.0"
}
`
	cfgMap := map[string][]byte{
		"flatcar.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid flatcar configuration should not produce any errors: %s", diags)
	}

	flatcar, diags := loadFlatcarConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing network config to NetworkConfig: %s", diags)
	}

	if flatcar.Channel != "beta" {
		t.Fatalf("expected 'channel' to be 'beta', got: %s", flatcar.Channel)
	}

	if flatcar.Version != "2303.1.0" {
		t.Fatalf("expected 'version' to be '2303.1.0', got: %s", flatcar.Version)
	}
}

func TestHCLLoaderParseFlatcarConfigDefaultConfig(t *testing.T) {
	cfg := `
flatcar {
}
`
	cfgMap := map[string][]byte{
		"flatcar.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	flatcar, diags := loadFlatcarConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing network config to NetworkConfig: %s", diags)
	}

	if flatcar.Channel != "stable" {
		t.Fatalf("expected 'channel' to be 'stable', got: %s", flatcar.Channel)
	}

	if flatcar.Version != "current" {
		t.Fatalf("expected 'version' to be 'current', got: %s", flatcar.Version)
	}
}

func TestHCLLoaderParseValidFlatcarConfigUnexpectedArgument(t *testing.T) {
	cfg := `
flatcar {
  channel = "beta"
	version = "2303.1.0"
	test_field = "test-field"
}
`
	cfgMap := map[string][]byte{
		"flatcar.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	flatcar, diags := loadFlatcarConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("Error parsing network config to NetworkConfig: %s", diags)
	}

	if flatcar != nil {
		t.Fatalf("expected flatcar to be nil, got: %v", flatcar)
	}
}

func TestHCLLoaderParseValidLocalBackendConfig(t *testing.T) {
	cfg := `
backend "local" {
	path = "test/local/backend/path"
}
`
	cfgMap := map[string][]byte{
		"backend.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	localBackend, diags := loadBackendConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing backend config to BackendConfig: %s", diags)
	}

	if localBackend == nil {
		t.Fatalf("expected 'backend' to be not nil, got: %s", localBackend)
	}
}

func TestHCLLoaderParseLocalBackendConfigUnsupportedArgument(t *testing.T) {
	cfg := `
backend "local" {
	path = "test/local/backend/path"
	test_field = "test-field"
}
`
	cfgMap := map[string][]byte{
		"backend.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	localBackend, diags := loadBackendConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("expected unsupported argument, got: %s", diags)
	}

	if localBackend != nil {
		t.Fatalf("expected 'backend' to be nil, got: %s", localBackend)
	}
}

func TestHCLLoaderParseValidS3BackendConfig(t *testing.T) {
	cfg := `
backend "s3" {
  bucket = "test-bucket"
	key = "test-key"
	region = "test-region"
	aws_creds_path = "/path/to/aws/creds"
	dynamodb_table = "test-dynamodb-table"
}
`
	cfgMap := map[string][]byte{
		"backend.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	s3Backend, diags := loadBackendConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("Error parsing backend config to BackendConfig: %s", diags)
	}

	if s3Backend == nil {
		t.Fatalf("expected 'backend' to be not nil, got: %s", s3Backend)
	}
}

func TestHCLLoaderParseS3BackendConfigUnsupportedArgument(t *testing.T) {
	cfg := `
backend "s3" {
  bucket = "test-bucket"
	key = "test-key"
	region = "test-region"
	aws_creds_path = "/path/to/aws/creds"
	dynamodb_table = "test-dynamodb-table"
  test_field = "test-field"
}
`
	cfgMap := map[string][]byte{
		"backend.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	s3Backend, diags := loadBackendConfiguration(config)
	if !diags.HasErrors() {
		t.Fatalf("expected unsupported argument, got: %s", diags)
	}

	if s3Backend != nil {
		t.Fatalf("expected 'backend' to be nil, got: %s", s3Backend)
	}
}

func TestHCLLoaderParsePlatformConfigValidPacketConfig(t *testing.T) {
	cfg := `
platform "packet" {
  project_id = "test-project-id"
  facility = "ams1"
  network {
    management_cidrs = ["0.0.0.0/0"]
    node_private_cidr = "10.0.0.0/8"
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
	cfgMap := map[string][]byte{
		"platform.lokocfg": []byte(cfg),
	}
	varconfigMap := map[string][]byte{}

	config, diags := ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	packet, diags := loadPlatformConfiguration(config)
	if diags.HasErrors() {
		t.Fatalf("expected configuration to load successfully, got: %s", diags)
	}

	if packet == nil {
		t.Fatalf("expected 'packet' to not be nil, got: %s", packet)
	}
}
