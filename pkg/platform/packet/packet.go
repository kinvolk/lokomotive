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

package packet

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/pkg/errors"

	configpkg "github.com/kinvolk/lokomotive/pkg/cluster/config"
	"github.com/kinvolk/lokomotive/pkg/dns"
	"github.com/kinvolk/lokomotive/pkg/platform/util"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	utilpkg "github.com/kinvolk/lokomotive/pkg/util"
)

type workerPool struct {
	Name           string `hcl:"pool_name,label"`
	Count          int    `hcl:"count"`
	DisableBGP     bool   `hcl:"disable_bgp,optional"`
	IPXEScriptURL  string `hcl:"ipxe_script_url,optional"`
	OSArch         string `hcl:"os_arch,optional"`
	OSChannel      string `hcl:"os_channel,optional"`
	OSVersion      string `hcl:"os_version,optional"`
	NodeType       string `hcl:"node_type,optional"`
	Labels         string `hcl:"labels,optional"`
	Taints         string `hcl:"taints,optional"`
	SetupRaid      bool   `hcl:"setup_raid,optional"`
	SetupRaidHDD   bool   `hcl:"setup_raid_hdd,optional"`
	SetupRaidSSD   bool   `hcl:"setup_raid_ssd,optional"`
	SetupRaidSSDFS bool   `hcl:"setup_raid_ssd_fs,optional"`
}

type flatcar struct {
	Arch          string `hcl:"os_arch,optional"`
	IPXEScriptURL string `hcl:"ipxe_script_url,optional"`
}

type network struct {
	ManagementCIDRs []string `hcl:"management_cidrs"`
	NodePrivateCIDR string   `hcl:"node_private_cidr"`
}

type controller struct {
	Type string            `hcl:"type,optional"`
	Tags map[string]string `hcl:"tags,optional"`
}

type config struct {
	Metadata              *configpkg.Metadata
	Controller            *controller       `hcl:"controller,block"`
	Flatcar               *flatcar          `hcl:"flatcar,block"`
	Network               *network          `hcl:"network,block"`
	AuthToken             string            `hcl:"auth_token,optional"`
	DNS                   dns.Config        `hcl:"dns,block"`
	Facility              string            `hcl:"facility"`
	ProjectID             string            `hcl:"project_id"`
	ReservationIDs        map[string]string `hcl:"reservation_ids,optional"`
	ReservationIDsDefault string            `hcl:"reservation_ids_default,optional"`
	WorkerPools           []workerPool      `hcl:"worker_pool,block"`
}

// init registers packet as a platform
//nolint:gochecknoinits
func init() {
	configpkg.Register("packet", newConfig())
}

// newConfig returns an instance of config specific to Packet.
func newConfig() *config {
	return &config{
		Flatcar: &flatcar{
			Arch: "amd64",
		},
		Network: &network{},
		Controller: &controller{
			Type: "baremetal_0",
		},
	}
}

func (c *config) Apply(ex *terraform.Executor) error {
	dnsProvider, err := dns.ParseDNS(&c.DNS)
	if err != nil {
		return errors.Wrap(err, "parsing DNS configuration failed")
	}

	return c.terraformSmartApply(ex, dnsProvider)
}

func (c *config) Destroy(ex *terraform.Executor) error {
	if err := ex.Destroy(); err != nil {
		return fmt.Errorf("failed to destroy cluster: %v", err)
	}

	return nil
}

func (c *config) SetMetadata(metadata *configpkg.Metadata) {
	c.Metadata = metadata
}

func (c *config) Validate() hcl.Diagnostics {
	return c.checkValidConfig()
}

func (c *config) Render(cfg *configpkg.LokomotiveConfig) (string, error) {
	keyListBytes, err := json.Marshal(cfg.Controller.SSHPubKeys)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal SSH public keys")
	}

	managementCIDRs, err := json.Marshal(c.Network.ManagementCIDRs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal management CIDRs")
	}

	// Packet does not accept tags as a key-value map but as an array of
	// strings.
	util.AppendTags(&c.Controller.Tags)
	tagsList := []string{}

	for k, v := range c.Controller.Tags {
		tagsList = append(tagsList, fmt.Sprintf("%s:%s", k, v))
	}

	sort.Strings(tagsList)
	tags, err := json.Marshal(tagsList)

	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal tags")
	}

	terraformCfg := struct {
		LokomotiveConfig *configpkg.LokomotiveConfig
		PacketConfig     *config
		ControllerTags   string
		SSHPubKeys       string
		ManagementCIDRs  string
	}{
		LokomotiveConfig: cfg,
		PacketConfig:     c,
		ControllerTags:   string(tags),
		SSHPubKeys:       string(keyListBytes),
		ManagementCIDRs:  string(managementCIDRs),
	}

	return utilpkg.RenderTemplate(terraformConfigTmpl, terraformCfg)
}

