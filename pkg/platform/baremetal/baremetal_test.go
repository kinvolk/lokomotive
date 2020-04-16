package baremetal

import (
	"testing"

	configpkg "github.com/kinvolk/lokomotive/pkg/cluster/config"
	hclconfigpkg "github.com/kinvolk/lokomotive/pkg/config"
)

func TestRenderSuccess(t *testing.T) {
	metadata := &configpkg.Metadata{
		AssetDir:    "/path/to/assetdir",
		ClusterName: "test-cluster",
	}

	controllerconfig := &configpkg.ControllerConfig{
		SSHPubKeys: []string{"test-ssh-key"},
	}

	controller := &controller{
		Names:   []string{"test-name-a", "test-name-b", "test-name-c"},
		Domains: []string{"test-domain-a", "test-domain-b", "test-domain-c"},
		MACs:    []string{"test-mac-a", "test-mac-b", "test-mac-c"},
	}
	worker := &worker{
		Names:   []string{"test-name-a", "test-name-b", "test-name-c"},
		Domains: []string{"test-domain-a", "test-domain-b", "test-domain-c"},
		MACs:    []string{"test-mac-a", "test-mac-b", "test-mac-c"},
	}
	matchbox := &matchbox{
		CAPath:         "ca/path",
		ClientCertPath: "client/cert/path",
		ClientKeyPath:  "client/key/path",
		Endpoint:       "http://test-endpoint.test",
		HTTPEndpoint:   "https://test.endpoint.test",
	}

	flatcar := &flatcar{
		OSChannel: "flatcar-stable",
		OSVersion: "current",
	}

	b := &config{
		Metadata:      metadata,
		CachedInstall: "true",
		K8sDomainName: "cluster.local",
		Controller:    controller,
		Worker:        worker,
		Matchbox:      matchbox,
		Flatcar:       flatcar,
	}

	lokocfg := &configpkg.LokomotiveConfig{
		Metadata:   metadata,
		Controller: controllerconfig,
	}

	renderedTemplate, err := b.Render(lokocfg)
	if err != nil {
		t.Fatalf("Expected render to succeed, got: %v", err)
	}

	if renderedTemplate == "" {
		t.Fatalf("Expected rendered string to be non-empty")
	}
}

//nolint:funlen
func TestRenderHCLConfigSuccess(t *testing.T) {
	cfg := `
metadata {
  cluster_name = "mercury"
  asset_dir = "test-asset-dir-path"
}

controller {
	count = 1
  ssh_pubkeys = ["test-ssh-key"]
}

platform "bare-metal" {
  cached_install = "true"
	k8s_domain_name = "node1.example.com"
	matchbox {
    matchbox_ca_path = pathexpand("~/pxe-testbed/.matchbox/ca.crt")
    matchbox_client_cert_path = pathexpand("~/pxe-testbed/.matchbox/client.crt")
    matchbox_client_key_path = pathexpand("~/pxe-testbed/.matchbox/client.key")
    matchbox_endpoint = "matchbox.example.com:8081"
    matchbox_http_endpoint = "http://matchbox.example.com:8080"
  }
  controller {
    controller_domains = [
      "node1.example.com",
    ]
    controller_macs = [
      "52:54:00:a1:9c:ae",
    ]
    controller_names = [
      "node1",
    ]
  }
	worker {
		worker_domains = [
			"node2.example.com",
			"node3.example.com",
		]
		worker_macs = [
			"52:54:00:b2:2f:86",
			"52:54:00:c3:61:77",
		]
		worker_names = [
			"node2",
			"node3",
		]
	}
}
`
	cfgMap := map[string][]byte{
		"baremetal.lokocfg": []byte(cfg),
	}

	varconfigMap := map[string][]byte{}

	hclConfig, diags := hclconfigpkg.ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	if diags := hclconfigpkg.ValidateHCLConfig(hclConfig); diags.HasErrors() {
		t.Fatalf("Loading valid hcl configuration should not produce any errors: %s", diags)
	}

	lokocfg, diags := hclconfigpkg.ParseToLokomotiveConfig(hclConfig)
	if diags.HasErrors() {
		t.Fatalf("Error parsing valid hcl baremetal config to LokomotiveConfig: %s", diags)
	}

	baremetal := lokocfg.Platform
	baremetal.SetMetadata(lokocfg.Metadata)

	renderedTemplate, err := baremetal.Render(lokocfg)
	if err != nil {
		t.Fatalf("Expected no error on rendering valid config: %v", err)
	}

	if renderedTemplate == "" {
		t.Fatal("Expected rendered template to be non-empty")
	}
}
