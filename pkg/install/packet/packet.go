package packet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/install"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type config struct {
	AssetDir string `hcl:"asset_dir"`
	// TODO AuthToken gets written to disk when Terraform files are generated. We should consider
	// reading this value directly from the environment.
	AuthToken       string   `hcl:"auth_token"`
	AWSCredsPath    string   `hcl:"aws_creds_path"`
	AWSRegion       string   `hcl:"aws_region"`
	ClusterName     string   `hcl:"cluster_name"`
	ControllerCount int      `hcl:"controller_count"`
	ControllerType  string   `hcl:"controller_type,optional"`
	DNSZone         string   `hcl:"dns_zone"`
	DNSZoneID       string   `hcl:"dns_zone_id"`
	Facility        string   `hcl:"facility"`
	ProjectID       string   `hcl:"project_id"`
	SSHPubKeys      []string `hcl:"ssh_pubkeys"`
	WorkerCount     int      `hcl:"worker_count"`
	WorkerType      string   `hcl:"worker_type,optional"`
	IPXEScriptURL   string   `hcl:"ipxe_script_url,optional"`
	ManagementCIDRs []string `hcl:"management_cidrs"`
	NodePrivateCIDR string   `hcl:"node_private_cidr"`
	OSChannel       string   `hcl:"os_channel,optional"`
	OSVersion       string   `hcl:"os_version,optional"`
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func NewConfig() *config {
	nodeType := "baremetal_0"
	iPXEScriptURL := "https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/master/packet.ipxe"
	return &config{
		ControllerType: nodeType,
		WorkerType:     nodeType,
		IPXEScriptURL:  iPXEScriptURL,
		OSChannel:      "stable",
		OSVersion:      "current",
	}
}

func Install(cfg *config) error {
	terraformModuleDir := filepath.Join(cfg.AssetDir, "lokomotive-kubernetes")
	if err := install.PrepareLokomotiveTerraformModuleAt(terraformModuleDir); err != nil {
		return err
	}

	terraformRootDir := filepath.Join(cfg.AssetDir, "terraform")
	if err := install.PrepareTerraformRootDir(terraformRootDir); err != nil {
		return err
	}

	if err := createTerraformConfigFile(cfg, terraformRootDir); err != nil {
		return err
	}

	return terraform.InitAndApply(terraformRootDir)
}

func createTerraformConfigFile(cfg *config, terraformPath string) error {
	tmplName := "cluster.tf"
	t := template.New(tmplName)
	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	path := filepath.Join(terraformPath, tmplName)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", path)
	}
	defer f.Close()

	source := filepath.Join(cfg.AssetDir, "lokomotive-kubernetes/packet/flatcar-linux/kubernetes")

	keyListBytes, err := json.Marshal(cfg.SSHPubKeys)
	if err != nil {
		return errors.Wrap(err, "failed to marshal SSH public keys")
	}

	managementCIDRs, err := json.Marshal(cfg.ManagementCIDRs)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal management CIDRs")
	}

	terraformCfg := struct {
		Config          config
		Source          string
		SSHPublicKeys   string
		ManagementCIDRs string
	}{
		Config:          *cfg,
		Source:          source,
		SSHPublicKeys:   string(keyListBytes),
		ManagementCIDRs: string(managementCIDRs),
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}
