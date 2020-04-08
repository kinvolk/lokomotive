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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/dns"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/platform/util"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

type workerPool struct {
	Name                  string            `hcl:"pool_name,label"`
	Count                 int               `hcl:"count"`
	DisableBGP            bool              `hcl:"disable_bgp,optional"`
	IPXEScriptURL         string            `hcl:"ipxe_script_url,optional"`
	OSArch                string            `hcl:"os_arch,optional"`
	OSChannel             string            `hcl:"os_channel,optional"`
	OSVersion             string            `hcl:"os_version,optional"`
	NodeType              string            `hcl:"node_type,optional"`
	Labels                string            `hcl:"labels,optional"`
	Taints                string            `hcl:"taints,optional"`
	ReservationIDs        map[string]string `hcl:"reservation_ids,optional"`
	ReservationIDsDefault string            `hcl:"reservation_ids_default,optional"`
	SetupRaid             bool              `hcl:"setup_raid,optional"`
	SetupRaidHDD          bool              `hcl:"setup_raid_hdd,optional"`
	SetupRaidSSD          bool              `hcl:"setup_raid_ssd,optional"`
	SetupRaidSSDFS        bool              `hcl:"setup_raid_ssd_fs,optional"`
}

type config struct {
	AssetDir                 string            `hcl:"asset_dir"`
	AuthToken                string            `hcl:"auth_token,optional"`
	ClusterName              string            `hcl:"cluster_name"`
	Tags                     map[string]string `hcl:"tags,optional"`
	ControllerCount          int               `hcl:"controller_count"`
	ControllerType           string            `hcl:"controller_type,optional"`
	DNS                      dns.Config        `hcl:"dns,block"`
	Facility                 string            `hcl:"facility"`
	ProjectID                string            `hcl:"project_id"`
	SSHPubKeys               []string          `hcl:"ssh_pubkeys"`
	OSArch                   string            `hcl:"os_arch,optional"`
	OSChannel                string            `hcl:"os_channel,optional"`
	OSVersion                string            `hcl:"os_version,optional"`
	IPXEScriptURL            string            `hcl:"ipxe_script_url,optional"`
	ManagementCIDRs          []string          `hcl:"management_cidrs"`
	NodePrivateCIDR          string            `hcl:"node_private_cidr"`
	EnableAggregation        bool              `hcl:"enable_aggregation,optional"`
	NetworkMTU               int               `hcl:"network_mtu,optional"`
	PodCIDR                  string            `hcl:"pod_cidr,optional"`
	ServiceCIDR              string            `hcl:"service_cidr,optional"`
	ClusterDomainSuffix      string            `hcl:"cluster_domain_suffix,optional"`
	EnableReporting          bool              `hcl:"enable_reporting,optional"`
	ReservationIDs           map[string]string `hcl:"reservation_ids,optional"`
	ReservationIDsDefault    string            `hcl:"reservation_ids_default,optional"`
	CertsValidityPeriodHours int               `hcl:"certs_validity_period_hours,optional"`

	WorkerPools []workerPool `hcl:"worker_pool,block"`
}

// init registers packet as a platform
func init() {
	platform.Register("packet", NewConfig())
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
		EnableAggregation: true,
	}
}

// GetAssetDir returns asset directory path
func (c *config) GetAssetDir() string {
	return c.AssetDir
}

func (c *config) Initialize(ex *terraform.Executor) error {
	if c.AuthToken == "" && os.Getenv("PACKET_AUTH_TOKEN") == "" {
		return fmt.Errorf("cannot find the Packet authentication token:\n" +
			"either specify AuthToken or use the PACKET_AUTH_TOKEN environment variable")
	}

	assetDir, err := homedir.Expand(c.AssetDir)
	if err != nil {
		return err
	}

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)

	return createTerraformConfigFile(c, terraformRootDir)
}

func (c *config) Apply(ex *terraform.Executor) error {
	assetDir, err := homedir.Expand(c.AssetDir)
	if err != nil {
		return err
	}

	c.AssetDir = assetDir

	dnsProvider, err := dns.ParseDNS(&c.DNS)
	if err != nil {
		return errors.Wrap(err, "parsing DNS configuration failed")
	}

	if err := c.Initialize(ex); err != nil {
		return err
	}

	return c.terraformSmartApply(ex, dnsProvider)
}

