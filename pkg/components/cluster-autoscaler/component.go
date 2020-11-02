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

package clusterautoscaler

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/packethost/packngo"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents Cluster Autoscaler component name as it should be referenced in function calls
	// and in configuration.
	Name = "cluster-autoscaler"

	chartValuesTmpl = `
cloudProvider: {{ .Provider }}
nodeSelector:
  node.kubernetes.io/controller: "true"
tolerations:
- effect: NoSchedule
  key: node-role.kubernetes.io/master
  operator: Exists
rbac:
  create: true
cloudConfigPath: /config

packetClusterName: {{ .ClusterName }}
packetAuthToken: {{ .Packet.AuthToken }}
packetCloudInit: {{ .Packet.UserData }}
packetProjectID: {{ .Packet.ProjectID }}
packetFacility: {{ .Packet.Facility }}
packetOSChannel: {{ .Packet.WorkerChannel }}
packetNodeType: {{ .Packet.WorkerType }}
autoscalingGroups:
- name: {{ .WorkerPool }}
  maxSize: {{ .MaxWorkers }}
  minSize: {{ .MinWorkers }}

extraArgs:
  scale-down-unneeded-time: {{ .ScaleDownUnneededTime }}
  scale-down-delay-after-add: {{ .ScaleDownDelayAfterAdd }}
  scale-down-unready-time: {{ .ScaleDownUnreadyTime }}

podDisruptionBudget: []
kubeTargetVersionOverride: v1.17.2

{{ if .ServiceMonitor }}
serviceMonitor:
  enabled: true
  namespace: {{ .Namespace }}
  selector:
    release: prometheus-operator
{{ end }}
`
)

func init() {
	components.Register(Name, NewConfig())
}

type component struct {
	// required parameters
	Provider    string `hcl:"provider,optional"`
	WorkerPool  string `hcl:"worker_pool,optional"`
	ClusterName string `hcl:"cluster_name,optional"`

	// optional parameters
	Namespace                 string `hcl:"namespace,optional"`
	MinWorkers                int    `hcl:"min_workers,optional"`
	MaxWorkers                int    `hcl:"max_workers,optional"`
	ScaleDownUnneededTime     time.Duration
	ScaleDownUnneededTimeRaw  string `hcl:"scale_down_unneeded_time,optional"`
	ScaleDownDelayAfterAdd    time.Duration
	ScaleDownDelayAfterAddRaw string `hcl:"scale_down_delay_after_add,optional"`
	ScaleDownUnreadyTime      time.Duration
	ScaleDownUnreadyTimeRaw   string `hcl:"scale_down_unready_time,optional"`

	// Prometheus Operator related parameters.
	ServiceMonitor bool `hcl:"service_monitor,optional"`

	// Packet-specific parameters
	Packet *packetConfiguration `hcl:"packet,block"`
}

type packetConfiguration struct {
	// required parameters
	ProjectID string `hcl:"project_id,optional"`
	Facility  string `hcl:"facility,optional"`

	// optional parameters
	WorkerType    string `hcl:"worker_type,optional"`
	WorkerChannel string `hcl:"worker_channel,optional"`
	UserData      string
	AuthToken     string
}

// NewConfig returns new Cluster Autoscaler component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	c := &component{
		Provider:               "packet",
		Namespace:              "kube-system",
		MinWorkers:             1,
		MaxWorkers:             4,
		ScaleDownUnneededTime:  10 * time.Minute,
		ScaleDownDelayAfterAdd: 10 * time.Minute,
		ScaleDownUnreadyTime:   20 * time.Minute,
		ServiceMonitor:         false,
	}

	switch c.Provider {
	case "packet":
		c.Packet = &packetConfiguration{
			WorkerType:    "c3.small.x86",
			WorkerChannel: "stable",
		}
	}

	return c
}

// getClusterWorkers takes a list of devices from the user and returns list
// of worker nodes belonging to the specified cluster name and facility.
func getClusterWorkers(clusterName, facility string, devices []packngo.Device) []packngo.Device {
	clusterWorkers := []packngo.Device{}

	for _, d := range devices {
		// Skip devices from other facilities.
		if d.Facility.Code != facility {
			continue
		}

		// Skip devices from other clusters.
		if !strings.Contains(d.Hostname, clusterName) {
			continue
		}

		// Skip non-worker nodes.
		if !strings.Contains(d.Hostname, "worker") {
			continue
		}

		clusterWorkers = append(clusterWorkers, d)
	}

	return clusterWorkers
}

// findDuplicatedDevices returns duplicated devices from the given set. This can be used to check
// if all devices are unique in a set.
func findDuplicatedDevices(devices []packngo.Device) []packngo.Device {
	duplicatedDevices := []packngo.Device{}

	deviceMap := map[string]string{}

	for _, d := range devices {
		id, ok := deviceMap[d.Hostname]
		if ok && d.ID != id {
			duplicatedDevices = append(duplicatedDevices, d)
		}

		deviceMap[d.Hostname] = d.ID
	}

	return duplicatedDevices
}

// devicesHostnames returns list of devices hostnames.
func devicesHostnames(devices []packngo.Device) []string {
	hostnames := []string{}

	for _, d := range devices {
		hostnames = append(hostnames, d.Hostname)
	}

	return hostnames
}

