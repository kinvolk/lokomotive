package aws

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/destroy"
	"github.com/kinvolk/lokoctl/pkg/platform"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type config struct {
	AssetDir                 string   `hcl:"asset_dir"`
	ClusterName              string   `hcl:"cluster_name"`
	OSImage                  string   `hcl:"os_image,optional"`
	DNSZone                  string   `hcl:"dns_zone"`
	DNSZoneID                string   `hcl:"dns_zone_id"`
	SSHPubKey                string   `hcl:"ssh_pubkey"`
	CredsPath                string   `hcl:"creds_path,optional"`
	ControllerCount          int      `hcl:"controller_count,optional"`
	ControllerType           string   `hcl:"controller_type,optional"`
	WorkerCount              int      `hcl:"worker_count,optional"`
	WorkerType               string   `hcl:"worker_type,optional"`
	ControllerCLCSnippets    []string `hcl:"controller_clc_snippets,optional"`
	WorkerCLCSnippets        []string `hcl:"worker_clc_snippets,optional"`
	Region                   string   `hcl:"region,optional"`
	EnableAggregation        bool     `hcl:"enable_aggregation,optional"`
	DiskSize                 string   `hcl:"disk_size,optional"`
	DiskType                 string   `hcl:"disk_type,optional"`
	DiskIOPS                 string   `hcl:"disk_iops,optional"`
	WorkerPrice              string   `hcl:"worker_price,optional"`
	WorkerTargetGroups       []string `hcl:"worker_target_groups,optional"`
	Networking               string   `hcl:"networking,optional"`
	NetworkMTU               string   `hcl:"network_mtu,optional"`
	HostCIDR                 string   `hcl:"host_cidr,optional"`
	PodCIDR                  string   `hcl:"pod_cidr,optional"`
	ServiceCIDR              string   `hcl:"service_cidr,optional"`
	ClusterDomainSuffix      string   `hcl:"cluster_domain_suffix,optional"`
	EnableReporting          bool     `hcl:"enable_reporting,optional"`
	CertsValidityPeriodHours int      `hcl:"certs_validity_period_hours,optional"`
}

// init registers aws as a platform
func init() {
	platform.Register("aws", NewConfig())
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func NewConfig() *config {
	return &config{
		OSImage:         "flatcar-stable",
		ControllerCount: 1,
		ControllerType:  "t3.small",
		WorkerCount:     2,
		WorkerType:      "t3.small",
		Region:          "eu-central-1",
		// Initialize the string slices to make sure they are
		// rendered as `[]` when no snippets are given and not
		// `null`, as the latter would lead to a terraform error
		ControllerCLCSnippets: make([]string, 0),
		WorkerCLCSnippets:     make([]string, 0),
		WorkerTargetGroups:    make([]string, 0),
	}
}

// GetAssetDir returns asset directory path
func (c *config) GetAssetDir() string {
	return c.AssetDir
}

func (cfg *config) readSSHPubKey() (string, error) {
	dat, err := ioutil.ReadFile(cfg.SSHPubKey)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(dat)), nil
}

func (cfg *config) Install() error {
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

func createTerraformConfigFile(cfg *config, terraformRootDir string) error {
	tmplName := "cluster.tf"
	t := template.New(tmplName)
	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	path := filepath.Join(terraformRootDir, tmplName)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", path)
	}
	defer f.Close()

	ssh_authorized_key, err := cfg.readSSHPubKey()
	if err != nil {
		return errors.Wrapf(err, "failed to read ssh public key: %s", cfg.SSHPubKey)
	}

	controllerCLCSnippetsBytes, err := json.Marshal(cfg.ControllerCLCSnippets)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	workerCLCSnippetsBytes, err := json.Marshal(cfg.WorkerCLCSnippets)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	workerTargetGroupsBytes, err := json.Marshal(cfg.WorkerTargetGroups)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	terraformCfg := struct {
		Config                config
		SSHAuthorizedKey      string
		ControllerCLCSnippets string
		WorkerCLCSnippets     string
		WorkerTargetGroups    string
	}{
		Config:                *cfg,
		SSHAuthorizedKey:      ssh_authorized_key,
		ControllerCLCSnippets: string(controllerCLCSnippetsBytes),
		WorkerCLCSnippets:     string(workerCLCSnippetsBytes),
		WorkerTargetGroups:    string(workerTargetGroupsBytes),
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}

// Destroy destroys the AWS cluster.
func (cfg *config) Destroy() error {
	return destroy.ExecuteTerraformDestroy(cfg.AssetDir)
}
