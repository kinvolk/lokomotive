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

package aws

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

type workerPool struct {
	Name         string            `hcl:"pool_name,label"`
	Count        int               `hcl:"count"`
	SSHPubKeys   []string          `hcl:"ssh_pubkeys"`
	InstanceType string            `hcl:"instance_type,optional"`
	OSChannel    string            `hcl:"os_channel,optional"`
	OSVersion    string            `hcl:"os_version,optional"`
	DiskSize     int               `hcl:"disk_size,optional"`
	DiskType     string            `hcl:"disk_type,optional"`
	DiskIOPS     int               `hcl:"disk_iops,optional"`
	SpotPrice    string            `hcl:"spot_price,optional"`
	TargetGroups []string          `hcl:"target_groups,optional"`
	CLCSnippets  []string          `hcl:"clc_snippets,optional"`
	Tags         map[string]string `hcl:"tags,optional"`
	LBHTTPPort   int               `hcl:"lb_http_port,optional"`
	LBHTTPSPort  int               `hcl:"lb_https_port,optional"`
}

type Config struct {
	AssetDir                 string            `hcl:"asset_dir"`
	ClusterName              string            `hcl:"cluster_name"`
	Tags                     map[string]string `hcl:"tags,optional"`
	OSName                   string            `hcl:"os_name,optional"`
	OSChannel                string            `hcl:"os_channel,optional"`
	OSVersion                string            `hcl:"os_version,optional"`
	DNSZone                  string            `hcl:"dns_zone"`
	DNSZoneID                string            `hcl:"dns_zone_id"`
	ExposeNodePorts          bool              `hcl:"expose_nodeports,optional"`
	SSHPubKeys               []string          `hcl:"ssh_pubkeys"`
	CredsPath                string            `hcl:"creds_path,optional"`
	ControllerCount          int               `hcl:"controller_count,optional"`
	ControllerType           string            `hcl:"controller_type,optional"`
	ControllerCLCSnippets    []string          `hcl:"controller_clc_snippets,optional"`
	Region                   string            `hcl:"region,optional"`
	EnableAggregation        bool              `hcl:"enable_aggregation,optional"`
	DiskSize                 int               `hcl:"disk_size,optional"`
	DiskType                 string            `hcl:"disk_type,optional"`
	DiskIOPS                 int               `hcl:"disk_iops,optional"`
	NetworkMTU               int               `hcl:"network_mtu,optional"`
	HostCIDR                 string            `hcl:"host_cidr,optional"`
	PodCIDR                  string            `hcl:"pod_cidr,optional"`
	ServiceCIDR              string            `hcl:"service_cidr,optional"`
	EnableCSI                bool              `hcl:"enable_csi,optional"`
	ClusterDomainSuffix      string            `hcl:"cluster_domain_suffix,optional"`
	EnableReporting          bool              `hcl:"enable_reporting,optional"`
	CertsValidityPeriodHours int               `hcl:"certs_validity_period_hours,optional"`
	WorkerPools              []workerPool      `hcl:"worker_pool,block"`
	DisableSelfHostedKubelet bool              `hcl:"disable_self_hosted_kubelet,optional"`
	OIDC                     *oidc.Config      `hcl:"oidc,block"`
	KubeAPIServerExtraFlags  []string
}

