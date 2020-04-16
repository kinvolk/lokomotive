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
	"net"

	"github.com/hashicorp/hcl/v2"
	configpkg "github.com/kinvolk/lokomotive/pkg/cluster/config"
	"github.com/kinvolk/lokomotive/pkg/platform/util"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	utilpkg "github.com/kinvolk/lokomotive/pkg/util"
	"github.com/pkg/errors"
)

const defaultDiskSize = 40

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

type flatcar struct {
	OSName string `hcl:"os_name,optional"`
}

type network struct {
	HostCIDR string `hcl:"host_cidr"`
}

type controller struct {
	Type        string            `hcl:"type,optional"`
	CLCSnippets []string          `hcl:"clc_snippets,optional"`
	Tags        map[string]string `hcl:"tags,optional"`
}

type disk struct {
	Size int    `hcl:"size,optional"`
	Type string `hcl:"type,optional"`
	IOPS int    `hcl:"iops,optional"`
}

type config struct {
	Metadata        *configpkg.Metadata
	DNSZone         string       `hcl:"dns_zone"`
	DNSZoneID       string       `hcl:"dns_zone_id"`
	ExposeNodePorts bool         `hcl:"expose_nodeports,optional"`
	CredsPath       string       `hcl:"creds_path,optional"`
	Region          string       `hcl:"region,optional"`
	WorkerPools     []workerPool `hcl:"worker_pool,block"`
	Disk            *disk        `hcl:"disk,block"`
	Flatcar         *flatcar     `hcl:"flatcar,block"`
	Network         *network     `hcl:"network,block"`
	Controller      *controller  `hcl:"controller,block"`
}

// init registers packet as a platform
//nolint:gochecknoinits
func init() {
	configpkg.Register("aws", newConfig())
}

func newConfig() *config {
	return &config{
		Flatcar: &flatcar{
			OSName: "flatcar",
		},
		Network: &network{
			HostCIDR: "10.0.0.0/16",
		},
		Controller: &controller{
			Type: "t3.medium",
		},
		Disk: &disk{
			Size: defaultDiskSize,
			IOPS: 0,
			Type: "gp2",
		},
		Region: "eu-central-1",
	}
}

func (c *config) Apply(ex *terraform.Executor) error {

	return ex.Apply()
}

func (c *config) Destroy(ex *terraform.Executor) error {

	return ex.Destroy()
}

func (c *config) SetMetadata(metadata *configpkg.Metadata) {
	c.Metadata = metadata
}

func (c *config) Render(cfg *configpkg.LokomotiveConfig) (string, error) {
	workerpoolCfgList := []map[string]string{}
	keyListBytes, err := json.Marshal(cfg.Controller.SSHPubKeys)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal SSH public keys")
	}

	controllerCLCSnippetsBytes, err := json.Marshal(c.Controller.CLCSnippets)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	util.AppendTags(&c.Controller.Tags)

	tags, err := json.Marshal(c.Controller.Tags)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal tags")
	}

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

	terraformCfg := struct {
		LokomotiveConfig      *configpkg.LokomotiveConfig
		AWSConfig             *config
		ControllerTags        string
		SSHPubKeys            string
		ControllerCLCSnippets string
		WorkerPoolsList       []map[string]string
	}{
		LokomotiveConfig:      cfg,
		AWSConfig:             c,
		ControllerTags:        string(tags),
		SSHPubKeys:            string(keyListBytes),
		ControllerCLCSnippets: string(controllerCLCSnippetsBytes),
		WorkerPoolsList:       workerpoolCfgList,
	}

	return utilpkg.RenderTemplate(terraformConfigTmpl, terraformCfg)
}

func (c *config) Validate() hcl.Diagnostics {
	// check all configuration
	// whether its valid or not
	return c.checkValidConfig()
}

func (c *config) GetExpectedNodes(cfg *configpkg.LokomotiveConfig) int {
	nodes := cfg.Controller.Count
	for _, workerpool := range c.WorkerPools {
		nodes += workerpool.Count
	}

	return nodes
}

// checkValidConfig validates cluster configuration.
func (c *config) checkValidConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	diagnostics = append(diagnostics, c.checkNotEmptyWorkers()...)
	diagnostics = append(diagnostics, c.checkWorkerPoolNamesUnique()...)
	diagnostics = append(diagnostics, c.checkNameSizes()...)
	diagnostics = append(diagnostics, c.checkAWSConfig()...)
	diagnostics = append(diagnostics, c.checkFlatcarConfig()...)
	diagnostics = append(diagnostics, c.checkNetworkConfig()...)
	diagnostics = append(diagnostics, c.checkControllerConfig()...)

	return diagnostics
}

func (c *config) checkAWSConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if c.Region == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected `region` to be non-empty"),
		})
	}

	if c.DNSZone == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected `dns_zone` to be non-empty"),
		})
	}

	if c.DNSZoneID == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected `dns_zone_id` to be non-empty"),
		})
	}

	return diagnostics
}
func (c *config) checkNetworkConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if c.Network.HostCIDR == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Detail:   "required field `host_cidr` is missing",
		})
	}

	if err := validCIDR(c.Network.HostCIDR); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid 'host_cidr' `%s`: %v", c.Network.HostCIDR, err),
		})
	}

	return diagnostics
}

func validCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)

	return err
}
func (c *config) checkFlatcarConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if c.Flatcar.OSName != "flatcar" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected 'flatcar', got: %s", c.Flatcar.OSName),
		})
	}

	return diagnostics
}

func (c *config) checkControllerConfig() hcl.Diagnostics {
	//TODO: Get a list of valid packet machine types and validate
	var diagnostics hcl.Diagnostics
	if c.Controller.Type == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "`type` cannot be empty",
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

	if len(c.Metadata.ClusterName) > maxNameLen {
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
			Summary:  "one or more worker pool blocks required",
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
			Summary:  fmt.Sprintf("worker pool '%v' is not unique", w.Name),
		})
	}

	return diagnostics
}
