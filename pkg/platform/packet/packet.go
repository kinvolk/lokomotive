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

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/dns"
	"github.com/kinvolk/lokomotive/pkg/oidc"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
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
	CLCSnippets           []string          `hcl:"clc_snippets,optional"`
	Tags                  map[string]string `hcl:"tags,optional"`
	NodesDependOn         []string          // Not exposed to the user
}

type config struct {
	AssetDir                 string            `hcl:"asset_dir"`
	AuthToken                string            `hcl:"auth_token,optional"`
	ClusterName              string            `hcl:"cluster_name"`
	Tags                     map[string]string `hcl:"tags,optional"`
	ControllerCount          int               `hcl:"controller_count"`
	ControllerType           string            `hcl:"controller_type,optional"`
	ControllerCLCSnippets    []string          `hcl:"controller_clc_snippets,optional"`
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
	DisableSelfHostedKubelet bool              `hcl:"disable_self_hosted_kubelet,optional"`
	OIDC                     *oidc.Config      `hcl:"oidc,block"`
	WorkerPools              []workerPool      `hcl:"worker_pool,block"`
	// Not exposed to the user
	KubeAPIServerExtraFlags []string
	NodesDependOn           []string
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

func (c *config) clusterDomain() string {
	return fmt.Sprintf("%s.%s", c.ClusterName, c.DNS.Zone)
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

func (c *config) Initialize(ex *terraform.Executor) error {
	if c.AuthToken == "" && os.Getenv("PACKET_AUTH_TOKEN") == "" {
		return fmt.Errorf("cannot find the Packet authentication token:\n" +
			"either specify AuthToken or use the PACKET_AUTH_TOKEN environment variable")
	}

	if err := c.DNS.Validate(); err != nil {
		return errors.Wrap(err, "parsing DNS configuration failed")
	}

	assetDir, err := homedir.Expand(c.AssetDir)
	if err != nil {
		return err
	}

	// Extract control plane chart files to cluster assets directory.
	for _, c := range platform.CommonControlPlaneCharts {
		src := filepath.Join(assets.ControlPlaneSource, c)
		dst := filepath.Join(assetDir, "cluster-assets", "charts", "kube-system", c)
		if err := assets.Extract(src, dst); err != nil {
			return errors.Wrapf(err, "Failed to extract charts")
		}
	}

	// Extract host protection chart.
	src := filepath.Join(assets.ControlPlaneSource, "calico-host-protection")
	dst := filepath.Join(assetDir,
		"cluster-assets", "charts", "kube-system", "calico-host-protection")
	if err := assets.Extract(src, dst); err != nil {
		return errors.Wrapf(err, "Failed to extract host protection chart")
	}

	// Extract self-hosted kubelet chart only when enabled in config.
	if !c.DisableSelfHostedKubelet {
		src := filepath.Join(assets.ControlPlaneSource, "kubelet")
		dst = filepath.Join(assetDir, "cluster-assets", "charts", "kube-system", "kubelet")
		if err := assets.Extract(src, dst); err != nil {
			return errors.Wrapf(err, "Failed to extract kubelet chart")
		}
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

	if err := c.Initialize(ex); err != nil {
		return err
	}

	return c.terraformSmartApply(ex, c.DNS)
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
	// Configure oidc flags and set it to KubeAPIServerExtraFlags.
	if cfg.OIDC != nil {
		// Skipping the error checking here because its done in checkValidConfig().
		oidcFlags, _ := cfg.OIDC.ToKubeAPIServerFlags(cfg.clusterDomain())
		//TODO: Use append instead of setting the oidcFlags to KubeAPIServerExtraFlags
		// append is not used for now because Initialize is called in cli/cmd/cluster.go
		// and again in Apply which duplicates the values.
		cfg.KubeAPIServerExtraFlags = oidcFlags
	}
	// Packet does not accept tags as a key-value map but as an array of
	// strings.
	platform.AppendVersionTag(&cfg.Tags)
	tagsList := []string{}
	for k, v := range cfg.Tags {
		tagsList = append(tagsList, fmt.Sprintf("%s:%s", k, v))
	}
	sort.Strings(tagsList)
	tags, err := json.Marshal(tagsList)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal tags")
	}
	// Append lokoctl-version tag to all worker pools.
	for i := range cfg.WorkerPools {
		// Using index as we are using []workerPool which creates a copy of the slice
		// Hence when the template is rendered worker pool Tags is empty.
		// TODO: Add tests for validating the worker pool configuration.
		platform.AppendVersionTag(&cfg.WorkerPools[i].Tags)
	}
	// Add explicit terraform dependencies for nodes with specific hw
	// reservation UUIDs.
	cfg.terraformAddDeps()

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
func (c *config) terraformSmartApply(ex *terraform.Executor, dc dns.Config) error {
	// If the provider isn't manual, apply everything in a single step.
	if dc.Provider != dns.Manual {
		return ex.Apply()
	}

	arguments := []string{"apply", "-auto-approve"}

	// Create controllers. We need the controllers' IP addresses before we can
	// apply the 'dns' module.
	arguments = append(arguments, fmt.Sprintf("-target=module.packet-%s.packet_device.controllers", c.ClusterName))
	if err := ex.Execute(arguments...); err != nil {
		return errors.Wrap(err, "creating controllers")
	}

	// Apply 'dns' module.
	arguments = append(arguments, "-target=module.dns")
	if err := ex.Execute(arguments...); err != nil {
		return errors.Wrap(err, "applying 'dns' module")
	}

	// Run `terraform refresh`. This is required in order to make the outputs from the previous
	// apply operations available.
	// TODO: Likely caused by https://github.com/hashicorp/terraform/issues/23158.
	if err := ex.Execute("refresh"); err != nil {
		return errors.Wrap(err, "refreshing")
	}

	// Prompt user to configure DNS.
	if err := dc.AskToConfigure(ex); err != nil {
		return errors.Wrap(err, "prompting for manual DNS configuration")
	}

	// Finish deployment.
	return ex.Apply()
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

// checkValidConfig validates cluster configuration.
func (c *config) checkValidConfig() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	diagnostics = append(diagnostics, c.checkNotEmptyWorkers()...)
	diagnostics = append(diagnostics, c.checkWorkerPoolNamesUnique()...)
	diagnostics = append(diagnostics, c.checkReservationIDs()...)
	diagnostics = append(diagnostics, c.validateOSVersion()...)

	if c.OIDC != nil {
		_, diags := c.OIDC.ToKubeAPIServerFlags(c.clusterDomain())
		diagnostics = append(diagnostics, diags...)
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
			Detail:   fmt.Sprintf("Worker pool %q is duplicated", w.Name),
		})
	}

	return diagnostics
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

// validateOSVersion ensures os_version is used only with ipxe_script_url.
func (c *config) validateOSVersion() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	// Ensure os_version is used only with ipxe_script_url.
	if c.OSVersion != "" && c.IPXEScriptURL == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "os_version is unexpected",
			Detail:   "os_version may only be specified with ipxe_script_url",
		})
	}

	for _, w := range c.WorkerPools {
		if w.OSVersion != "" && w.IPXEScriptURL == "" {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "os_version is unexpected",
				Detail:   fmt.Sprintf("os_version may only be specified with ipxe_script_url for worker pool %q", w.Name),
			})
		}
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
