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

package kvmlibvirt

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

type workerPool struct {
	Name          string   `hcl:"pool_name,label"`
	Count         int      `hcl:"count"`
	VirtualCPUs   int      `hcl:"virtual_cpus,optional"`
	VirtualMemory int      `hcl:"virtual_memory,optional"`
	CLCSnippets   []string `hcl:"clc_snippets,optional"`
	Labels        string   `hcl:"labels,optional"`
}

type config struct {
	AssetDir                     string       `hcl:"asset_dir"`
	ClusterName                  string       `hcl:"cluster_name"`
	ControllerCount              int          `hcl:"controller_count,optional"`
	MachineDomain                string       `hcl:"machine_domain"`
	OSImage                      string       `hcl:"os_image"`
	NodeIpPool                   string       `hcl:"node_ip_pool,optional"`
	SSHPubKeys                   []string     `hcl:"ssh_pubkeys"`
	WorkerPools                  []workerPool `hcl:"worker_pool,block"`
	DisableSelfHostedKubelet     bool         `hcl:"disable_self_hosted_kubelet,optional"`
	KubeAPIServerExtraFlags      []string     `hcl:"kube_apiserver_extra_flags,optional"`
	ControllerVirtualCPUs        int          `hcl:"controller_virtual_cpus,optional"`
	ControllerVirtualMemory      int          `hcl:"controller_virtual_memory,optional"`
	ControllerCLCSnippets        []string     `hcl:"controller_clc_snippets,optional"`
	NetworkMTU                   int          `hcl:"network_mtu,optional"`
	NetworkIpAutodetectionMethod string       `hcl:"network_ip_autodetection_method,optional"`
	PodCidr                      string       `hcl:"pod_cidr,optional"`
	ServiceCidr                  string       `hcl:"service_cidr,optional"`
	ClusterDomainSuffix          string       `hcl:"cluster_domain_suffix,optional"`
	EnableReporting              bool         `hcl:"enable_reporting,optional"`
	EnableAggregation            bool         `hcl:"enable_aggregation,optional"`
	CertsValidityPeriodHours     int          `hcl:"certs_validity_period_hours,optional"`
}

func init() {
	platform.Register("kvm-libvirt", NewConfig())
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
	nodes := c.ControllerCount
	for _, workerpool := range c.WorkerPools {
		nodes += workerpool.Count
	}
	return platform.Meta{
		AssetDir:      c.AssetDir,
		ExpectedNodes: nodes,
	}
}

func NewConfig() *config {
	return &config{}
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
	tmplName := "cluster.tf"
	t := template.New(tmplName).Funcs(template.FuncMap{"StringsJoin": strings.Join})
	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	path := filepath.Join(terraformRootDir, tmplName)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", path, err)
	}
	defer f.Close()

	terraformCfg := struct {
		Config config
	}{
		Config: *cfg,
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