// terraformSmartApply applies cluster configuration.
func (c *config) terraformSmartApply(ex *terraform.Executor, dnsProvider dns.DNSProvider) error {
	// If the provider isn't manual, apply everything in a single step.
	if dnsProvider != dns.DNSManual {
		return ex.Apply()
	}

	arguments := []string{"apply", "-auto-approve"}

	// Get DNS entries (it forces the creation of the controller nodes).
	str := fmt.Sprintf("-target=module.packet-%s.null_resource.dns_entries", c.Metadata.ClusterName)
	arguments = append(arguments, str)

	// Add worker nodes to speed things up.
	for _, w := range c.WorkerPools {
		arguments = append(arguments, fmt.Sprintf("-target=module.worker-%v.packet_device.nodes", w.Name))
	}

	// Create controller and workers nodes.
	if err := ex.Execute(arguments...); err != nil {
		return errors.Wrap(err, "failed executing Terraform")
	}

	if err := dns.AskToConfigure(ex, &c.DNS); err != nil {
		return errors.Wrap(err, "failed to configure DNS entries")
	}

	// Finish deployment.
	return ex.Apply()
}

func (c *config) GetExpectedNodes(cfg *configpkg.LokomotiveConfig) int {
	workers := 0

	for _, wp := range c.WorkerPools {
		workers += wp.Count
	}

	return cfg.Controller.Count + workers
}

// checkValidConfig validates cluster configuration.
func (c *config) checkValidConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	diagnostics = append(diagnostics, c.checkNotEmptyWorkers()...)
	diagnostics = append(diagnostics, c.checkWorkerPoolNamesUnique()...)
	diagnostics = append(diagnostics, c.checkPacketConfig()...)
	diagnostics = append(diagnostics, c.checkFlatcarConfig()...)
	diagnostics = append(diagnostics, c.checkNetworkConfig()...)
	diagnostics = append(diagnostics, c.checkControllerConfig()...)

	return diagnostics
}

func (c *config) checkPacketConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if c.AuthToken == "" && os.Getenv("PACKET_AUTH_TOKEN") == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: fmt.Sprintf("Cannot find the Packet authentication token:\n" +
				"either specify AuthToken or use the PACKET_AUTH_TOKEN environment variable"),
		})
	}

	// TODO: Get a list of valid Packet facilities and test against
	// user input
	if c.Facility == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected `facility` to be non-empty"),
		})
	}

	if c.ProjectID == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("expected `project_id` to be non-empty"),
		})
	}

	return diagnostics
}

func (c *config) checkNetworkConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if len(c.Network.ManagementCIDRs) == 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "required field `management_cidrs` is missing",
		})
	}

	for _, cidr := range c.Network.ManagementCIDRs {
		if err := validCIDR(cidr); err != nil {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid management_cidr `%s`: %v", cidr, err),
			})
		}
	}

	if c.Network.NodePrivateCIDR == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "required field `management_cidrs` is missing",
		})
	}

	if err := validCIDR(c.Network.NodePrivateCIDR); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid node_private_cidr `%s`: %v", c.Network.NodePrivateCIDR, err),
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

	archs := []string{"amd64", "arm64"}
	valid := false

	for _, arch := range archs {
		if c.Flatcar.Arch == arch {
			valid = true
		}
	}

	if !valid {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unsupported architecture, got: %s", c.Flatcar.Arch),
		})
	}

	if c.Flatcar.Arch == "arm64" && c.Flatcar.IPXEScriptURL == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "if arch is `arm64`, `ipxe_script_url` cannot be empty",
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

// checkNotEmptyWorkers checks if the cluster has at least 1 node pool defined.
func (c *config) checkNotEmptyWorkers() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if len(c.WorkerPools) == 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "one or more worker pools required",
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
			Summary:  fmt.Sprintf("worker pool name %v is not unique", w.Name),
		})
	}

	return diagnostics
}
