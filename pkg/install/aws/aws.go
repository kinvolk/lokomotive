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
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/platform/util"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

type config struct {
	AssetDir                 string            `hcl:"asset_dir"`
	ClusterName              string            `hcl:"cluster_name"`
	Tags                     map[string]string `hcl:"tags,optional"`
	OSName                   string            `hcl:"os_name,optional"`
	OSChannel                string            `hcl:"os_channel,optional"`
	OSVersion                string            `hcl:"os_version,optional"`
	DNSZone                  string            `hcl:"dns_zone"`
	DNSZoneID                string            `hcl:"dns_zone_id"`
	SSHPubKeys               []string          `hcl:"ssh_pubkeys"`
	CredsPath                string            `hcl:"creds_path,optional"`
	ControllerCount          int               `hcl:"controller_count,optional"`
	ControllerType           string            `hcl:"controller_type,optional"`
	WorkerCount              int               `hcl:"worker_count,optional"`
	WorkerType               string            `hcl:"worker_type,optional"`
	ControllerCLCSnippets    []string          `hcl:"controller_clc_snippets,optional"`
	WorkerCLCSnippets        []string          `hcl:"worker_clc_snippets,optional"`
	Region                   string            `hcl:"region,optional"`
	EnableAggregation        bool              `hcl:"enable_aggregation,optional"`
	DiskSize                 int               `hcl:"disk_size,optional"`
	DiskType                 string            `hcl:"disk_type,optional"`
	DiskIOPS                 int               `hcl:"disk_iops,optional"`
	WorkerPrice              string            `hcl:"worker_price,optional"`
	WorkerTargetGroups       []string          `hcl:"worker_target_groups,optional"`
	Networking               string            `hcl:"networking,optional"`
	NetworkMTU               int               `hcl:"network_mtu,optional"`
	HostCIDR                 string            `hcl:"host_cidr,optional"`
	PodCIDR                  string            `hcl:"pod_cidr,optional"`
	ServiceCIDR              string            `hcl:"service_cidr,optional"`
	ClusterDomainSuffix      string            `hcl:"cluster_domain_suffix,optional"`
	EnableReporting          bool              `hcl:"enable_reporting,optional"`
	CertsValidityPeriodHours int               `hcl:"certs_validity_period_hours,optional"`
}

// init registers aws as a platform
func init() {
	platform.Register("aws", NewConfig())
}

func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func NewConfig() *config {
	return &config{
		OSName:          "flatcar",
		OSChannel:       "stable",
		OSVersion:       "current",
		ControllerCount: 1,
		ControllerType:  "t3.small",
		WorkerCount:     2,
		WorkerType:      "t3.small",
		Region:          "eu-central-1",
		// Initialize the string slices to make sure they are
		// rendered as `[]` when no snippets are given and not
		// `null`, as the latter would lead to a terraform error
		ControllerCLCSnippets: make([]string, 0),
		WorkerCLCSnippets:     make([]string, 0),
		WorkerTargetGroups:    make([]string, 0),
		EnableAggregation:     true,
	}
}

// GetAssetDir returns asset directory path
func (c *config) GetAssetDir() string {
	return c.AssetDir
}

func (c *config) Install(ex *terraform.Executor) error {
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

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)

	return createTerraformConfigFile(c, terraformRootDir)
}

func createTerraformConfigFile(cfg *config, terraformRootDir string) error {
	tmplName := "cluster.tf"
	t := template.New(tmplName)
	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	path := filepath.Join(terraformRootDir, tmplName)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", path)
	}
	defer f.Close()

	keyListBytes, err := json.Marshal(cfg.SSHPubKeys)
	if err != nil {
		return errors.Wrap(err, "failed to marshal SSH public keys")
	}

	controllerCLCSnippetsBytes, err := json.Marshal(cfg.ControllerCLCSnippets)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	workerCLCSnippetsBytes, err := json.Marshal(cfg.WorkerCLCSnippets)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	workerTargetGroupsBytes, err := json.Marshal(cfg.WorkerTargetGroups)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal CLC snippets")
	}

	util.AppendTags(&cfg.Tags)
	tags, err := json.Marshal(cfg.Tags)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal tags")
	}

	terraformCfg := struct {
		Config                config
		Tags                  string
		SSHPublicKeys         string
		ControllerCLCSnippets string
		WorkerCLCSnippets     string
		WorkerTargetGroups    string
	}{
		Config:                *cfg,
		Tags:                  string(tags),
		SSHPublicKeys:         string(keyListBytes),
		ControllerCLCSnippets: string(controllerCLCSnippetsBytes),
		WorkerCLCSnippets:     string(workerCLCSnippetsBytes),
		WorkerTargetGroups:    string(workerTargetGroupsBytes),
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}

func (c *config) GetExpectedNodes() int {
	return c.ControllerCount + c.WorkerCount
}
