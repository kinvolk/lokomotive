// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package baremetal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/kinvolk/lokomotive/pkg/util"
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
	SSHPubKeys             []string `hcl:"ssh_pubkeys"`
	WorkerNames            []string `hcl:"worker_names"`
	WorkerMacs             []string `hcl:"worker_macs"`
	WorkerDomains          []string `hcl:"worker_domains"`

	ControllerDomainsRaw string
	ControllerMacsRaw    string
	ControllerNamesRaw   string
	K8sDomainNameRaw     string
	SSHPubKeysRaw        string
	WorkerNamesRaw       string
	WorkerMacsRaw        string
	WorkerDomainsRaw     string
}

// init registers bare-metal as a platform
func init() {
	platform.Register("bare-metal", NewConfig())
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

func (c *config) Apply(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Apply()
}

func (c *config) Destroy(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Destroy()
}

func (c *config) Initialize(ex *terraform.Executor) error {
	assetDir, err := homedir.Expand(c.AssetDir)
	if err != nil {
		return err
	}

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)

	return createTerraformConfigFile(c, terraformRootDir)
}

func (c *config) Render() (string, error) {
	keyListBytes, err := json.Marshal(c.SSHPubKeys)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal SSH public keys")
	}

	workerDomains, err := json.Marshal(c.WorkerDomains)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.WorkerDomains)
	}

	workerMacs, err := json.Marshal(c.WorkerMacs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.WorkerMacs)
	}

	workerNames, err := json.Marshal(c.WorkerNames)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.WorkerNames)
	}

	controllerDomains, err := json.Marshal(c.ControllerDomains)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.ControllerDomains)
	}

	controllerMacs, err := json.Marshal(c.ControllerMacs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.ControllerMacs)
	}

	controllerNames, err := json.Marshal(c.ControllerNames)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.ControllerNames)
	}

	c.ControllerDomainsRaw = string(controllerDomains)
	c.ControllerMacsRaw = string(controllerMacs)
	c.ControllerNamesRaw = string(controllerNames)
	c.SSHPubKeysRaw = string(keyListBytes)
	c.WorkerNamesRaw = string(workerNames)
	c.WorkerMacsRaw = string(workerMacs)
	c.WorkerDomainsRaw = string(workerDomains)

	return util.RenderTemplate(terraformConfigTmpl, c)
}

func (c *config) Validate() hcl.Diagnostics {

	return hcl.Diagnostics{}
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

	cfg.ControllerDomainsRaw = string(controllerDomains)
	cfg.ControllerMacsRaw = string(controllerMacs)
	cfg.ControllerNamesRaw = string(controllerNames)
	cfg.SSHPubKeysRaw = string(keyListBytes)
	cfg.WorkerNamesRaw = string(workerNames)
	cfg.WorkerMacsRaw = string(workerMacs)
	cfg.WorkerDomainsRaw = string(workerDomains)

	if err := t.Execute(f, cfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}

func (c *config) GetExpectedNodes() int {
	return len(c.ControllerMacs) + len(c.WorkerMacs)
}
