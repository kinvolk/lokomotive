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

	"github.com/kinvolk/lokoctl/pkg/install"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type config struct {
	AssetDir    string `hcl:"asset_dir"`
	ClusterName string `hcl:"cluster_name"`
	OSImage     string `hcl:"os_image,optional"`
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
	return &config{
		OSImage: "flatcar-stable",
	}
}

func (cfg *config) readSSHPubKey() (string, error) {
	dat, err := ioutil.ReadFile(cfg.SSHPubKey)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(dat)), nil
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

	source := filepath.Join(cfg.AssetDir, "lokomotive-kubernetes/aws/container-linux/kubernetes")
	ssh_authorized_key, err := cfg.readSSHPubKey()
	if err != nil {
		return errors.Wrapf(err, "failed to read ssh public key: %s", cfg.SSHPubKey)
	}

	terraformCfg := struct {
		Config           config
		Source           string
		SSHAuthorizedKey string
	}{
		Config:           *cfg,
		Source:           source,
		SSHAuthorizedKey: ssh_authorized_key,
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}
