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
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/dns"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/platform/util"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	utilpkg "github.com/kinvolk/lokomotive/pkg/util"
)

type nodeRole int

const (
	controller = iota
	worker
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
	NodesDependOn         []string          // Not exposed to the user
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
	NodesDependOn            []string          // Not exposed to the user
	WorkerPools              []workerPool      `hcl:"worker_pool,block"`
	// Raw fields that will store the strings after unmarshalling of
	// SSHPubKeys, ManagementCIDRs and Tags
	SSHPubKeysRaw      string
	ManagementCIDRsRaw string
	TagsRaw            string
}

// init registers packet as a platform
func init() {
	platform.Register("packet", NewConfig())
}

func NewConfig() *config {
	return &config{
		ControllerCount:          1,
		ControllerType:           "baremetal_0",
		OSArch:                   "amd64",
		OSChannel:                "stable",
		OSVersion:                "current",
		NetworkMTU:               1480,
		PodCIDR:                  "10.2.0.0/16",
		ServiceCIDR:              "10.3.0.0/16",
		ClusterDomainSuffix:      "cluster.local",
		EnableReporting:          false,
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
	if c.AuthToken == "" && os.Getenv("PACKET_AUTH_TOKEN") == "" {
		return fmt.Errorf("cannot find the Packet authentication token:\n" +
			"either specify AuthToken or use the PACKET_AUTH_TOKEN environment variable")
	}

	dnsProvider, err := dns.ParseDNS(&c.DNS)
	if err != nil {
		return errors.Wrap(err, "parsing DNS configuration failed")
	}

	return c.terraformSmartApply(ex, dnsProvider)
}

// terraformSmartApply applies cluster configuration.
func (c *config) terraformSmartApply(ex *terraform.Executor, dnsProvider dns.DNSProvider) error {
	// If the provider isn't manual, apply everything in a single step.
	if dnsProvider != dns.DNSManual {
		return ex.Apply()
	}

	arguments := []string{"apply", "-auto-approve"}

	// Get DNS entries (it forces the creation of the controller nodes).
	arguments = append(arguments, fmt.Sprintf("-target=module.packet-%s.null_resource.dns_entries", c.ClusterName))

	// Create controller
	if err := ex.Execute(arguments...); err != nil {
		return errors.Wrap(err, "failed executing Terraform")
	}

	if err := dns.AskToConfigure(ex, &c.DNS); err != nil {
		return errors.Wrap(err, "failed to configure DNS entries")
	}

	// Finish deployment.
	return ex.Apply()
}

func (c *config) GetExpectedNodes() int {
	workers := 0

	for _, wp := range c.WorkerPools {
		workers += wp.Count
	}

	return c.ControllerCount + workers
}

func (c *config) Destroy(ex *terraform.Executor) error {
	return ex.Destroy()
}

func (c *config) Render() (string, error) {
	keyListBytes, err := json.Marshal(c.SSHPubKeys)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal SSH public keys")
	}

	managementCIDRs, err := json.Marshal(c.ManagementCIDRs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal management CIDRs")
	}

	// Packet does not accept tags as a key-value map but as an array of
	// strings.
	util.AppendTags(&c.Tags)
	tagsList := []string{}

	for k, v := range c.Tags {
		tagsList = append(tagsList, fmt.Sprintf("%s:%s", k, v))
	}
	sort.Strings(tagsList)
	tags, err := json.Marshal(tagsList)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal tags")
	}
	// Add explicit terraform dependencies for nodes with specific hw
	// reservation UUIDs.
	cfg.terraformAddDeps()

	c.TagsRaw = string(tags)
	c.SSHPubKeysRaw = string(keyListBytes)
	c.ManagementCIDRsRaw = string(managementCIDRs)

	return utilpkg.RenderTemplate(terraformConfigTmpl, c)
}

//nolint:funlen
func (c *config) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if err := c.setExpandedAssetDir(); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("error expanding 'asset_dir' path: %v", err),
		})
	}

	if c.AuthToken == "" && os.Getenv("PACKET_AUTH_TOKEN") == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: fmt.Sprintf("Cannot find the Packet authentication token:\n" +
				"either specify 'auth_token' or use the PACKET_AUTH_TOKEN environment variable"),
		})
	}

	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.Facility, "facility")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.ProjectID, "project_id")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.ClusterDomainSuffix, "cluster_domain_suffix")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.OSVersion, "os_version")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.AssetDir, "asset_dir")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.ClusterName, "cluster_name")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.NodePrivateCIDR, "node_private_cidr")...)
	diagnostics = append(diagnostics, util.CheckIsEmptyField(c.ControllerType, "controller_type")...)

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

	for _, cidr := range c.ManagementCIDRs {
		if err := util.IsValidCIDR(cidr); err != nil {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid management_cidr `%s`: %v", cidr, err),
			})
		}
	}

	if err := util.IsValidCIDR(c.NodePrivateCIDR); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid node_private_cidr `%s`: %v", c.NodePrivateCIDR, err),
		})
	}

	if !util.IsValidOSArch(c.OSArch) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unsupported architecture, got: %s", c.OSArch),
		})
	}

	if c.OSArch == "arm64" && c.IPXEScriptURL == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "if arch is `arm64`, `ipxe_script_url` cannot be empty",
		})
	}

	for _, wp := range c.WorkerPools {
		diagnostics = append(diagnostics, wp.checkValidWorkerPoolConfig()...)
	}

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