func (c *config) Destroy(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Destroy()
}

func createTerraformConfigFile(cfg *config, terraformPath string) error {
	tmplName := "cluster.tf"
	t := template.New(tmplName)
	t, err := t.Parse(terraformConfigTmpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	path := filepath.Join(terraformPath, tmplName)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", path)
	}
	defer f.Close()

	keyListBytes, err := json.Marshal(cfg.SSHPubKeys)
	if err != nil {
		return errors.Wrap(err, "failed to marshal SSH public keys")
	}

	managementCIDRs, err := json.Marshal(cfg.ManagementCIDRs)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal management CIDRs")
	}

	// Packet does not accept tags as a key-value map but as an array of
	// strings.
	util.AppendTags(&cfg.Tags)
	tagsList := []string{}
	for k, v := range cfg.Tags {
		tagsList = append(tagsList, fmt.Sprintf("%s:%s", k, v))
	}
	sort.Strings(tagsList)
	tags, err := json.Marshal(tagsList)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal tags")
	}

	terraformCfg := struct {
		Config          config
		Tags            string
		SSHPublicKeys   string
		ManagementCIDRs string
	}{
		Config:          *cfg,
		Tags:            string(tags),
		SSHPublicKeys:   string(keyListBytes),
		ManagementCIDRs: string(managementCIDRs),
	}

	if err := t.Execute(f, terraformCfg); err != nil {
		return errors.Wrapf(err, "failed to write template to file: %q", path)
	}
	return nil
}

// terraformSmartApply applies cluster configuration.
func (c *config) terraformSmartApply(ex *terraform.Executor, dnsProvider dns.DNSProvider) error {
	// Create first nodes (controllers or workers) with a reservation UUID
	// This guarantees that nodes using hardware reservation
	// "next-available" won't use reservation IDS that another worker pool
	// may specify with a specific UUID, and thus fail to create the node.
	// This race condition is best explained here, if you want more info:
	// https://github.com/terraform-providers/terraform-provider-packet/issues/176
	if err := c.terraformCreateReservations(ex); err != nil {
		return err
	}

	// If the provider isn't manual, apply everything else in a single step.
	if dnsProvider != dns.DNSManual {
		return ex.Apply()
	}

	arguments := []string{"apply", "-auto-approve"}

	// Get DNS entries (it forces the creation of the controller nodes).
	arguments = append(arguments, fmt.Sprintf("-target=module.packet-%s.null_resource.dns_entries", c.ClusterName))

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

// terraformCreateReservations creates nodes that use a specific hardware
// reservation ID.
func (c *config) terraformCreateReservations(ex *terraform.Executor) error {
	targets := []string{}

	// Create workers that use specific UUIDS as hardware reservation.
	for _, w := range c.WorkerPools {
		if len(w.ReservationIDs) > 0 {
			targets = append(targets, fmt.Sprintf("-target=module.worker-%v.packet_device.nodes", w.Name))
		}
	}

	// Create controllers that use specific UUIDS as hardware reservation.
	if len(c.ReservationIDs) > 0 {
		targets = append(targets, fmt.Sprintf("-target=module.packet-%v.packet_device.controllers", c.ClusterName))
	}

	// No "-target" arg was added, no nodes with hw reservations to create.
	if len(targets) == 0 {
		return nil
	}

	arguments := []string{"apply", "-auto-approve"}
	arguments = append(arguments, targets...)

	return ex.Execute(arguments...)
}

func (c *config) GetExpectedNodes() int {
	workers := 0

	for _, wp := range c.WorkerPools {
		workers += wp.Count
	}

	return c.ControllerCount + workers
}

// checkValidConfig validates cluster configuration.
func (c *config) checkValidConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	diagnostics = append(diagnostics, c.checkNotEmptyWorkers()...)
	diagnostics = append(diagnostics, c.checkWorkerPoolNamesUnique()...)
	diagnostics = append(diagnostics, c.checkReservationIDs()...)

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
			Detail:   fmt.Sprintf("Worker pool %q is duplicated", w.Name),
		})
	}

	return diagnostics
}

