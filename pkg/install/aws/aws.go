package aws

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hpcloud/tail"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/tar"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

type config struct {
	AssetDir    string
	ClusterName string
	DNSZone     string
	DNSZoneID   string
	SSHPubKey   string
	CredsPath   string
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
	if err := extractAWSConfig(cfg.AssetDir); err != nil {
		return errors.Wrapf(err, "failed to extract AWS config at: %s", cfg.AssetDir)
	}

	if err := createTerraformConfigFile(cfg, terraformPath); err != nil {
		return err
	}

	return installUsingTerraform(terraformPath)
}

func extractAWSConfig(path string) error {
	tarFile, err := Asset("lokomotive-kubernetes.tar.gz")
	if err != nil {
		return err
	}

	tarFileReader := bytes.NewReader(tarFile)
	return tar.Untar(tarFileReader, path)
}

func createTerraformConfigFile(cfg *config, terraformPath string) error {
	tmplData, err := ioutil.ReadFile("templates/aws")
	if err != nil {
		return err
	}

	tmplName := "cluster.tf"
	t := template.New(tmplName)
	t, err = t.Parse(string(tmplData))
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

func installUsingTerraform(exPath string) error {
	ex, err := terraform.NewExecutor(exPath)
	if err != nil {
		return errors.Wrap(err, "failed to create terraform executor")
	}

	if err := executeTerraform(ex, "init", "-no-color"); err != nil {
		return err
	}

	if err := executeTerraform(ex, "apply", "-auto-approve", "-no-color"); err != nil {
		return err
	}
	return nil
}

func executeTerraform(ex *terraform.Executor, args ...string) error {
	pid, done, err := ex.Execute(args...)
	if err != nil {
		return errors.Wrapf(err, "failed to run 'terraform %s'", strings.Join(args, " "))
	}

	pathToFile := filepath.Join(ex.WorkingDirectory(), "logs", fmt.Sprintf("%d%s", pid, ".log"))
	t, err := tail.TailFile(pathToFile, tail.Config{Follow: true})
	if err != nil {
		return err
	}

	go func() {
		for line := range t.Lines {
			fmt.Println(line.Text)
		}
	}()

	<-done
	return t.Stop()
}