// NewConfig creates a new Config and returns a pointer to it as well as any HCL diagnostics.
func NewConfig(b *hcl.Body, ctx *hcl.EvalContext) (*Config, hcl.Diagnostics) {
	diags := hcl.Diagnostics{}

	// Create config with default values.
	c := &Config{
		Region:            "eu-central-1",
		EnableAggregation: true,
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
	nodes := c.config.ControllerCount
	for _, workerpool := range c.config.WorkerPools {
		nodes += workerpool.Count
	}

	return nodes
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

func (c *Config) clusterDomain() string {
	return fmt.Sprintf("%s.%s", c.ClusterName, c.DNSZone)
}

// validate validates cluster configuration.
func (c *Config) validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	diagnostics = append(diagnostics, c.checkNotEmptyWorkers()...)
	diagnostics = append(diagnostics, c.checkWorkerPoolNamesUnique()...)
	diagnostics = append(diagnostics, c.checkNameSizes()...)

	if c.OIDC != nil {
		_, diags := c.OIDC.ToKubeAPIServerFlags(c.clusterDomain())
		diagnostics = append(diagnostics, diags...)
	}

	return diagnostics
}

// checkNameSizes checks the size of names since AWS has a limit of 32
// characters on resources.
func (c *Config) checkNameSizes() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	maxAWSResourceName := 32
	maxNameLen := maxAWSResourceName - len("-workers-https") // This is the longest resource suffix.

	if len(c.ClusterName) > maxNameLen {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Cluster name too long",
			Detail:   fmt.Sprintf("Maximum length is %d", maxNameLen),
		})
	}

	for _, wp := range c.WorkerPools {
		if len(wp.Name) > maxNameLen {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Worker pool name too long",
				Detail:   fmt.Sprintf("Maximum length is %d", maxNameLen),
			})
		}
	}

	return diagnostics
}

// checkNotEmptyWorkers checks if the cluster has at least 1 node pool defined.
func (c *Config) checkNotEmptyWorkers() hcl.Diagnostics {
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
func (c *Config) checkWorkerPoolNamesUnique() hcl.Diagnostics {
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
			Summary:  "Worker pools name should be unique",
			Detail:   fmt.Sprintf("Worker pool '%v' is duplicated", w.Name),
		})
	}

	return diagnostics
}

func renderRootModule(conf *Config) (string, error) {
	t, err := template.New("rootModule").Parse(terraformConfigTmpl)
	if err != nil {
		return "", fmt.Errorf("parsing template: %v", err)
	}

	keyListBytes, err := json.Marshal(conf.SSHPubKeys)
	if err != nil {
		return "", fmt.Errorf("marshaling SSH public keys: %v", err)
	}

	controllerCLCSnippetsBytes, err := json.Marshal(conf.ControllerCLCSnippets)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	// Configure OIDC flags and set them to KubeAPIServerExtraFlags.
	if conf.OIDC != nil {
		// Skipping the error checking here because it's done in validate().
		oidcFlags, _ := conf.OIDC.ToKubeAPIServerFlags(conf.clusterDomain())
		//TODO: Use append instead of setting the oidcFlags to KubeAPIServerExtraFlags
		// append is not used for now because Initialize is called in cli/cmd/cluster.go
		// and again in Apply which duplicates the values.
		conf.KubeAPIServerExtraFlags = oidcFlags
	}

	platform.AppendVersionTag(&conf.Tags)

	tags, err := json.Marshal(conf.Tags)
	if err != nil {
		return "", fmt.Errorf("marshaling tags: %v", err)
	}

	workerpoolCfgList := []map[string]string{}
	for _, workerpool := range conf.WorkerPools {
		input := map[string]interface{}{
			"clc_snippets":  workerpool.CLCSnippets,
			"target_groups": workerpool.TargetGroups,
			"ssh_pub_keys":  workerpool.SSHPubKeys,
			"tags":          workerpool.Tags,
		}

		output := map[string]string{}

		platform.AppendVersionTag(&workerpool.Tags)

		for k, v := range input {
			bytes, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("marshaling %q for worker pool %q failed: %w", k, workerpool.Name, err)
			}

			output[k] = string(bytes)
		}

		workerpoolCfgList = append(workerpoolCfgList, output)
	}

	terraformCfg := struct {
		Config                Config
		Tags                  string
		SSHPublicKeys         string
		ControllerCLCSnippets string
		WorkerCLCSnippets     string
		WorkerTargetGroups    string
		WorkerpoolCfg         []map[string]string
	}{
		Config:                *conf,
		Tags:                  string(tags),
		SSHPublicKeys:         string(keyListBytes),
		ControllerCLCSnippets: string(controllerCLCSnippetsBytes),
		WorkerpoolCfg:         workerpoolCfgList,
	}

	var rendered bytes.Buffer
	if err := t.Execute(&rendered, terraformCfg); err != nil {
		return "", fmt.Errorf("rendering template: %v", err)
	}

	return rendered.String(), nil
}
