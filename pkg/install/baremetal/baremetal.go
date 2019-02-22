package baremetal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/tar"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type config struct {
	AssetDir               string   `hcl:"asset_dir"`
	CachedInstall          *string  `hcl:"cached_install"`
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
	OSChannel              *string  `hcl:"os_channel"`
	OSVersion              *string  `hcl:"os_version"`
	SSHPubKey              string   `hcl:"ssh_pubkey"`
	WorkerNames            []string `hcl:"worker_names"`
	WorkerMacs             []string `hcl:"worker_macs"`
	WorkerDomains          []string `hcl:"worker_domains"`
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func NewConfig() *config {
	defaultCachedInstall := "false"
	defaultOSChannel := "flatcar-stable"
	defaultOSVersion := "current"
	return &config{
		CachedInstall: &defaultCachedInstall,
		OSChannel:     &defaultOSChannel,
		OSVersion:     &defaultOSVersion,
	}
}

func Install(cfg *config) error {
	terraformPath := filepath.Join(cfg.AssetDir, "terraform")

	// Create assets directory tree.
	if err := os.MkdirAll(terraformPath, 0755); err != nil {
		return errors.Wrapf(err, "failed to create assets directory tree at: %s", terraformPath)
	}

	// TODO: skip if the dir already exists
	if err := tar.UntarFromAsset(Asset, "lokomotive-kubernetes-baremetal.tar.gz", cfg.AssetDir); err != nil {
		return errors.Wrapf(err, "failed to extract bare-metal config at: %s", cfg.AssetDir)
	}

	if err := createTerraformConfigFile(cfg, terraformPath); err != nil {
		return err
	}

	return terraform.InitAndApply(terraformPath)
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

	source := filepath.Join(cfg.AssetDir, "lokomotive-kubernetes/bare-metal/container-linux/kubernetes")

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
		AssetDir             string
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
		Source               string
		SSHAuthorizedKey     string
		WorkerNames          string
		WorkerMacs           string
		WorkerDomains        string
	}{
		AssetDir:             cfg.AssetDir,
		CachedInstall:        *cfg.CachedInstall,
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
		OSChannel:            *cfg.OSChannel,
		OSVersion:            *cfg.OSVersion,
		Source:               source,
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