// checkValidConfig validates worker pool configuration.
func (wp *workerPool) checkValidWorkerPoolConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if wp.OSArch != "" && !util.IsValidOSArch(wp.OSArch) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unsupported architecture, got: %s", wp.OSArch),
		})
	}

	if wp.OSArch == "arm64" && wp.IPXEScriptURL == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "if arch is `arm64`, `ipxe_script_url` cannot be empty",
		})
	}

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
// terraformAddDeps adds explicit dependencies to cluster nodes so nodes
// with a specific hw reservation UUID are created before nodes that don't have
// a specific hw reservation UUID.
// The function modifies c.NodesDependOn and c.workerPools[].NodesDependOn,
// assigning to them a slice, of all Terraform targets needed to be created
// first. For example:
//
//	// Suppose worker pool "example" is the only to use a specific hw
//	// reservation IDs. IOW, the controllers (c.NodesDependOn) depends on
//	// worker pool "example" to be created first.
//	// Then, after calling this function, the attribute will be:
// 	c.NodesDependOn = []string{"module.worker-example.device_ids"}
//
//
// The explicit Terraform dependency is needed to guarantees that nodes using
// hardware reservation "next-available" won't use reservation IDS that another
// worker pool may specify with a specific UUID, and thus fail to create the
// node. This race condition is best explained here, if you want more info:
// https://github.com/terraform-providers/terraform-provider-packet/issues/176
// https://github.com/terraform-providers/terraform-provider-packet/pull/208
func (c *config) terraformAddDeps() {
	// Nodes with specific hw reservation IDs.
	nodesWithRes := make([]string, 0)

	// Note that dependencies expressed in Terraform are using the module
	// output "device_ids". And it is very important to keep it this way.
	//
	// If we modify it to depend only on the module, for example (just
	// "module.packet" instead of "module.packet.device_ids") it
	// seems to work fine. However, it breaks if the dependency later
	// becomes on the controller and another worker pool (e.g.
	// [ "module.packet-cluster", "module.worker-1"]) as the resources aren't of
	// the same *type*. In that case, Terraform throws this error:
	//
	// 	The given value is not suitable for child module variable
	// 	"nodes_depend_on" defined at ...:
	//	all list elements must have the same type.
	//
	// Therefore, using the output of the resources ids, this issue is
	// solved: all elements of the list (no matter if the dependency is on
	// workers, controller or both) will always have the same type and work
	// correctly, they are just resources ids (strings).
	// Also, it makes nodes depend on nodes, that is the strict dependency
	// that we really have, instead of depending in the whole module. So, it
	// allows Terraform to handle parallelization, and we only add
	// fine-grained dependencies.

	if len(c.ReservationIDs) > 0 {
		// Use a dummy tf output to wait on controllers nodes.
		tfTarget := clusterTarget(c.ClusterName, "device_ids")
		nodesWithRes = append(nodesWithRes, tfTarget)
	}

	for _, w := range c.WorkerPools {
		if len(w.ReservationIDs) > 0 {
			// Use a dummy tf output to wait on workers nodes.
			tfTarget := poolTarget(w.Name, "device_ids")
			nodesWithRes = append(nodesWithRes, tfTarget)
		}
	}

	// Collected all nodes with reservations, create a dependency on others
	// to them, so those nodes are created first.

	if len(c.ReservationIDs) == 0 {
		c.NodesDependOn = nodesWithRes
	}

	for i := range c.WorkerPools {
		if len(c.WorkerPools[i].ReservationIDs) > 0 {
			continue
		}

		c.WorkerPools[i].NodesDependOn = nodesWithRes
	}
}

// poolToTarget returns a string that can be used as "-target" argument to Terraform.
// For example:
//	// target will be "module.worker-pool1.ex".
//	target := poolTarget("pool1", "ex")
//nolint: unparam
func poolTarget(name, resource string) string {
	return fmt.Sprintf("module.worker-%v.%v", name, resource)
}

// clusterTarget returns a string that can be used as "-target" argument to Terraform.
// For example:
//	// target will be "module.packet-clusterName.ex".
//	target := clusterTarget("clusterName", "ex")
//nolint: unparam
func clusterTarget(name, resource string) string {
	return fmt.Sprintf("module.packet-%v.%v", name, resource)
}

// checkReservationIDs checks that reservations configured for controllers and
// workers are valid according to checkEachReservation().
func (c *config) checkReservationIDs() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	d := checkEachReservation(c.ReservationIDs, c.ReservationIDsDefault, c.ClusterName, controller)
	diagnostics = append(diagnostics, d...)

	for _, w := range c.WorkerPools {
		d := checkEachReservation(w.ReservationIDs, w.ReservationIDsDefault, w.Name, worker)
		diagnostics = append(diagnostics, d...)
	}

	return diagnostics
}

// checkEachReservation checks that hardware reservations are in the correct
// format and, when it will cause problems, that reservation IDs values in this
// pool are not mixed between using "next-available" and specific UUIDs, as this
// can't work reliably.
// For more info, see comment when calling terraformCreateReservations().
func checkEachReservation(reservationIDs map[string]string, resDefault, name string, role nodeRole) hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	errorPrefix := "Worker pool"
	if role == controller {
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
	// If not, Terraform code will silently ignore it. We don't want that.
	resPrefix := "worker-"
	if role == controller {
		resPrefix = "controller-"
	}

	d := checkResFormat(reservationIDs, name, errorPrefix, resPrefix)
	diagnostics = append(diagnostics, d...)

	return diagnostics
}

// checkResFormat checks that format for every key in reservationIDs is:
// <resPrefix>-<int>. resPrefix can't contain a "-" character.
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