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
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokomotive/pkg/oidc"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/pkg/errors"
)

type Config struct {
	AssetDir                 string            `hcl:"asset_dir"`
	CachedInstall            string            `hcl:"cached_install,optional"`
	ClusterName              string            `hcl:"cluster_name"`
	ControllerDomains        []string          `hcl:"controller_domains"`
	ControllerMacs           []string          `hcl:"controller_macs"`
	ControllerNames          []string          `hcl:"controller_names"`
	DisableSelfHostedKubelet bool              `hcl:"disable_self_hosted_kubelet,optional"`
	K8sDomainName            string            `hcl:"k8s_domain_name"`
	MatchboxCAPath           string            `hcl:"matchbox_ca_path"`
	MatchboxClientCertPath   string            `hcl:"matchbox_client_cert_path"`
	MatchboxClientKeyPath    string            `hcl:"matchbox_client_key_path"`
	MatchboxEndpoint         string            `hcl:"matchbox_endpoint"`
	MatchboxHTTPEndpoint     string            `hcl:"matchbox_http_endpoint"`
	OSChannel                string            `hcl:"os_channel,optional"`
	OSVersion                string            `hcl:"os_version,optional"`
	SSHPubKeys               []string          `hcl:"ssh_pubkeys"`
	WorkerNames              []string          `hcl:"worker_names"`
	WorkerMacs               []string          `hcl:"worker_macs"`
	WorkerDomains            []string          `hcl:"worker_domains"`
	Labels                   map[string]string `hcl:"labels,optional"`
	OIDC                     *oidc.Config      `hcl:"oidc,block"`
	KubeAPIServerExtraFlags  []string
}

// NewConfig creates a new Config and returns a pointer to it as well as any HCL diagnostics.
func NewConfig(b *hcl.Body, ctx *hcl.EvalContext) (*Config, hcl.Diagnostics) {
	diags := hcl.Diagnostics{}

	// Create config with default values.
	c := &Config{
		CachedInstall: "false",
		OSChannel:     "flatcar-stable",
		OSVersion:     "current",
	}

	if b == nil {
		return nil, hcl.Diagnostics{}
	}

	if d := gohcl.DecodeBody(*b, ctx, c); len(d) != 0 {
		diags = append(diags, d...)
		return nil, diags
	}

	if d := c.validate(); len(d) != 0 {
		diags = append(diags, d...)
		return nil, diags
	}

	return c, diags
}

// Cluster implements the Cluster interface for Packet.
type Cluster struct {
	config *Config
	// A string containing the rendered Terraform code of the root module.
	rootModule string
}

func (c *Cluster) AssetDir() string {
	return c.config.AssetDir
}

func (c *Cluster) ControlPlaneCharts() []string {
	charts := platform.CommonControlPlaneCharts
	if !c.config.DisableSelfHostedKubelet {
		charts = append(charts, "kubelet")
	}

	return charts
}

func (c *Cluster) Managed() bool {
	return false
}

func (c *Cluster) Nodes() int {
	return len(c.config.ControllerMacs) + len(c.config.WorkerMacs)
}

func (c *Cluster) TerraformExecutionPlan() []terraform.ExecutionStep {
	return []terraform.ExecutionStep{
		terraform.ExecutionStep{
			Description: "Create infrastructure",
			Args:        []string{"apply", "-auto-approve"},
		},
	}
}

func (c *Cluster) TerraformRootModule() string {
	return c.rootModule
}

// NewCluster constructs a Cluster based on the provided config and returns a pointer to it.
func NewCluster(c *Config) (*Cluster, error) {
	rendered, err := renderRootModule(c)
	if err != nil {
		return nil, fmt.Errorf("rendering root module: %v", err)
	}

	return &Cluster{config: c, rootModule: rendered}, nil
}

func renderRootModule(conf *Config) (string, error) {
	t, err := template.New("rootModule").Parse(terraformConfigTmpl)
	if err != nil {
		return "", fmt.Errorf("parsing template: %v", err)
	}

	// Configure oidc flags and set it to KubeAPIServerExtraFlags.
	if conf.OIDC != nil {
		// Skipping the error checking here because its done in checkValidConfig().
		oidcFlags, _ := conf.OIDC.ToKubeAPIServerFlags(conf.K8sDomainName)
		//TODO: Use append instead of setting the oidcFlags to KubeAPIServerExtraFlags
		// append is not used for now because Initialize is called in cli/cmd/cluster.go
		// and again in Apply which duplicates the values.
		conf.KubeAPIServerExtraFlags = oidcFlags
	}

	keyListBytes, err := json.Marshal(conf.SSHPubKeys)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal SSH public keys")
	}

	workerDomains, err := json.Marshal(conf.WorkerDomains)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", conf.WorkerDomains)
	}

	workerMacs, err := json.Marshal(conf.WorkerMacs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", conf.WorkerMacs)
	}

	workerNames, err := json.Marshal(conf.WorkerNames)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", conf.WorkerNames)
	}

	controllerDomains, err := json.Marshal(conf.ControllerDomains)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", conf.ControllerDomains)
	}

	controllerMacs, err := json.Marshal(conf.ControllerMacs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", conf.ControllerMacs)
	}

	controllerNames, err := json.Marshal(conf.ControllerNames)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %q", conf.ControllerNames)
	}

	terraformCfg := struct {
		CachedInstall            string
		ClusterName              string
		ControllerDomains        string
		ControllerMacs           string
		ControllerNames          string
		K8sDomainName            string
		MatchboxClientCert       string
		MatchboxClientKey        string
		MatchboxCA               string
		MatchboxEndpoint         string
		MatchboxHTTPEndpoint     string
		OSChannel                string
		OSVersion                string
		SSHPublicKeys            string
		WorkerNames              string
		WorkerMacs               string
		WorkerDomains            string
		DisableSelfHostedKubelet bool
		KubeAPIServerExtraFlags  []string
		Labels                   map[string]string
	}{
		CachedInstall:            conf.CachedInstall,
		ClusterName:              conf.ClusterName,
		ControllerDomains:        string(controllerDomains),
		ControllerMacs:           string(controllerMacs),
		ControllerNames:          string(controllerNames),
		K8sDomainName:            conf.K8sDomainName,
		MatchboxCA:               conf.MatchboxCAPath,
		MatchboxClientCert:       conf.MatchboxClientCertPath,
		MatchboxClientKey:        conf.MatchboxClientKeyPath,
		MatchboxEndpoint:         conf.MatchboxEndpoint,
		MatchboxHTTPEndpoint:     conf.MatchboxHTTPEndpoint,
		OSChannel:                conf.OSChannel,
		OSVersion:                conf.OSVersion,
		SSHPublicKeys:            string(keyListBytes),
		WorkerNames:              string(workerNames),
		WorkerMacs:               string(workerMacs),
		WorkerDomains:            string(workerDomains),
		DisableSelfHostedKubelet: conf.DisableSelfHostedKubelet,
		KubeAPIServerExtraFlags:  conf.KubeAPIServerExtraFlags,
		Labels:                   conf.Labels,
	}

	var rendered bytes.Buffer
	if err := t.Execute(&rendered, terraformCfg); err != nil {
		return "", fmt.Errorf("rendering template: %v", err)
	}

	return rendered.String(), nil
}

// validate validates cluster configuration.
func (c *Config) validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if c.OIDC != nil {
		_, diags := c.OIDC.ToKubeAPIServerFlags(c.K8sDomainName)
		diagnostics = append(diagnostics, diags...)
	}

	return diagnostics
}
