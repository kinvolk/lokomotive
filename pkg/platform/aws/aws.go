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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/mitchellh/go-homedir"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/oidc"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

type workerPool struct {
	Name         string            `hcl:"pool_name,label"`
	Count        int               `hcl:"count"`
	SSHPubKeys   []string          `hcl:"ssh_pubkeys"`
	InstanceType string            `hcl:"instance_type,optional"`
	OSChannel    string            `hcl:"os_channel,optional"`
	OSVersion    string            `hcl:"os_version,optional"`
	Labels       string            `hcl:"labels,optional"`
	Taints       string            `hcl:"taints,optional"`
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

type config struct {
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
	EnableTLSBootstrap       bool              `hcl:"enable_tls_bootstrap,optional"`
	EncryptPodTraffic        bool              `hcl:"encrypt_pod_traffic,optional"`
	KubeAPIServerExtraFlags  []string
}

// init registers aws as a platform
func init() {
	platform.Register("aws", NewConfig())
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	if diags := gohcl.DecodeBody(*configBody, evalContext, c); len(diags) != 0 {
		return diags
	}

	return c.checkValidConfig()
}

func NewConfig() *config {
	return &config{
		Region:             "eu-central-1",
		EnableAggregation:  true,
		EnableTLSBootstrap: true,
		NetworkMTU:         platform.NetworkMTU,
	}
}

func (c *config) clusterDomain() string {
	return fmt.Sprintf("%s.%s", c.ClusterName, c.DNSZone)
}

// Meta is part of Platform interface and returns common information about the platform configuration.
func (c *config) Meta() platform.Meta {
	nodes := c.ControllerCount
	for _, workerpool := range c.WorkerPools {
		nodes += workerpool.Count
	}

	return platform.Meta{
		AssetDir:      c.AssetDir,
		ExpectedNodes: nodes,
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

func createTerraformConfigFile(cfg *config, terraformRootDir string) error {
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
	defer f.Close()

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
	var diagnostics hcl.Diagnostics

	diagnostics = append(diagnostics, c.checkNotEmptyWorkers()...)
	diagnostics = append(diagnostics, c.checkWorkerPoolNamesUnique()...)
	diagnostics = append(diagnostics, c.checkNameSizes()...)
	diagnostics = append(diagnostics, c.checkLBPortsUnique()...)

	if c.OIDC != nil {
		_, diags := c.OIDC.ToKubeAPIServerFlags(c.clusterDomain())
		diagnostics = append(diagnostics, diags...)
	}

	return diagnostics
}

// checkLBPortsUnique checks that the lb_http_port and lb_https_port
// flags have different values if the user is using multiple worker
// pools.
func (c *config) checkLBPortsUnique() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	portMap := make(map[int]struct {
		field    string
		poolName string
	})

	for _, wp := range c.WorkerPools {
		httpPort := wp.LBHTTPPort
		if httpPort == 0 {
			httpPort = 80
		}

		httpsPort := wp.LBHTTPSPort
		if httpsPort == 0 {
			httpsPort = 443
		}

		ports := map[int]string{
			httpPort:  "lb_http_port",
			httpsPort: "lb_https_port",
		}

		for p, field := range ports {
			if v, ok := portMap[p]; ok {
				diagnostics = append(diagnostics, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unique ports required",
					Detail:   fmt.Sprintf("'worker_pool.%s.%s' collides with 'worker_pool.%s.%s'", wp.Name, field, v.poolName, v.field), //nolint:lll
				})
			}

			portMap[p] = struct {
				field    string
				poolName string
			}{field, wp.Name}
		}
	}

	return diagnostics
}

// checkNameSizes checks the size of names since AWS has a limit of 32
// characters on resources.
func (c *config) checkNameSizes() hcl.Diagnostics {
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
			Summary:  "Worker pools name should be unique",
			Detail:   fmt.Sprintf("Worker pool '%v' is duplicated", w.Name),
		})
	}

	return diagnostics
}
