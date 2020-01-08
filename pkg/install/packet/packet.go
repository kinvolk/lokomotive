package packet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/destroy"
	"github.com/kinvolk/lokoctl/pkg/platform"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type workerPool struct {
	Name           string `hcl:"pool_name,label"`
	Count          int    `hcl:"count"`
	IPXEScriptURL  string `hcl:"ipxe_script_url,optional"`
	OSArch         string `hcl:"os_arch,optional"`
	OSChannel      string `hcl:"os_channel,optional"`
	OSVersion      string `hcl:"os_version,optional"`
	NodeType       string `hcl:"node_type,optional"`
	Labels         string `hcl:"labels,optional"`
	Taints         string `hcl:"taints,optional"`
	SetupRaid      bool   `hcl:"setup_raid,optional"`
	SetupRaidHDD   bool   `hcl:"setup_raid_hdd,optional"`
	SetupRaidSSD   bool   `hcl:"setup_raid_ssd,optional"`
	SetupRaidSSDFS bool   `hcl:"setup_raid_ssd_fs,optional"`
}

type config struct {
	AssetDir                 string            `hcl:"asset_dir"`
	AuthToken                string            `hcl:"auth_token,optional"`
	AWSCredsPath             string            `hcl:"aws_creds_path,optional"`
	AWSRegion                string            `hcl:"aws_region"`
	ClusterName              string            `hcl:"cluster_name"`
	ControllerCount          int               `hcl:"controller_count"`
	ControllerType           string            `hcl:"controller_type,optional"`
	DNSZone                  string            `hcl:"dns_zone"`
	DNSZoneID                string            `hcl:"dns_zone_id"`
	Facility                 string            `hcl:"facility"`
	ProjectID                string            `hcl:"project_id"`
	SSHPubKeys               []string          `hcl:"ssh_pubkeys"`
	OSArch                   string            `hcl:"os_arch,optional"`
	OSChannel                string            `hcl:"os_channel,optional"`
	OSVersion                string            `hcl:"os_version,optional"`
	IPXEScriptURL            string            `hcl:"ipxe_script_url,optional"`
	ManagementCIDRs          []string          `hcl:"management_cidrs"`
	NodePrivateCIDR          string            `hcl:"node_private_cidr"`
	EnableAggregation        bool              `hcl:"enable_aggregation,optional"`
	Networking               string            `hcl:"networking,optional"`
	NetworkMTU               string            `hcl:"network_mtu,optional"`
	PodCIDR                  string            `hcl:"pod_cidr,optional"`
	ServiceCIDR              string            `hcl:"service_cidr,optional"`
	ClusterDomainSuffix      string            `hcl:"cluster_domain_suffix,optional"`
	EnableReporting          bool              `hcl:"enable_reporting,optional"`
	ReservationIDs           map[string]string `hcl:"reservation_ids,optional"`
	ReservationIDsDefault    string            `hcl:"reservation_ids_default,optional"`
	CertsValidityPeriodHours int               `hcl:"certs_validity_period_hours,optional"`

	WorkerPools []workerPool `hcl:"worker_pool,block"`
}

// init registers packet as a platform
func init() {
	platform.Register("packet", NewConfig())
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

// GetAssetDir returns asset directory path
func (c *config) GetAssetDir() string {
	return c.AssetDir
}

func (cfg *config) Install() error {
	if cfg.AuthToken == "" && os.Getenv("PACKET_AUTH_TOKEN") == "" {
		return fmt.Errorf("cannot find the Packet authentication token:\n" +
			"either specify AuthToken or use the PACKET_AUTH_TOKEN environment variable")
	}

	assetDir, err := homedir.Expand(cfg.AssetDir)
	if err != nil {
		return err
	}

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)
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
		SSHPublicKeys   string
		ManagementCIDRs string
	}{
		Config:          *cfg,
		SSHPublicKeys:   string(keyListBytes),
		ManagementCIDRs: string(managementCIDRs),
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}

func (cfg *config) GetExpectedNodes() int {
	workers := 0

	for _, wp := range cfg.WorkerPools {
		workers += wp.Count
	}

	return cfg.ControllerCount + workers
}

// Destroy destroys the Packet cluster.
func (cfg *config) Destroy() error {
	return destroy.ExecuteTerraformDestroy(cfg.AssetDir)
}
