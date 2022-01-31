// Copyright 2022 The Lokomotive Authors
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

// Package azure provides the implenentation of the Platform interface
// for Azure cloud provider.
package azure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/mitchellh/go-homedir"

	"github.com/kinvolk/lokomotive/pkg/dns"
	"github.com/kinvolk/lokomotive/pkg/oidc"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

// workerPool defines "worker_pool" block.
type workerPool struct {
	Name             string            `hcl:"pool_name,label"`
	SSHPubKeys       []string          `hcl:"ssh_pubkeys"`
	Count            int               `hcl:"count"`
	CPUManagerPolicy string            `hcl:"cpu_manager_policy,optional"`
	VMType           string            `hcl:"vm_type,optional"`
	OSImage          string            `hcl:"os_image,optional"`
	CLCSnippets      []string          `hcl:"clc_snippets,optional"`
	Priority         string            `hcl:"priority,optional"`
	Tags             map[string]string `hcl:"tags,optional"`
	Labels           map[string]string `hcl:"labels,optional"`
	Taints           map[string]string `hcl:"taints,optional"`
}

// config defines "cluster" block for Azure.
type config struct { //nolint:maligned
	AssetDir              string            `hcl:"asset_dir"`
	ClusterName           string            `hcl:"cluster_name"`
	ControllerType        string            `hcl:"controller_type,optional"`
	ControllerCLCSnippets []string          `hcl:"controller_clc_snippets,optional"`
	WorkerType            string            `hcl:"worker_type,optional"`
	SSHPubKeys            []string          `hcl:"ssh_pubkeys"`
	Tags                  map[string]string `hcl:"tags,optional"`
	ControllerCount       int               `hcl:"controller_count"`
	DNS                   dns.Config        `hcl:"dns,block"`
	Region                string            `hcl:"region,optional"`
	EnableAggregation     bool              `hcl:"enable_aggregation,optional"`
	EnableReporting       bool              `hcl:"enable_reporting,optional"`
	OSImage               string            `hcl:"os_image,optional"`
	ClusterDomainSuffix   string            `hcl:"cluster_domain_suffix,optional"`
	// CustomImageResourceGroupName string            `hcl:"custom_image_resource_group_name,optional"`
	// CustomImageName              string            `hcl:"custom_image_name,optional"`
	EnableNodeLocalDNS       bool         `hcl:"enable_node_local_dns,optional"`
	DisableSelfHostedKubelet bool         `hcl:"disable_self_hosted_kubelet,optional"`
	OIDC                     *oidc.Config `hcl:"oidc,block"`
	EnableTLSBootstrap       bool         `hcl:"enable_tls_bootstrap,optional"`
	EncryptPodTraffic        bool         `hcl:"encrypt_pod_traffic,optional"`
	PodCIDR                  string       `hcl:"pod_cidr,optional"`
	ServiceCIDR              string       `hcl:"service_cidr,optional"`
	CertsValidityPeriodHours int          `hcl:"certs_validity_period_hours,optional"`
	ConntrackMaxPerCore      int          `hcl:"conntrack_max_per_core,optional"`
	WorkerPools              []workerPool `hcl:"worker_pool,block"`

	KubernetesVersion string

	// Not exposed to the user
	KubeAPIServerExtraFlags []string
}

const (
	// Name represents azure platform name as it should be referenced in function calls and configuration.
	Name = "azure"

	kubernetesVersion = "1.21.2"
)

// NewConfig returns new Azure platform configuration with default values set.
//
//nolint:golint
func NewConfig() *config {
	return &config{
		Region:              "West Europe",
		KubernetesVersion:   kubernetesVersion,
		ConntrackMaxPerCore: platform.ConntrackMaxPerCore,
	}
}

// LoadConfig loads configuration values into the config struct from given HCL configuration.
func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	if d := gohcl.DecodeBody(*configBody, evalContext, c); len(d) != 0 {
		return d
	}

	return c.checkValidConfig()
}

