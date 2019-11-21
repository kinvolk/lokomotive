package velero

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
	"github.com/kinvolk/lokoctl/pkg/components/velero/azure"
)

const name = "velero"

// init registers velero component to components list, so it shows up as available to install
func init() {
	components.Register(name, newComponent())
}

// component represents component configuration data
type component struct {
	// Once we support more than one provider, this field should not be optional anymore
	Provider string `hcl:"provider,optional"`
	// Namespace where velero resources should be installed. Defaults to 'velero'.
	Namespace string `hcl:"namespace,optional"`
	// Metrics specific configuration
	Metrics *Metrics `hcl:"metrics,block"`

	// Azure specific parameters
	Azure *azure.Configuration `hcl:"azure,block"`
}

// Metrics represents prometheus specific parameters
type Metrics struct {
	Enabled        bool `hcl:"enabled,optional"`
	ServiceMonitor bool `hcl:"service_monitor,optional"`
}

// Provider requires implementing config validation function for each provider
type provider interface {
	Validate() hcl.Diagnostics
}

// newComponent creates new velero component struct with default values initialized
func newComponent() *component {
	return &component{
		Namespace: "velero",
		// Once we have more than one provider supported, we should remove the default value
		Provider: "azure",
	}
}

const chartValuesTmpl = `
configuration:
  provider: {{ .Provider }}
  backupStorageLocation:
    name: {{ .Provider }}
    bucket: {{ .Azure.BackupStorageLocation.Bucket }}
    config:
      resourceGroup: {{ .Azure.BackupStorageLocation.ResourceGroup }}
      storageAccount: {{ .Azure.BackupStorageLocation.StorageAccount }}
  volumeSnapshotLocation:
    name: {{ .Provider }}
    config:
      {{- if .Azure.VolumeSnapshotLocation.ResourceGroup }}
      resourceGroup: {{ .Azure.VolumeSnapshotLocation.ResourceGroup }}
      {{- end }}
      apitimeout: {{ .Azure.VolumeSnapshotLocation.APITimeout }}
credentials:
  secretContents:
    AZURE_SUBSCRIPTION_ID: "{{ .Azure.SubscriptionID }}"
    AZURE_TENANT_ID: "{{ .Azure.TenantID }}"
    AZURE_CLIENT_ID: "{{ .Azure.ClientID }}"
    AZURE_CLIENT_SECRET: "{{ .Azure.ClientSecret }}"
    AZURE_RESOURCE_GROUP: "{{ .Azure.ResourceGroup }}"
metrics:
  enabled: {{ .Metrics.Enabled }}
  serviceMonitor:
    enabled: {{ .Metrics.ServiceMonitor }}
`

// LoadConfig decodes given HCL and validates the configuration.
//
// If it finds any problems, HCL diagnostics array is returned containing error messages.
func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diagnostics := hcl.Diagnostics{}

	// If config is not defined at all, replace it with just empty struct, so we can
	// deserialize it and proceed
	if configBody == nil {
		// Perhaps we can skip this error?
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "component requires configuration",
			Detail:   "component has required fields in it's configuration, so configuration block must be created",
		})
		emptyConfig := hcl.EmptyBody()
		configBody = &emptyConfig
	}

	if err := gohcl.DecodeBody(*configBody, evalContext, c); err != nil {
		diagnostics = append(diagnostics, err...)
	}

	// Set default values in the component configuration if they are missing
	c.setDefaults()

	// Validate component's configuration
	diagnostics = append(diagnostics, c.validate()...)

	if diagnostics.HasErrors() {
		return diagnostics
	}

	return nil
}

// RenderManifest read helm chart from assets and renders it into list of files
func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := util.LoadChartFromAssets(fmt.Sprintf("/components/%s/manifests", name))
	if err != nil {
		return nil, errors.Wrap(err, "load chart from assets")
	}

	releaseOptions := &chartutil.ReleaseOptions{
		Name:      name,
		Namespace: c.Namespace,
		IsInstall: true,
	}

	values, err := util.RenderTemplate(chartValuesTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render chart values template")
	}

	chartConfig := &chart.Config{Raw: values}

	renderedFiles, err := util.RenderChart(helmChart, chartConfig, releaseOptions)
	if err != nil {
		return nil, errors.Wrap(err, "render chart")
	}

	return renderedFiles, nil
}

// setDefaults set default values for all nested blocks
//
// Since nested blocks in hcl2 does not support default values during DecodeBody,
// we need to set the default value here, rather then adding diagnostics.
// Once PR https://github.com/hashicorp/hcl2/pull/120 is released, this value can be set in
// newComponent() and diagnostic can be added.
func (c *component) setDefaults() {
	if c.Metrics == nil {
		c.Metrics = &Metrics{
			Enabled:        false,
			ServiceMonitor: false,
		}
	}
}

// validate validates component configuration
func (c *component) validate() hcl.Diagnostics {
	diagnostics := hcl.Diagnostics{}

	// Select provider and validate it's configuration
	p, err := c.getProvider()
	if err != nil {
		// Slice can't be constant, so just use a variable
		supportedProviders := []string{"azure"}
		return append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("provider must be one of: '%s'", strings.Join(supportedProviders[:], "', '")),
			Detail:   "Make sure to set provider to one of supported values",
		})
	}

	return append(diagnostics, p.Validate()...)
}

// getProvider returns correct provider interface based on component configuration
func (c *component) getProvider() (provider, error) {
	switch c.Provider {
	case "azure":
		return c.Azure, nil
	default:
		return nil, fmt.Errorf("unsupported provider '%s'", c.Provider)
	}
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: c.Namespace,
	}
}