// checkReservationIDs checks that reservations configured for controllers and
// workers are valid according to checkEachReservation().
func (c *config) checkReservationIDs() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	d := checkEachReservation(c.ReservationIDs, c.ReservationIDsDefault, "controller", c.ClusterName)
	diagnostics = append(diagnostics, d...)

	for _, w := range c.WorkerPools {
		d := checkEachReservation(w.ReservationIDs, w.ReservationIDsDefault, "worker", w.Name)
		diagnostics = append(diagnostics, d...)
	}

	return diagnostics
}

// checkEachReservation checks that hardware reservations are in the correct
// format and, when it will cause problems, that reservation IDs values in this
// pool are not mixed between using "next-available" and specific UUIDs, as this
// can't work reliably.
// For more info, see comment when calling terraformCreateReservations().
func checkEachReservation(reservationIDs map[string]string, resDefault, nodeRole, name string) hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	errorPrefix := "Worker pool"
	if nodeRole == "controller" {
		errorPrefix = "Cluster"
	}

	// The following (several) checks try to avoid this: having a worker
	// pool that a node uses specific UUID as hardware reservation ID and
	// another node in the same pool that uses "next-available".
	// All different variations that in the end result in that are checked
	// below, and the reason is simple: we can't guarantee for those cases
	// that nodes can be created reliably. Creation granularity is per pool,
	// so if one pool mixes both, we can't guarantee that another pool
	// created later that needs specific UUIDs won't have them used by the
	// instances using "next-available" in the previous worker pool created.
	// This can be solved in two ways: adding more granularity, or forbidding
	// those cases. We opt for the second option, for simplicity, given that
	// in the rare case that the user needs to mix them, it can specify another
	// identical worker pool with "next-available".

	// Avoid cases that set (to non default values) reservation_ids and
	// reservation_ids_default.
	if len(reservationIDs) > 0 && resDefault != "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%v can't set both: reservation_ids and reservation_ids_default", errorPrefix),
			Detail:   fmt.Sprintf("%v: %q sets both, instead add an entry in reservations_ids for each node", errorPrefix, name),
		})
	}

	// Check reservation_ids map doesn't use "next-available" as a value.
	for _, v := range reservationIDs {
		if v != "next-available" {
			continue
		}

		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%v reservations_ids entries can't use \"next-available\"", errorPrefix),
			Detail:   fmt.Sprintf("%v: %q uses it, use specific UUIDs or reservations_ids_default only", errorPrefix, name),
		})
	}

	// Check format is:
	// controller-<int> or worker-<int>
	// If not, terraform code will silently ignore it. We don't want that.
	resPrefix := "worker-"
	if nodeRole == "controller" {
		resPrefix = "controller-"
	}

	d := checkResFormat(reservationIDs, name, errorPrefix, resPrefix)
	diagnostics = append(diagnostics, d...)

	return diagnostics
}

// checkResFormat checks that format for every key in reservationIDs is:
// <resPrefix>-<int>.
func checkResFormat(reservationIDs map[string]string, name, errorPrefix, resPrefix string) hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	for key := range reservationIDs {
		hclErr := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid reservation ID",
			Detail: fmt.Sprintf("%v: %q used %q, format should be \"%v<int>\"",
				errorPrefix, name, key, resPrefix),
		}

		// The expected format is: <resPrefix>-<int>.
		// Let's check it is this way.

		if !strings.HasPrefix(key, resPrefix) {
			diagnostics = append(diagnostics, hclErr)
			// Don't duplicate the same error, show it one per key.
			continue
		}

		resEntry := strings.Split(key, "-")
		if len(resEntry) != 2 { //nolint:gomnd
			diagnostics = append(diagnostics, hclErr)
			// Don't duplicate the same error, show it one per key.
			continue
		}

		// Check a valid number is used after "controller-" or
		// "worker-".
		index := resEntry[1]
		if _, err := strconv.Atoi(index); err != nil {
			diagnostics = append(diagnostics, hclErr)
			// Don't duplicate the same error, show it one per key.
			continue
		}
	}

	return diagnostics
}
