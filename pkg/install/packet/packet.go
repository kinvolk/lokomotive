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

type workerPool struct {
	Name      string `hcl:"pool_name,label"`
	Count     int    `hcl:"count"`
	OSChannel string `hcl:"os_channel,optional"`
	OSVersion string `hcl:"os_version,optional"`
	NodeType  string `hcl:"node_type,optional"`
}

type config struct {
	AssetDir string `hcl:"asset_dir"`
	// TODO AuthToken gets written to disk when Terraform files are generated. We should consider
	// reading this value directly from the environment.
	AuthToken         string   `hcl:"auth_token"`
	AWSCredsPath      string   `hcl:"aws_creds_path"`
	AWSRegion         string   `hcl:"aws_region"`
	ClusterName       string   `hcl:"cluster_name"`
	ControllerCount   int      `hcl:"controller_count"`
	ControllerType    string   `hcl:"controller_type,optional"`
	DNSZone           string   `hcl:"dns_zone"`
	DNSZoneID         string   `hcl:"dns_zone_id"`
	Facility          string   `hcl:"facility"`
	ProjectID         string   `hcl:"project_id"`
	SSHPubKeys        []string `hcl:"ssh_pubkeys"`
	IPXEScriptURL     string   `hcl:"ipxe_script_url,optional"`
	ManagementCIDRs   []string `hcl:"management_cidrs"`
	NodePrivateCIDR   string   `hcl:"node_private_cidr"`
	EnableAggregation string   `hcl:"enable_aggregation,optional"`

	WorkerPools []workerPool `hcl:"worker_pool,block"`
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	if diags := gohcl.DecodeBody(*configBody, evalContext, c); len(diags) != 0 {
		return diags
	}
	if len(c.WorkerPools) == 0 {
		err := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "At least one worker pool must be defined",
			Detail:   "Make sure to define at least one worker pool block in your cluster block",
		}
		return hcl.Diagnostics{err}
	}
	return nil
}

func NewConfig() *config {
	return &config{}
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

	var workerCount int
	for _, pool := range cfg.WorkerPools {
		workerCount += pool.Count
	}

	terraformCfg := struct {
		Config          config
		Source          string
		SSHPublicKeys   string
		ManagementCIDRs string
		WorkerCount     int
	}{
		Config:          *cfg,
		Source:          source,
		SSHPublicKeys:   string(keyListBytes),
		ManagementCIDRs: string(managementCIDRs),
		WorkerCount:     workerCount,
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}
