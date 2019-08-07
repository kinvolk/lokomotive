package clusterautoscaler

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/packethost/packngo"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

const name = "cluster-autoscaler"

const chartValuesTmpl = `
cloudProvider: {{ .Provider }}
image:
  repository: kinvolk/packet-autoscaler
  tag: 1.15.0-packet-4a17475a
  pullPolicy: IfNotPresent
nodeSelector:
  node-role.kubernetes.io/controller: "true"
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
packetWorkerPool:
- name: {{ .WorkerPool }}
  maxSize: {{ .MaxWorkers }}
  minSize: {{ .MinWorkers }}

extraArgs:
  scale-down-unneeded-time: {{ .ScaleDownUnneededTime }}
  scale-down-delay-after-add: {{ .ScaleDownDelayAfterAdd }}
  scale-down-unready-time: {{ .ScaleDownUnreadyTime }}

podDisruptionBudget: []
`

func init() {
	components.Register(name, newComponent())
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

func newComponent() *component {
	c := &component{
		Provider:               "packet",
		Namespace:              "kube-system",
		MinWorkers:             1,
		MaxWorkers:             4,
		ScaleDownUnneededTime:  10 * time.Minute,
		ScaleDownDelayAfterAdd: 10 * time.Minute,
		ScaleDownUnreadyTime:   20 * time.Minute,
	}

	switch c.Provider {
	case "packet":
		c.Packet = &packetConfiguration{
			WorkerType:    "baremetal_0",
			WorkerChannel: "stable",
		}
	}

	return c
}

// getWorkerUserdata finds a worker from clusterName in facility given a list
// of devices in a project and returns its user data. If two devices with the
// same name are found it returns an error.
func getWorkerUserdata(clusterName, facility string, devices []packngo.Device) (string, error) {
	var userData string
	deviceSet := make(map[string]struct{})

	for _, d := range devices {
		if d.Facility.Code != facility {
			continue
		}

		if _, ok := deviceSet[d.Hostname]; !ok {
			deviceSet[d.Hostname] = struct{}{}
		} else {
			return "", fmt.Errorf("having two devices with the same name (%q) in the same facility is not supported", d.Hostname)
		}

		// if device hostname contains the cluster name and "worker", we want
		// its user data
		if strings.Contains(d.Hostname, clusterName) &&
			strings.Contains(d.Hostname, "worker") {
			userData = base64.StdEncoding.EncodeToString([]byte(d.UserData))
		}
	}

	if userData == "" {
		return "", fmt.Errorf("cluster %q must have at least one worker node but no worker was found", clusterName)
	}

	return userData, nil
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
			Detail:   "When using Packet provider, 'project_id' must be set but it was not found",
		})
	}

	return diagnostics
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := util.LoadChartFromAssets(fmt.Sprintf("/components/%s", name))
	if err != nil {
		return nil, errors.Wrap(err, "load chart from assets")
	}

	releaseOptions := &chartutil.ReleaseOptions{
		Name:      name,
		Namespace: c.Namespace,
		IsInstall: true,
	}

	if c.Provider == "packet" {
		cl, err := packngo.NewClient()
		if err != nil {
			return nil, errors.Wrap(err, "create packet API client")
		}

		devices, _, err := cl.Devices.List(c.Packet.ProjectID, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "listing devices in project %q", c.Packet.ProjectID)
		}

		userData, err := getWorkerUserdata(c.ClusterName, c.Packet.Facility, devices)
		if err != nil {
			return nil, errors.Wrapf(err, "getting worker data for cluster %q", c.ClusterName)
		}

		c.Packet.UserData = userData
		c.Packet.AuthToken = base64.StdEncoding.EncodeToString([]byte(os.Getenv("PACKET_AUTH_TOKEN")))
	}

	values, err := util.RenderTemplate(chartValuesTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render chart values template")
	}

	chartConfig := &chart.Config{Raw: values}

	return util.RenderChart(helmChart, chartConfig, releaseOptions)
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