func (c *config) clusterDomain() string {
	return fmt.Sprintf("%s.%s", c.ClusterName, c.DNS.Zone)
}

// Meta is part of Platform interface and returns common information about the platform configuration.
func (c *config) Meta() platform.Meta {
	nodes := c.ControllerCount
	for _, workerpool := range c.WorkerPools {
		nodes += workerpool.Count
	}

	charts := platform.CommonControlPlaneCharts(platform.ControlPlanCharts{
		Kubelet:      !c.DisableSelfHostedKubelet,
		NodeLocalDNS: c.EnableNodeLocalDNS,
	})

	return platform.Meta{
		AssetDir:             c.AssetDir,
		ExpectedNodes:        nodes,
		ControlplaneCharts:   charts,
		ControllerModuleName: fmt.Sprintf("%s-%s", Name, c.ClusterName),
		Deployments:          platform.CommonDeployments(c.ControllerCount),
		DaemonSets:           platform.CommonDaemonSets(c.ControllerCount, c.DisableSelfHostedKubelet),
	}
}

// Apply creates Azure infrastructure via Terraform.
func (c *config) Apply(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return c.terraformSmartApply(ex, c.DNS, []string{terraform.WithParallelism})
}

// ApplyWithoutParallel applies Terraform configuration without parallel execution.
func (c *config) ApplyWithoutParallel(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return fmt.Errorf("initializing Terraform configuration: %w", err)
	}

	return c.terraformSmartApply(ex, c.DNS, []string{terraform.WithParallelism})
}

// Destroy destroys Azure infrastructure via Terraform.
func (c *config) Destroy(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Destroy()
}

// Initialize creates Terrafrom files required for Azure.
func (c *config) Initialize(ex *terraform.Executor) error {
	assetDir, err := homedir.Expand(c.AssetDir)
	if err != nil {
		return err
	}

	if err := c.DNS.Validate(); err != nil {
		return fmt.Errorf("parsing DNS configuration: %w", err)
	}

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)

	return createTerraformConfigFile(c, terraformRootDir)
}

// terraformSmartApply applies cluster configuration.
func (c *config) terraformSmartApply(ex *terraform.Executor, dc dns.Config, extraArgs []string) error {
	// If the provider isn't manual, apply everything in a single step.
	if dc.Provider != dns.Manual {
		return ex.Apply(extraArgs)
	}

	steps := []terraform.ExecutionStep{
		// We need the controllers' IP addresses before we can apply the 'dns' module.
		{
			Description: "create controllers",
			Args: append([]string{
				"apply",
				"-auto-approve",
				fmt.Sprintf("-target=module.azure-%s.azurerm_linux_virtual_machine.controllers", c.ClusterName),
			}, extraArgs...),
		},
		{
			Description: "construct DNS records",
			Args:        append([]string{"apply", "-auto-approve", "-target=module.dns"}, extraArgs...),
		},
		// Run `terraform refresh`. This is required in order to make the outputs from the previous
		// apply operations available.
		// TODO: Likely caused by https://github.com/hashicorp/terraform/issues/23158.
		{
			Description: "refresh Terraform state",
			Args:        []string{"refresh"},
		},
		{
			Description:      "complete infrastructure creation",
			Args:             append([]string{"apply", "-auto-approve"}, extraArgs...),
			PreExecutionHook: c.DNS.ManualConfigPrompt(),
		},
	}

	return ex.Execute(steps...)
}