// getWorkerUserdata finds a worker from clusterName in facility given a list
// of devices in a project and returns its user data. If two devices with the
// same name are found it returns an error.
func getWorkerUserdata(clusterName, facility string, devices []packngo.Device) (string, error) {
	workers := getClusterWorkers(clusterName, facility, devices)

	duplicates := findDuplicatedDevices(workers)
	if len(duplicates) > 0 {
		hostnames := strings.Join(devicesHostnames(duplicates), ",")

		return "", fmt.Errorf("having two devices with the same name (%q) in the same facility is not supported", hostnames)
	}

	for _, d := range workers {
		// If user data is empty for some reason, don't return it.
		if d.UserData == "" {
			continue
		}

		return base64.StdEncoding.EncodeToString([]byte(d.UserData)), nil
	}

	return "", fmt.Errorf("cluster %q must have at least one worker node with user data", clusterName)
}

// parseDurations takes the raw string time parameters from component and sets
// parsed time.Duration parameters.
func (c *component) parseDurations() hcl.Diagnostics {
	var (
		err         error
		diagnostics hcl.Diagnostics
	)

	if c.ScaleDownUnneededTimeRaw != "" {
		c.ScaleDownUnneededTime, err = time.ParseDuration(c.ScaleDownUnneededTimeRaw)
		if err != nil {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "error parsing 'scale_down_unneeded_time'",
				Detail:   fmt.Sprintf("error parsing 'scale_down_unneeded_time': %v", err),
			})
		}
	}
	if c.ScaleDownDelayAfterAddRaw != "" {
		c.ScaleDownDelayAfterAdd, err = time.ParseDuration(c.ScaleDownDelayAfterAddRaw)
		if err != nil {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "error parsing 'scale_down_delay_after_add'",
				Detail:   fmt.Sprintf("error parsing 'scale_down_delay_after_add': %v", err),
			})
		}
	}
	if c.ScaleDownUnreadyTimeRaw != "" {
		c.ScaleDownUnreadyTime, err = time.ParseDuration(c.ScaleDownUnreadyTimeRaw)
		if err != nil {
			diagnostics = append(diagnostics, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "error parsing 'scale_down_unready_time'",
				Detail:   fmt.Sprintf("error parsing 'scale_down_unready_time': %v", err),
			})
		}
	}

	return diagnostics
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diagnostics := hcl.Diagnostics{}

	// If config is not defined at all, replace it with just empty struct, so we can
	// deserialize it and proceed
	if configBody == nil {
		// Perhaps we can skip this error?
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "component requires configuration",
			Detail:   "component has required fields in its configuration, so configuration block must be created",
		})
		emptyConfig := hcl.EmptyBody()
		configBody = &emptyConfig
	}

	if err := gohcl.DecodeBody(*configBody, evalContext, c); err != nil {
		diagnostics = append(diagnostics, err...)
	}

	// work around HCL not supporting time.Duration values
	diagnostics = append(diagnostics, c.parseDurations()...)

	switch c.Provider {
	case "packet":
		diagnostics = c.validatePacket(diagnostics)
	default:
		// Slice can't be constant, so just use a variable
		supportedProviders := []string{"packet"}
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Make sure to set provider to one of supported values",
			Detail:   fmt.Sprintf("provider must be one of: '%s'", strings.Join(supportedProviders[:], "', '")),
		})
	}

	if c.WorkerPool == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'worker_pool' must be set",
			Detail:   "'worker_pool' must be set but it was not found",
		})
	}

	if c.ClusterName == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'cluster_name' must be set",
			Detail:   "'cluster_name' must be set but it was not found",
		})
	}

	return diagnostics
}

func (c *component) validatePacket(diagnostics hcl.Diagnostics) hcl.Diagnostics {
	if c.Packet == nil {
		c.Packet = &packetConfiguration{}
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'packet' block must exist",
			Detail:   "When using Packet provider, 'packet' block must exist",
		})
	}

	if c.Packet.ProjectID == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'project_id' must be set",
			Detail:   "When using Packet provider, 'project_id' must be set but it was not found",
		})
	}

	if c.Packet.Facility == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'facility' must be set",
			Detail:   "When using Packet provider, 'facility' must be set but it was not found",
		})
	}

	return diagnostics
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	if c.Provider == "packet" {
		cl, err := packngo.NewClient()
		if err != nil {
			return nil, fmt.Errorf("creating Packet API client: %w", err)
		}

		devices, _, err := cl.Devices.List(c.Packet.ProjectID, nil)
		if err != nil {
			return nil, fmt.Errorf("listing devices in project %q: %w", c.Packet.ProjectID, err)
		}

		userData, err := getWorkerUserdata(c.ClusterName, c.Packet.Facility, devices)
		if err != nil {
			return nil, fmt.Errorf("getting worker data for cluster %q: %w", c.ClusterName, err)
		}

		c.Packet.UserData = userData
		c.Packet.AuthToken = base64.StdEncoding.EncodeToString([]byte(os.Getenv("PACKET_AUTH_TOKEN")))
	}

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering chart values template: %w", err)
	}

	return util.RenderChart(helmChart, Name, c.Namespace, values)
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: c.Namespace,
		},
	}
}
