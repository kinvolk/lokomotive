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
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
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

// Meta is part of Platform interface and returns common information about the platform configuration.
func (c *config) Meta() platform.Meta {
	return platform.Meta{
		AssetDir:      c.AssetDir,
		ExpectedNodes: len(c.ControllerMacs) + len(c.WorkerMacs),
	}
}

func NewConfig() *config {
	return &config{
		CachedInstall: "false",
		OSChannel:     "flatcar-stable",
		OSVersion:     "current",
	}
}

func (c *config) Apply(ex *terraform.Executor) error {
	return ex.Apply()
}

func (c *config) Destroy(ex *terraform.Executor) error {
	return ex.Destroy()
}

//nolint:funlen
func (c *config) Render() (string, error) {
	keyListBytes, err := json.Marshal(c.SSHPubKeys)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SSH public keys: %w", err)
	}

	workerDomains, err := json.Marshal(c.WorkerDomains)
	if err != nil {
		return "", fmt.Errorf("failed to parse '%q', got: %w", c.WorkerDomains, err)
	}

	workerMacs, err := json.Marshal(c.WorkerMacs)
	if err != nil {
		return "", fmt.Errorf("failed to parse '%q', got: %w", c.WorkerMacs, err)
	}

	workerNames, err := json.Marshal(c.WorkerNames)
	if err != nil {
		return "", fmt.Errorf("failed to parse '%q', got: %w", c.WorkerNames, err)
	}

	controllerDomains, err := json.Marshal(c.ControllerDomains)
	if err != nil {
		return "", fmt.Errorf("failed to parse '%q', got: %w", c.ControllerDomains, err)
	}

	controllerMacs, err := json.Marshal(c.ControllerMacs)
	if err != nil {
		return "", fmt.Errorf("failed to parse '%q', got: %w", c.ControllerMacs, err)
	}

	controllerNames, err := json.Marshal(c.ControllerNames)
	if err != nil {
		return "", fmt.Errorf("failed to parse '%q', got: %w", c.ControllerNames, err)
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
		SSHPublicKeys        string
		WorkerNames          string
		WorkerMacs           string
		WorkerDomains        string
	}{
		CachedInstall:        c.CachedInstall,
		ClusterName:          c.ClusterName,
		ControllerDomains:    string(controllerDomains),
		ControllerMacs:       string(controllerMacs),
		ControllerNames:      string(controllerNames),
		K8sDomainName:        c.K8sDomainName,
		MatchboxCA:           c.MatchboxCAPath,
		MatchboxClientCert:   c.MatchboxClientCertPath,
		MatchboxClientKey:    c.MatchboxClientKeyPath,
		MatchboxEndpoint:     c.MatchboxEndpoint,
		MatchboxHTTPEndpoint: c.MatchboxHTTPEndpoint,
		OSChannel:            c.OSChannel,
		OSVersion:            c.OSVersion,
		SSHPublicKeys:        string(keyListBytes),
		WorkerNames:          string(workerNames),
		WorkerMacs:           string(workerMacs),
		WorkerDomains:        string(workerDomains),
	}

	return template.Render(terraformConfigTmpl, terraformCfg)
}