func createTerraformConfigFile(cfg *config, terraformRootDir string) error { //nolint:funlen
	workerpoolCfgList := []map[string]string{}
	tmplName := "cluster.tf"
	t := template.New(tmplName)

	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		// TODO: Use template.Must().
		return fmt.Errorf("parsing template: %w", err)
	}

	path := filepath.Join(terraformRootDir, tmplName)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", path, err)
	}

	defer f.Close() //nolint:errcheck,gosec

	keyListBytes, err := json.Marshal(cfg.SSHPubKeys)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling SSH public keys: %w", err)
	}

	controllerCLCSnippetsBytes, err := json.Marshal(cfg.ControllerCLCSnippets)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling CLC snippets: %w", err)
	}

	// Configure oidc flags and set it to KubeAPIServerExtraFlags.
	if cfg.OIDC != nil {
		// Skipping the error checking here because its done in checkValidConfig().
		oidcFlags, _ := cfg.OIDC.ToKubeAPIServerFlags(cfg.clusterDomain())
		// TODO: Use append instead of setting the oidcFlags to KubeAPIServerExtraFlags
		// append is not used for now because Initialize is called in cli/cmd/cluster.go
		// and again in Apply which duplicates the values.
		cfg.KubeAPIServerExtraFlags = oidcFlags
	}

	platform.AppendVersionTag(&cfg.Tags)

	tags, err := json.Marshal(cfg.Tags)
	if err != nil {
		// TODO: Render manually instead of marshaling.
		return fmt.Errorf("marshaling tags: %w", err)
	}

	for _, workerpool := range cfg.WorkerPools {
		input := map[string]interface{}{
			"clc_snippets": workerpool.CLCSnippets,
			"ssh_pub_keys": workerpool.SSHPubKeys,
			"tags":         workerpool.Tags,
		}

		output := map[string]string{}

		platform.AppendVersionTag(&workerpool.Tags)

		for k, v := range input {
			bytes, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("marshaling %q for worker pool %q failed: %w", k, workerpool.Name, err)
			}

			output[k] = string(bytes)
		}

		workerpoolCfgList = append(workerpoolCfgList, output)
	}

	terraformCfg := struct {
		Config                config
		Tags                  string
		SSHPublicKeys         string
		ControllerCLCSnippets string
		WorkerCLCSnippets     string
		WorkerTargetGroups    string
		WorkerpoolCfg         []map[string]string
	}{
		Config:                *cfg,
		Tags:                  string(tags),
		SSHPublicKeys:         string(keyListBytes),
		ControllerCLCSnippets: string(controllerCLCSnippetsBytes),
		WorkerpoolCfg:         workerpoolCfgList,
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

// checkValidConfig validates cluster configuration.
func (c *config) checkValidConfig() hcl.Diagnostics {
	var d hcl.Diagnostics

	d = append(d, c.checkNotEmptyWorkers()...)
	d = append(d, c.checkWorkerPoolNamesUnique()...)
	d = append(d, c.checkRequiredFields()...)

	if c.ConntrackMaxPerCore < 0 {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "conntrack_max_per_core can't be negative value",
			Detail:   fmt.Sprintf("'conntrack_max_per_core' value is %d", c.ConntrackMaxPerCore),
		})
	}

	if c.OIDC != nil {
		_, diags := c.OIDC.ToKubeAPIServerFlags(c.clusterDomain())
		d = append(d, diags...)
	}

	return d
}

// checkNotEmptyWorkers checks if the cluster has at least 1 node pool defined.
func (c *config) checkNotEmptyWorkers() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if len(c.WorkerPools) == 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "At least one worker pool must be defined",
			Detail:   "Make sure to define at least one worker pool block in your cluster block",
		})
	}

	return diagnostics
}

// checkWorkerPoolNamesUnique verifies that all worker pool names are unique.
func (c *config) checkWorkerPoolNamesUnique() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	dup := make(map[string]bool)

	for _, w := range c.WorkerPools {
		if !dup[w.Name] {
			dup[w.Name] = true

			continue
		}

		// It is duplicated.
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Worker pool names should be unique",
			Detail:   fmt.Sprintf("Worker pool '%v' is duplicated", w.Name),
		})
	}

	return diagnostics
}

// checkRequiredFields checks if that all required fields are populated in the top level configuration.
func (c *config) checkRequiredFields() hcl.Diagnostics {
	var d hcl.Diagnostics

	f := map[string]string{
		"AssetDir":    c.AssetDir,
		"ClusterName": c.ClusterName,
	}

	for k, v := range f {
		if v == "" {
			d = append(d, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("field %q can't be empty", k),
			})
		}
	}

	return d
}
