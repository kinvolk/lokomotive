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
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/oidc"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

type config struct {
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
	EnableTLSBootstrap       bool              `hcl:"enable_tls_bootstrap,optional"`
	KubeAPIServerExtraFlags  []string
}

// init registers bare-metal as a platform
func init() {
	platform.Register("bare-metal", NewConfig())
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	if diags := gohcl.DecodeBody(*configBody, evalContext, c); diags.HasErrors() {
		return diags
	}

	return c.checkValidConfig()
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
		CachedInstall:      "false",
		OSChannel:          "flatcar-stable",
		OSVersion:          "current",
		EnableTLSBootstrap: true,
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

	// TODO: A transient change which shall be reverted in a follow up PR to handle
	// https://github.com/kinvolk/lokomotive/issues/716.
	// Extract control plane chart files to cluster assets directory.
	for _, c := range platform.CommonControlPlaneCharts() {
		src := filepath.Join(assets.ControlPlaneSource, c.Name)
		dst := filepath.Join(assetDir, "cluster-assets", "charts", c.Namespace, c.Name)
		if err := assets.Extract(src, dst); err != nil {
			return fmt.Errorf("extracting charts: %w", err)
		}
	}

	// TODO: A transient change which shall be reverted in a follow up PR to handle
	// https://github.com/kinvolk/lokomotive/issues/716.
	// Extract self-hosted kubelet chart only when enabled in config.
	if !c.DisableSelfHostedKubelet {
		src := filepath.Join(assets.ControlPlaneSource, "kubelet")
		dst := filepath.Join(assetDir, "cluster-assets", "charts", "kube-system", "kubelet")
		if err := assets.Extract(src, dst); err != nil {
			return fmt.Errorf("extracting kubelet chart: %w", err)
		}
	}

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)

	return createTerraformConfigFile(c, terraformRootDir)
}

func createTerraformConfigFile(cfg *config, terraformPath string) error {
	tmplName := "cluster.tf"
	t := template.New(tmplName)
	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		// TODO: Use template.Must().
		return errors.Wrap(err, "failed to parse template")
	}

	path := filepath.Join(terraformPath, tmplName)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", path, err)
	}

	defer f.Close()

	// Configure oidc flags and set it to KubeAPIServerExtraFlags.
	if cfg.OIDC != nil {
		// Skipping the error checking here because its done in checkValidConfig().
		oidcFlags, _ := cfg.OIDC.ToKubeAPIServerFlags(cfg.K8sDomainName)
		// TODO: Use append instead of setting oidcFlags to KubeAPIServerExtraFlags.
		// Append is not used for now because Initialize is called in cli/cmd/cluster.go
		// and again in Apply which duplicates the values.
		cfg.KubeAPIServerExtraFlags = oidcFlags
	}

	keyListBytes, err := json.Marshal(cfg.SSHPubKeys)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return errors.Wrap(err, "failed to marshal SSH public keys")
	}

	workerDomains, err := json.Marshal(cfg.WorkerDomains)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return errors.Wrapf(err, "failed to parse %q", cfg.WorkerDomains)
	}

	workerMacs, err := json.Marshal(cfg.WorkerMacs)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return errors.Wrapf(err, "failed to parse %q", cfg.WorkerMacs)
	}

	workerNames, err := json.Marshal(cfg.WorkerNames)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return errors.Wrapf(err, "failed to parse %q", cfg.WorkerNames)
	}

	controllerDomains, err := json.Marshal(cfg.ControllerDomains)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return errors.Wrapf(err, "failed to parse %q", cfg.ControllerDomains)
	}

	controllerMacs, err := json.Marshal(cfg.ControllerMacs)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return errors.Wrapf(err, "failed to parse %q", cfg.ControllerMacs)
	}

	controllerNames, err := json.Marshal(cfg.ControllerNames)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return errors.Wrapf(err, "failed to parse %q", cfg.ControllerNames)
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
		EnableTLSBootstrap       bool
	}{
		CachedInstall:            cfg.CachedInstall,
		ClusterName:              cfg.ClusterName,
		ControllerDomains:        string(controllerDomains),
		ControllerMacs:           string(controllerMacs),
		ControllerNames:          string(controllerNames),
		K8sDomainName:            cfg.K8sDomainName,
		MatchboxCA:               cfg.MatchboxCAPath,
		MatchboxClientCert:       cfg.MatchboxClientCertPath,
		MatchboxClientKey:        cfg.MatchboxClientKeyPath,
		MatchboxEndpoint:         cfg.MatchboxEndpoint,
		MatchboxHTTPEndpoint:     cfg.MatchboxHTTPEndpoint,
		OSChannel:                cfg.OSChannel,
		OSVersion:                cfg.OSVersion,
		SSHPublicKeys:            string(keyListBytes),
		WorkerNames:              string(workerNames),
		WorkerMacs:               string(workerMacs),
		WorkerDomains:            string(workerDomains),
		DisableSelfHostedKubelet: cfg.DisableSelfHostedKubelet,
		KubeAPIServerExtraFlags:  cfg.KubeAPIServerExtraFlags,
		Labels:                   cfg.Labels,
		EnableTLSBootstrap:       cfg.EnableTLSBootstrap,
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

// checkValidConfig validates cluster configuration.
func (c *config) checkValidConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if c.OIDC != nil {
		_, diags := c.OIDC.ToKubeAPIServerFlags(c.K8sDomainName)
		diagnostics = append(diagnostics, diags...)
	}

	return diagnostics
}
