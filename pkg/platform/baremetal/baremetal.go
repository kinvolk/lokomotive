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

	"github.com/hashicorp/hcl/v2"
	configpkg "github.com/kinvolk/lokomotive/pkg/cluster/config"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/kinvolk/lokomotive/pkg/util"
	"github.com/pkg/errors"
)

type controller struct {
	Domains []string `hcl:"controller_domains"`
	MACs    []string `hcl:"controller_macs"`
	Names   []string `hcl:"controller_names"`
}

type worker struct {
	Names   []string `hcl:"worker_names"`
	MACs    []string `hcl:"worker_macs"`
	Domains []string `hcl:"worker_domains"`
}

type matchbox struct {
	CAPath         string `hcl:"matchbox_ca_path"`
	ClientCertPath string `hcl:"matchbox_client_cert_path"`
	ClientKeyPath  string `hcl:"matchbox_client_key_path"`
	Endpoint       string `hcl:"matchbox_endpoint"`
	HTTPEndpoint   string `hcl:"matchbox_http_endpoint"`
}

type flatcar struct {
	OSChannel string `hcl:"os_channel,optional"`
	OSVersion string `hcl:"os_version,optional"`
}

type config struct {
	Metadata      *configpkg.Metadata
	Controller    *controller `hcl:"controller,block"`
	Worker        *worker     `hcl:"worker,block"`
	Matchbox      *matchbox   `hcl:"matchbox,block"`
	Flatcar       *flatcar    `hcl:"flatcar,block"`
	CachedInstall string      `hcl:"cached_install,optional"`
	K8sDomainName string      `hcl:"k8s_domain_name"`
}

// init registers bare-metal as a platform
//nolint:gochecknoinits
func init() {
	configpkg.Register("bare-metal", newConfig())
}

// newConfig returns an instance on baremetal specific config.
func newConfig() *config {
	return &config{
		CachedInstall: "false",
		Flatcar: &flatcar{
			OSChannel: "flatcar-stable",
			OSVersion: "current",
		},
	}
}

func (c *config) Render(cfg *configpkg.LokomotiveConfig) (string, error) {
	keyListBytes, err := json.Marshal(cfg.Controller.SSHPubKeys)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal SSH public keys")
	}

	workerDomains, err := json.Marshal(c.Worker.Domains)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.Worker.Domains)
	}

	workerMACs, err := json.Marshal(c.Worker.MACs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.Worker.MACs)
	}

	workerNames, err := json.Marshal(c.Worker.Names)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.Worker.Names)
	}

	controllerDomains, err := json.Marshal(c.Controller.Domains)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.Controller.Domains)
	}

	controllerMACs, err := json.Marshal(c.Controller.MACs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.Controller.MACs)
	}

	controllerNames, err := json.Marshal(c.Controller.Names)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", c.Controller.Names)
	}

	terraformCfg := struct {
		Config            *config
		ControllerDomains string
		ControllerMACs    string
		ControllerNames   string
		SSHPubKeys        string
		WorkerNames       string
		WorkerMACs        string
		WorkerDomains     string
	}{
		Config:            c,
		ControllerDomains: string(controllerDomains),
		ControllerMACs:    string(controllerMACs),
		ControllerNames:   string(controllerNames),
		SSHPubKeys:        string(keyListBytes),
		WorkerNames:       string(workerNames),
		WorkerMACs:        string(workerMACs),
		WorkerDomains:     string(workerDomains),
	}

	return util.RenderTemplate(terraformConfigTmpl, terraformCfg)
}

func (c *config) SetMetadata(metadata *configpkg.Metadata) {
	c.Metadata = metadata
}

func (c *config) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (c *config) Apply(ex *terraform.Executor) error {
	return ex.Apply()
}

func (c *config) Destroy(ex *terraform.Executor) error {
	return ex.Destroy()
}

func (c *config) GetExpectedNodes(cfg *configpkg.LokomotiveConfig) int {
	return len(c.Controller.MACs) + len(c.Worker.MACs)
}
