package baremetal

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	AssetDir               string   `hcl:"asset_dir"`
	CachedInstall          string   `hcl:"cached_install,optional"`
	ClusterName            string   `hcl:"cluster_name"`
	ControllerDomains      []string `hcl:"controller_domains"`
	ControllerMacs         []string `hcl:"controller_macs"`
	ControllerNames        []string `hcl:"controller_names"`
	K8sDomainName          string   `hcl:"k8s_domain_name"`
	MatchboxCAPath         string   `hcl:"matchbox_ca_path"`
	MatchboxClientCertPath string   `hcl:"matchbox_client_cert_path"`
	MatchboxClientKeyPath  string   `hcl:"matchbox_client_key_path"`
	MatchboxEndpoint       string   `hcl:"matchbox_endpoint"`
	MatchboxHTTPEndpoint   string   `hcl:"matchbox_http_endpoint"`
	OSChannel              string   `hcl:"os_channel,optional"`
	OSVersion              string   `hcl:"os_version,optional"`
	SSHPubKey              string   `hcl:"ssh_pubkey"`
	WorkerNames            []string `hcl:"worker_names"`
	WorkerMacs             []string `hcl:"worker_macs"`
	WorkerDomains          []string `hcl:"worker_domains"`
}

// init registers bare-metal as a platform
func init() {
	platform.Register("bare-metal", NewConfig())
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

// GetAssetDir returns asset directory path
func (c *config) GetAssetDir() string {
	return c.AssetDir
}

func NewConfig() *config {
	return &config{
		CachedInstall: "false",
		OSChannel:     "flatcar-stable",
		OSVersion:     "current",
	}
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

	workerDomains, err := json.Marshal(cfg.WorkerDomains)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %q", cfg.WorkerDomains)
	}

	workerMacs, err := json.Marshal(cfg.WorkerMacs)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %q", cfg.WorkerMacs)
	}

	workerNames, err := json.Marshal(cfg.WorkerNames)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %q", cfg.WorkerNames)
	}

	controllerDomains, err := json.Marshal(cfg.ControllerDomains)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %q", cfg.ControllerDomains)
	}

	controllerMacs, err := json.Marshal(cfg.ControllerMacs)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %q", cfg.ControllerMacs)
	}

	controllerNames, err := json.Marshal(cfg.ControllerNames)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %q", cfg.ControllerNames)
	}

	terraformCfg := struct {
		CachedInstall        string
		ClusterName          string
		ControllerDomains    string
		ControllerMacs       string
		ControllerNames      string
		K8sDomainName        string
		MatchboxClientCert   string
		MatchboxClientKey    string
		MatchboxCA           string
		MatchboxEndpoint     string
		MatchboxHTTPEndpoint string
		OSChannel            string
		OSVersion            string
		SSHAuthorizedKey     string
		WorkerNames          string
		WorkerMacs           string
		WorkerDomains        string
	}{
		CachedInstall:        cfg.CachedInstall,
		ClusterName:          cfg.ClusterName,
		ControllerDomains:    string(controllerDomains),
		ControllerMacs:       string(controllerMacs),
		ControllerNames:      string(controllerNames),
		K8sDomainName:        cfg.K8sDomainName,
		MatchboxCA:           cfg.MatchboxCAPath,
		MatchboxClientCert:   cfg.MatchboxClientCertPath,
		MatchboxClientKey:    cfg.MatchboxClientKeyPath,
		MatchboxEndpoint:     cfg.MatchboxEndpoint,
		MatchboxHTTPEndpoint: cfg.MatchboxHTTPEndpoint,
		OSChannel:            cfg.OSChannel,
		OSVersion:            cfg.OSVersion,
		SSHAuthorizedKey:     cfg.SSHPubKey,
		WorkerNames:          string(workerNames),
		WorkerMacs:           string(workerMacs),
		WorkerDomains:        string(workerDomains),
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}

// Destroy destroys the Baremetal cluster.
func (cfg *config) Destroy() error {
	return destroy.ExecuteTerraformDestroy(cfg.AssetDir)
}
