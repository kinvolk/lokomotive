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

	"github.com/kinvolk/lokomotive/pkg/oidc"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

// Labels represent the map of key value string pairs added the kubelet.
type Labels map[string]string

type config struct {
	AssetDir                     string              `hcl:"asset_dir"`
	CachedInstall                string              `hcl:"cached_install,optional"`
	ClusterName                  string              `hcl:"cluster_name"`
	ControllerDomains            []string            `hcl:"controller_domains"`
	ControllerMacs               []string            `hcl:"controller_macs"`
	ControllerNames              []string            `hcl:"controller_names"`
	DisableSelfHostedKubelet     bool                `hcl:"disable_self_hosted_kubelet,optional"`
	K8sDomainName                string              `hcl:"k8s_domain_name"`
	MatchboxCAPath               string              `hcl:"matchbox_ca_path"`
	MatchboxClientCertPath       string              `hcl:"matchbox_client_cert_path"`
	MatchboxClientKeyPath        string              `hcl:"matchbox_client_key_path"`
	MatchboxEndpoint             string              `hcl:"matchbox_endpoint"`
	MatchboxHTTPEndpoint         string              `hcl:"matchbox_http_endpoint"`
	NetworkMTU                   int                 `hcl:"network_mtu,optional"`
	OSChannel                    string              `hcl:"os_channel,optional"`
	OSVersion                    string              `hcl:"os_version,optional"`
	SSHPubKeys                   []string            `hcl:"ssh_pubkeys"`
	WorkerNames                  []string            `hcl:"worker_names"`
	WorkerMacs                   []string            `hcl:"worker_macs"`
	WorkerDomains                []string            `hcl:"worker_domains"`
	Labels                       Labels              `hcl:"labels,optional"`
	NodeSpecificLabels           map[string]Labels   `hcl:"node_specific_labels,optional"`
	OIDC                         *oidc.Config        `hcl:"oidc,block"`
	EncryptPodTraffic            bool                `hcl:"encrypt_pod_traffic,optional"`
	IgnoreX509CNCheck            bool                `hcl:"ignore_x509_cn_check,optional"`
	ConntrackMaxPerCore          int                 `hcl:"conntrack_max_per_core,optional"`
	InstallToSmallestDisk        bool                `hcl:"install_to_smallest_disk,optional"`
	InstallDisk                  string              `hcl:"install_disk,optional"`
	KernelArgs                   []string            `hcl:"kernel_args,optional"`
	KernelConsole                []string            `hcl:"kernel_console,optional"`
	DownloadProtocol             string              `hcl:"download_protocol,optional"`
	NetworkIPAutodetectionMethod string              `hcl:"network_ip_autodetection_method,optional"`
	CLCSnippets                  map[string][]string `hcl:"clc_snippets,optional"`
	InstallerCLCSnippets         map[string][]string `hcl:"installer_clc_snippets,optional"`
	CertsValidityPeriodHours     int                 `hcl:"certs_validity_period_hours,optional"`
	WipeAdditionalDisks          bool                `hcl:"wipe_additional_disks,optional"`
	KubeAPIServerExtraFlags      []string
}

const (
	// Name represents Bare Metal platform name as it should be referenced in function calls and configuration.
	Name = "bare-metal"
)

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
		AssetDir:             c.AssetDir,
		ExpectedNodes:        len(c.ControllerMacs) + len(c.WorkerMacs),
		ControlplaneCharts:   platform.CommonControlPlaneCharts(!c.DisableSelfHostedKubelet),
		Deployments:          platform.CommonDeployments(len(c.ControllerMacs)),
		DaemonSets:           platform.CommonDaemonSets(len(c.ControllerMacs), c.DisableSelfHostedKubelet),
		ControllerModuleName: fmt.Sprintf("%s-%s", Name, c.ClusterName),
	}
}

func NewConfig() *config {
	return &config{
		CachedInstall:                "false",
		OSChannel:                    "stable",
		OSVersion:                    "current",
		NetworkMTU:                   platform.NetworkMTU,
		ConntrackMaxPerCore:          platform.ConntrackMaxPerCore,
		DownloadProtocol:             "https",
		NetworkIPAutodetectionMethod: "first-found",
	}
}

func (c *config) Apply(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Apply([]string{terraform.WithParallelism})
}

// ApplyWithoutParallel applies Terraform configuration without parallel execution.
func (c *config) ApplyWithoutParallel(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return fmt.Errorf("initializing Terraform configuration: %w", err)
	}

	return ex.Apply([]string{terraform.WithoutParallelism})
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

