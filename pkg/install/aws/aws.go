package aws

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/tar"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type config struct {
	AssetDir    string `hcl:"asset_dir"`
	ClusterName string `hcl:"cluster_name"`
	DNSZone     string `hcl:"dns_zone"`
	DNSZoneID   string `hcl:"dns_zone_id"`
	SSHPubKey   string `hcl:"ssh_pubkey"`
	CredsPath   string `hcl:"creds_path"`
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func NewConfig() *config {
	return &config{}
}

func (cfg *config) readSSHPubKey() (string, error) {
	dat, err := ioutil.ReadFile(cfg.SSHPubKey)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(dat)), nil
}

func Install(cfg *config) error {
	terraformPath := filepath.Join(cfg.AssetDir, "terraform")

	// Create assets directory tree.
	if err := os.MkdirAll(terraformPath, 0755); err != nil {
		return errors.Wrapf(err, "failed to create assets directory tree at: %s", terraformPath)
	}

	// TODO: skip if the dir already exists
	if err := tar.UntarFromAsset(Asset, "lokomotive-kubernetes-aws.tar.gz", cfg.AssetDir); err != nil {
		return errors.Wrapf(err, "failed to extract AWS config at: %s", cfg.AssetDir)
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

	source := filepath.Join(cfg.AssetDir, "lokomotive-kubernetes/aws/container-linux/kubernetes")
	ssh_authorized_key, err := cfg.readSSHPubKey()
	if err != nil {
		return errors.Wrapf(err, "failed to read ssh public key: %s", cfg.SSHPubKey)
	}

	terraformCfg := struct {
		AssetDir         string
		Source           string
		ClusterName      string
		DNSZone          string
		DNSZoneID        string
		SSHAuthorizedKey string
		CredsPath        string
	}{
		AssetDir:         cfg.AssetDir,
		Source:           source,
		ClusterName:      cfg.ClusterName,
		DNSZone:          cfg.DNSZone,
		DNSZoneID:        cfg.DNSZoneID,
		SSHAuthorizedKey: ssh_authorized_key,
		CredsPath:        cfg.CredsPath,
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}
