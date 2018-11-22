package baremetal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/tar"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type config struct {
	AssetDir               string
	CachedInstall          string
	ClusterName            string
	ControllerDomain       string
	ControllerMac          string
	ControllerName         string
	K8sDomainName          string
	MatchboxCAPath         string
	MatchboxClientCertPath string
	MatchboxClientKeyPath  string
	MatchboxEndpoint       string
	MatchboxHTTPEndpoint   string
	OSChannel              string
	OSVersion              string
	SSHPubKeyPath          string
	WorkerNames            []string
	WorkerMacs             []string
	WorkerDomains          []string
}

func NewConfig() *config {
	return &config{}
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
		return errors.Wrapf(err, "Failed to parse %q", cfg.WorkerDomains)
	}

	workerMacs, err := json.Marshal(cfg.WorkerMacs)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse %q", cfg.WorkerMacs)
	}

	workerNames, err := json.Marshal(cfg.WorkerNames)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse %q", cfg.WorkerNames)
	}

	terraformCfg := struct {
		AssetDir             string
		CachedInstall        string
		ClusterName          string
		ControllerDomain     string
		ControllerMac        string
		ControllerName       string
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
		CachedInstall:        cfg.CachedInstall,
		ClusterName:          cfg.ClusterName,
		ControllerDomain:     cfg.ControllerDomain,
		ControllerMac:        cfg.ControllerMac,
		ControllerName:       cfg.ControllerName,
		K8sDomainName:        cfg.K8sDomainName,
		MatchboxCA:           cfg.MatchboxCAPath,
		MatchboxClientCert:   cfg.MatchboxClientCertPath,
		MatchboxClientKey:    cfg.MatchboxClientKeyPath,
		MatchboxEndpoint:     cfg.MatchboxEndpoint,
		MatchboxHTTPEndpoint: cfg.MatchboxHTTPEndpoint,
		OSChannel:            cfg.OSChannel,
		OSVersion:            cfg.OSVersion,
		Source:               source,
		SSHAuthorizedKey:     cfg.SSHPubKeyPath,
		WorkerNames:          string(workerNames),
		WorkerMacs:           string(workerMacs),
		WorkerDomains:        string(workerDomains),
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}