func createTerraformConfigFile(cfg *config, terraformPath string) error {
	tmplName := "cluster.tf"
	t := template.New(tmplName)
	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		// TODO: Use template.Must().
		return fmt.Errorf("parsing template: %w", err)
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
		return fmt.Errorf("marshaling SSH public keys: %w", err)
	}

	workerDomains, err := json.Marshal(cfg.WorkerDomains)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling worker domains: %w", err)
	}

	workerMacs, err := json.Marshal(cfg.WorkerMacs)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling worker MAC addresses: %w", err)
	}

	workerNames, err := json.Marshal(cfg.WorkerNames)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling worker names: %w", err)
	}

	controllerDomains, err := json.Marshal(cfg.ControllerDomains)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling controller domains: %w", err)
	}

	controllerMacs, err := json.Marshal(cfg.ControllerMacs)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling controller MAC addresses: %w", err)
	}

	controllerNames, err := json.Marshal(cfg.ControllerNames)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling controller names: %w", err)
	}

	terraformCfg := struct {
		CachedInstall                string
		ClusterName                  string
		ControllerDomains            string
		ControllerMacs               string
		ControllerNames              string
		K8sDomainName                string
		MatchboxClientCert           string
		MatchboxClientKey            string
		MatchboxCA                   string
		MatchboxEndpoint             string
		MatchboxHTTPEndpoint         string
		NetworkMTU                   int
		OSChannel                    string
		OSVersion                    string
		SSHPublicKeys                string
		WorkerNames                  string
		WorkerMacs                   string
		WorkerDomains                string
		DisableSelfHostedKubelet     bool
		KubeAPIServerExtraFlags      []string
		Labels                       Labels
		NodeSpecificLabels           map[string]Labels
		EncryptPodTraffic            bool
		IgnoreX509CNCheck            bool
		CertsValidityPeriodHours     int
		ConntrackMaxPerCore          int
		InstallDisk                  string
		InstallToSmallestDisk        bool
		KernelArgs                   []string
		KernelConsole                []string
		DownloadProtocol             string
		NetworkIPAutodetectionMethod string
		CLCSnippets                  map[string][]string
		InstallerCLCSnippets         map[string][]string
		WipeAdditionalDisks          bool
	}{
		CachedInstall:                cfg.CachedInstall,
		ClusterName:                  cfg.ClusterName,
		ControllerDomains:            string(controllerDomains),
		ControllerMacs:               string(controllerMacs),
		ControllerNames:              string(controllerNames),
		K8sDomainName:                cfg.K8sDomainName,
		MatchboxCA:                   cfg.MatchboxCAPath,
		MatchboxClientCert:           cfg.MatchboxClientCertPath,
		MatchboxClientKey:            cfg.MatchboxClientKeyPath,
		MatchboxEndpoint:             cfg.MatchboxEndpoint,
		MatchboxHTTPEndpoint:         cfg.MatchboxHTTPEndpoint,
		NetworkMTU:                   cfg.NetworkMTU,
		OSChannel:                    cfg.OSChannel,
		OSVersion:                    cfg.OSVersion,
		SSHPublicKeys:                string(keyListBytes),
		WorkerNames:                  string(workerNames),
		WorkerMacs:                   string(workerMacs),
		WorkerDomains:                string(workerDomains),
		DisableSelfHostedKubelet:     cfg.DisableSelfHostedKubelet,
		KubeAPIServerExtraFlags:      cfg.KubeAPIServerExtraFlags,
		Labels:                       cfg.Labels,
		NodeSpecificLabels:           cfg.NodeSpecificLabels,
		EncryptPodTraffic:            cfg.EncryptPodTraffic,
		IgnoreX509CNCheck:            cfg.IgnoreX509CNCheck,
		CertsValidityPeriodHours:     cfg.CertsValidityPeriodHours,
		ConntrackMaxPerCore:          cfg.ConntrackMaxPerCore,
		InstallDisk:                  cfg.InstallDisk,
		InstallToSmallestDisk:        cfg.InstallToSmallestDisk,
		KernelArgs:                   cfg.KernelArgs,
		KernelConsole:                cfg.KernelConsole,
		DownloadProtocol:             cfg.DownloadProtocol,
		NetworkIPAutodetectionMethod: cfg.NetworkIPAutodetectionMethod,
		CLCSnippets:                  cfg.CLCSnippets,
		InstallerCLCSnippets:         cfg.InstallerCLCSnippets,
		WipeAdditionalDisks:          cfg.WipeAdditionalDisks,
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

// checkValidConfig validates cluster configuration.
func (c *config) checkValidConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if c.InstallToSmallestDisk && c.InstallDisk != "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "`install_disk` and `install_to_smallest_disk` are mutually exclusive",
			Detail:   "Provide either `install_disk` or `install_to_smallest_disk` or none, but not both",
		})
	}

	if c.DownloadProtocol != "http" && c.DownloadProtocol != "https" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid value for `download_protocol`",
			Detail:   fmt.Sprintf("expected 'http' or 'https', got: %q", c.DownloadProtocol),
		})
	}

	if c.ConntrackMaxPerCore < 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "conntrack_max_per_core can't be negative value",
			Detail:   fmt.Sprintf("'conntrack_max_per_core' value is %d", c.ConntrackMaxPerCore),
		})
	}

	if c.OIDC != nil {
		_, diags := c.OIDC.ToKubeAPIServerFlags(c.K8sDomainName)
		diagnostics = append(diagnostics, diags...)
	}

	for key, list := range c.CLCSnippets {
		if key == "" || len(list) == 0 {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "key/value for clc_snippets map can't be empty",
				Detail:   fmt.Sprintf("either key or value for clc_snippets is empty: %q : %v", key, list),
			})
		}

		for _, data := range list {
			if data == "" {
				diagnostics = append(diagnostics, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Values list for clc_snippets cannot contain an empty element",
					Detail:   fmt.Sprintf("found empty element in the key value pair: %q : %v", key, list),
				})
			}
		}
	}

	return diagnostics
}
