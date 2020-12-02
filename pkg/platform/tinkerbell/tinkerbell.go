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

package tinkerbell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/mitchellh/go-homedir"

	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

// Config represents Tinkerbell platform configuration.
type Config struct { //nolint:maligned
	AssetDir              string   `hcl:"asset_dir"`
	Name                  string   `hcl:"name"`
	DNSZone               string   `hcl:"dns_zone"`
	SSHPublicKeys         []string `hcl:"ssh_public_keys"`
	ControllerIPAddresses []string `hcl:"controller_ip_addresses"`

	ControllerCLCSnippets           []string `hcl:"controller_clc_snippets,optional"`
	ControllerFlatcarInstallBaseURL string   `hcl:"controller_flatcar_install_base_url,optional"`

	OSChannel string `hcl:"os_channel,optional"`
	OSVersion string `hcl:"os_version,optional"`

	// If true, Tinkerbell server will be created, then used to setup Lokomotive cluster.
	Sandbox *Sandbox `hcl:"experimental_sandbox,block"`

	// Generic options.
	EnableAggregation        bool   `hcl:"enable_aggregation,optional"`
	EnableReporting          bool   `hcl:"enable_reporting,optional"`
	PodCIDR                  string `hcl:"pod_cidr,optional"`
	ServiceCIDR              string `hcl:"service_cidr,optional"`
	ClusterDomainSuffix      string `hcl:"cluster_domain_suffix,optional"`
	CertsValidityPeriodHours int    `hcl:"certs_validity_period_hours,optional"`
	NetworkMTU               int    `hcl:"network_mtu,optional"`
	DisableSelfHostedKubelet bool   `hcl:"disable_self_hosted_kubelet,optional"`
	ConntrackMaxPerCore      int    `hcl:"conntrack_max_per_core,optional"`

	WorkerPools []WorkerPool `hcl:"worker_pool,block"`
}

// Sandbox represents Tinkerbell sandbox setup configuration.
type Sandbox struct {
	HostsCIDR        string `hcl:"hosts_cidr"`
	FlatcarImagePath string `hcl:"flatcar_image_path"`
	PoolPath         string `hcl:"pool_path"`
}

// WorkerPool represents Tinkerbell worker pool configuration.
type WorkerPool struct {
	PoolName string `hcl:"name,label"`

	IPAddresses   []string `hcl:"ip_addresses"`
	SSHPublicKeys []string `hcl:"ssh_public_keys,optional"`

	OSChannel             string   `hcl:"os_channel,optional"`
	OSVersion             string   `hcl:"os_version,optional"`
	FlatcarInstallBaseURL string   `hcl:"flatcar_install_base_url,optional"`
	CLCSnippets           []string `hcl:"clc_snippets,optional"`

	// Generic options.
	Labels map[string]string `hcl:"labels,optional"`
	Taints map[string]string `hcl:"taints,optional"`
}

// Name returns worker pool name.
func (w *WorkerPool) Name() string {
	return w.PoolName
}

// init registers tinkerbell as a platform.
func init() { //nolint:gochecknoinits
	platform.Register("tinkerbell", NewConfig())
}

// LoadConfig loads platform configuration using given HCL structs.
func (c *Config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		emptyConfig := hcl.EmptyBody()
		configBody = &emptyConfig
	}

	if diags := gohcl.DecodeBody(*configBody, evalContext, c); len(diags) != 0 {
		return diags
	}

	for i, k := range c.SSHPublicKeys {
		c.SSHPublicKeys[i] = strings.TrimSpace(k)
	}

	for _, p := range c.WorkerPools {
		for i, k := range p.SSHPublicKeys {
			p.SSHPublicKeys[i] = strings.TrimSpace(k)
		}
	}

	return c.Validate()
}

// NewConfig returns Tinkerbell default configuration.
func NewConfig() *Config {
	return &Config{
		EnableAggregation:   true,
		ConntrackMaxPerCore: platform.ConntrackMaxPerCore,
	}
}

// Meta is part of Platform interface and returns common information about the platform configuration.
func (c *Config) Meta() platform.Meta {
	nodes := len(c.ControllerIPAddresses)
	for _, workerpool := range c.WorkerPools {
		nodes += len(workerpool.IPAddresses)
	}

	return platform.Meta{
		AssetDir:           c.AssetDir,
		ExpectedNodes:      nodes,
		ControlplaneCharts: platform.CommonControlPlaneCharts(!c.DisableSelfHostedKubelet),
	}
}

// Initialize unpacks control plane Helm charts into assets directory and creates
// Terraform configuration file.
func (c *Config) Initialize(ex *terraform.Executor) error {
	assetDir, err := homedir.Expand(c.AssetDir)
	if err != nil {
		return err
	}

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)

	return c.createTerraformConfigFile(terraformRootDir)
}

// Apply applies Terraform configuration.
func (c *Config) Apply(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Apply()
}

// Destroy destroys Terraform managed resources.
func (c *Config) Destroy(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Destroy()
}

func (c *Config) createTerraformConfigFile(terraformPath string) error {
	tmplName := "cluster.tf"

	t := template.Must(template.New("cluster.tf").Parse(terraformConfigTmpl))

	path := filepath.Join(terraformPath, tmplName)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", path, err)
	}

	if err := t.Execute(f, c); err != nil {
		return fmt.Errorf("failed to write template to file %q: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed closing file %q: %w", path, err)
	}

	return nil
}

// Validate validates cluster configuration.
func (c *Config) Validate() hcl.Diagnostics {
	var d hcl.Diagnostics

	// Convert Tinkerbell worker pool to generic workerpool collection.
	x := []platform.WorkerPool{}
	for i, pool := range c.WorkerPools {
		x = append(x, &c.WorkerPools[i])

		if pool.PoolName == "" {
			d = append(d, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Worker pools name can't be empty",
				Detail:   fmt.Sprintf("Worker pool %d name is empty", i),
			})
		}

		if len(pool.IPAddresses) == 0 {
			d = append(d, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Worker pool must have at least one IP address configured",
				Detail:   fmt.Sprintf("Worker pool %q IP addresses list is empty", pool.PoolName),
			})
		}
	}

	if c.ConntrackMaxPerCore < 0 {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "conntrack_max_per_core can't be negative value",
			Detail:   fmt.Sprintf("'conntrack_max_per_core' value is %d", c.ConntrackMaxPerCore),
		})
	}

	d = append(d, platform.WorkerPoolNamesUnique(x)...)
	d = append(d, c.validateRequiredFields()...)

	return d
}

func (c *Config) validateRequiredFields() hcl.Diagnostics {
	var d hcl.Diagnostics

	if c.AssetDir == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Asset dir can't be empty",
		})
	}

	if c.Name == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Cluster name can't be empty",
		})
	}

	if c.DNSZone == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "DNS zone can't be empty",
		})
	}

	if len(c.SSHPublicKeys) == 0 {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Controllers must have at least one SSH key configured",
		})
	}

	if c.Sandbox != nil && c.Sandbox.FlatcarImagePath == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Tinkerbell sandbox must have Flatcar image path configured",
		})
	}

	if c.Sandbox != nil && c.Sandbox.HostsCIDR == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Tinkerbell sandbox must have hosts CIDR configured",
		})
	}

	if c.Sandbox != nil && c.Sandbox.PoolPath == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Tinkerbell sandbox must have pool path configured",
		})
	}

	return d
}
