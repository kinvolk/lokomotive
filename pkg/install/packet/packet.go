package packet

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/tar"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type config struct {
	AssetDir        string
	AuthToken       string
	AWSCredsPath    string
	AWSRegion       string
	ClusterName     string
	ControllerCount int
	ControllerType  string
	DNSZone         string
	DNSZoneID       string
	Facility        string
	ProjectID       string
	SSHPubKey       string
	WorkerCount     int
	WorkerType      string
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
	if err := tar.UntarFromAsset(Asset, "lokomotive-kubernetes-packet.tar.gz", cfg.AssetDir); err != nil {
		return errors.Wrapf(err, "failed to extract Packet config at: %s", cfg.AssetDir)
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

	source := filepath.Join(cfg.AssetDir, "lokomotive-kubernetes/packet/flatcar-linux/kubernetes")

	sshPubKey, err := cfg.readSSHPubKey()
	if err != nil {
		return errors.Wrapf(err, "failed to read ssh public key: %s", cfg.SSHPubKey)
	}

	// Terraform module expects a list of SSH keys - wrap the single key we have in a list.
	// TODO - add support for multiple keys?
	sshKeys, err := json.Marshal([]string{sshPubKey})
	if err != nil {
		return errors.Wrap(err, "constructing SSH key list")
	}

	terraformCfg := struct {
		AssetDir        string
		AWSRegion       string
		ClusterName     string
		ControllerCount int
		ControllerType  string
		CredsPath       string
		DNSZone         string
		DNSZoneID       string
		Facility        string
		ProjectID       string
		Source          string
		SSHKeys         string
		WorkerCount     int
		WorkerType      string
	}{
		AssetDir:        cfg.AssetDir,
		AWSRegion:       cfg.AWSRegion,
		ClusterName:     cfg.ClusterName,
		ControllerCount: cfg.ControllerCount,
		ControllerType:  cfg.ControllerType,
		CredsPath:       cfg.CredsPath,
		DNSZone:         cfg.DNSZone,
		DNSZoneID:       cfg.DNSZoneID,
		Facility:        cfg.Facility,
		ProjectID:       cfg.ProjectID,
		Source:          source,
		SSHKeys:         string(sshKeys),
		WorkerCount:     cfg.WorkerCount,
		WorkerType:      cfg.WorkerType,
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}
