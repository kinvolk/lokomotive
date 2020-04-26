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

	"github.com/hashicorp/hcl/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/platform/util"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	utilpkg "github.com/kinvolk/lokomotive/pkg/util"
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
	ClusterDomainSuffix      string            `hcl:"cluster_domain_suffix,optional"`
	EnableReporting          bool              `hcl:"enable_reporting,optional"`
	CertsValidityPeriodHours int               `hcl:"certs_validity_period_hours,optional"`
	WorkerPools              []workerPool      `hcl:"worker_pool,block"`
	// Raw fields that will store the strings after unmarshalling of
	// SSHPubKeys, ControllerCLCSnippets, Tags and unmarshalling of
	// some fields in WorkerPools
	TagsRaw                  string
	SSHPubKeysRaw            string
	ControllerCLCSnippetsRaw string
	WorkerPoolsListRaw       []map[string]string
}

// init registers aws as a platform
func init() {
	platform.Register("aws", NewConfig())
}

func NewConfig() *config {
	return &config{
		OSVersion:                "current",
		OSChannel:                "stable",
		ControllerCount:          1,
		ControllerType:           "t3.small",
		DiskSize:                 40,
		DiskType:                 "gp2",
		NetworkMTU:               1480,
		PodCIDR:                  "10.2.0.0/16",
		ServiceCIDR:              "10.3.0.0/16",
		Region:                   "eu-central-1",
		ClusterDomainSuffix:      "cluster.local",
		CertsValidityPeriodHours: 8760,
		EnableAggregation:        true,
	}
}

// GetAssetDir returns asset directory path
func (c *config) GetAssetDir() string {
	return c.AssetDir
}

func (c *config) setExpandedAssetDir() error {
	assetDir, err := homedir.Expand(c.AssetDir)
	if err != nil {
		return err
	}

	c.AssetDir = assetDir

	return nil
}

func (c *config) Apply(ex *terraform.Executor) error {
	return ex.Apply()
}

func (c *config) Destroy(ex *terraform.Executor) error {

	return ex.Destroy()
}

func (c *config) Render() (string, error) {
	keyListBytes, err := json.Marshal(c.SSHPubKeys)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal SSH public keys")
	}

	controllerCLCSnippetsBytes, err := json.Marshal(c.ControllerCLCSnippets)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	util.AppendTags(&c.Tags)

	tags, err := json.Marshal(c.Tags)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal tags")
	}

	workerpoolCfgList := []map[string]string{}

	for _, workerpool := range c.WorkerPools {
		input := map[string]interface{}{
			"clc_snippets":  workerpool.CLCSnippets,
			"target_groups": workerpool.TargetGroups,
			"ssh_pub_keys":  workerpool.SSHPubKeys,
			"tags":          workerpool.Tags,
		}

		output := map[string]string{}

		util.AppendTags(&workerpool.Tags)

		for k, v := range input {
			bytes, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("marshaling %q for worker pool %q failed: %w", k, workerpool.Name, err)
			}

			output[k] = string(bytes)
		}

		workerpoolCfgList = append(workerpoolCfgList, output)
	}

	c.TagsRaw = string(tags)
	c.SSHPubKeysRaw = string(keyListBytes)
	c.ControllerCLCSnippetsRaw = string(controllerCLCSnippetsBytes)
	c.WorkerPoolsListRaw = workerpoolCfgList

	return utilpkg.RenderTemplate(terraformConfigTmpl, c)
}

func (c *config) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if err := c.setExpandedAssetDir(); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("error expanding 'asset_dir' path: %v", err),
		})
	}

	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.DNSZone, "dns_zone")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.DNSZoneID, "dns_zone_id")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.ClusterDomainSuffix, "cluster_domain_suffix")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.OSVersion, "os_version")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.AssetDir, "asset_dir")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.ClusterName, "cluster_name")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.HostCIDR, "host_cidr")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.ControllerType, "controller_type")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.DiskType, "disk_type")...)

	if c.CertsValidityPeriodHours <= 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("`certs_validity_period_hours` should be more than zero, got: %d", c.CertsValidityPeriodHours),
		})
	}

	if c.ControllerCount < 1 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected 'controller_count' greater than 0, got: %d", c.ControllerCount),
		})
	}

	if len(c.SSHPubKeys) == 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected atleast one public ssh-key in 'ssh_pubkeys', got: 0"),
		})
	}

	if !util.IsFlatcarChannelSupported(c.OSChannel) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unsupported channel '%s'", c.OSChannel),
		})
	}

	if c.NetworkMTU <= 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected 'network_mtu' to be greater than zero, got: %d", c.NetworkMTU),
		})
	}

	if err := util.IsValidCIDR(c.PodCIDR); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid 'pod_cidr': %s", c.PodCIDR),
		})
	}

	if err := util.IsValidCIDR(c.ServiceCIDR); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid 'service_cidr': %s", c.ServiceCIDR),
		})
	}

	if err := util.IsValidCIDR(c.HostCIDR); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid 'host_cidr': %s", c.HostCIDR),
		})
	}

	for _, wp := range c.WorkerPools {
		diagnostics = append(diagnostics, wp.checkValidWorkerPoolConfig()...)
	}

	diagnostics = append(diagnostics, c.checkNotEmptyWorkers()...)

	diagnostics = append(diagnostics, c.checkWorkerPoolNamesUnique()...)

	diagnostics = append(diagnostics, c.checkNameSizes()...)

	return diagnostics
}

func (c *config) GetExpectedNodes() int {
	nodes := c.ControllerCount
	for _, workerpool := range c.WorkerPools {
		nodes += workerpool.Count
	}

	return nodes
}

// checkValidWorkerPoolConfig validates cluster configuration.
func (wp *workerPool) checkValidWorkerPoolConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if wp.Count < 1 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected 'count' in worker_pool '%s' greater than 0, got: %d", wp.Name, wp.Count),
		})
	}

	if wp.OSChannel != "" && !util.IsFlatcarChannelSupported(wp.OSChannel) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unsupported channel '%s'", wp.OSChannel),
		})
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
			Detail:   fmt.Sprintf("Maximum lenth is %d", maxNameLen),
		})
	}

	for _, wp := range c.WorkerPools {
		if len(wp.Name) > maxNameLen {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Worker pool name too long",
				Detail:   fmt.Sprintf("Maximum lenth is %d", maxNameLen),
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
