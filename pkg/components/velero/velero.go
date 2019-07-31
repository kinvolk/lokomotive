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
)

const name = "velero"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	// Once we support more than one provider, this field should not be optional anymore
	Provider  string   `hcl:"provider,optional"`
	Namespace string   `hcl:"namespace,optional"`
	Metrics   *Metrics `hcl:"metrics,block"`

	// Azure-specific parameters
	Azure *AzureConfiguration `hcl:"azure,block"`
}

type Metrics struct {
	Enabled        bool `hcl:"enabled,optional"`
	ServiceMonitor bool `hcl:"service_monitor,optional"`
}

type AzureConfiguration struct {
	SubscriptionId         string                       `hcl:"subscription_id,optional"`
	TenantId               string                       `hcl:"tenant_id,optional"`
	ClientId               string                       `hcl:"client_id,optional"`
	ClientSecret           string                       `hcl:"client_secret,optional"`
	ResourceGroup          string                       `hcl:"resource_group,optional"`
	BackupStorageLocation  *AzureBackupStorageLocation  `hcl:"backup_storage_location,block"`
	VolumeSnapshotLocation *AzureVolumeSnapshotLocation `hcl:"volume_snapshot_location,block"`
}

type AzureBackupStorageLocation struct {
	ResourceGroup  string `hcl:"resource_group,optional"`
	StorageAccount string `hcl:"storage_account,optional"`
	Bucket         string `hcl:"bucket,optional"`
}

type AzureVolumeSnapshotLocation struct {
	ResourceGroup string `hcl:"resource_group,optional"`
	ApiTimeout    string `hcl:"api_timeout,optional"`
}

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
      apitimeout: {{ .Azure.VolumeSnapshotLocation.ApiTimeout }}
credentials:
  secretContents:
    AZURE_SUBSCRIPTION_ID: "{{ .Azure.SubscriptionId }}"
    AZURE_TENANT_ID: "{{ .Azure.TenantId }}"
    AZURE_CLIENT_ID: "{{ .Azure.ClientId }}"
    AZURE_CLIENT_SECRET: "{{ .Azure.ClientSecret }}"
    AZURE_RESOURCE_GROUP: "{{ .Azure.ResourceGroup }}"
metrics:
  enabled: {{ .Metrics.Enabled }}
  serviceMonitor:
    enabled: {{ .Metrics.ServiceMonitor }}
`

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

	switch c.Provider {
	case "azure":
		diagnostics = c.validateAzure(diagnostics)
	default:
		// Slice can't be constant, so just use a variable
		supportedProviders := []string{"azure"}
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("provider must be one of: '%s'", strings.Join(supportedProviders[:], "', '")),
			Detail:   "Make sure to set provider to one of supported values",
		})
	}

	if diagnostics.HasErrors() {
		return diagnostics
	}

	return nil
}

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

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}

func (c *component) validateAzure(diagnostics hcl.Diagnostics) hcl.Diagnostics {
	if c.Azure == nil {
		c.Azure = &AzureConfiguration{}
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'azure' block must exist",
			Detail:   "When using Azure provider, 'azure' block must exist",
		})
	}
	if c.Azure.SubscriptionId == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'subscription_id' must be set",
			Detail:   "When using Azure provider, 'subscription_id' property must be set",
		})
	}

	if c.Azure.TenantId == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'tenant_id' must be set",
			Detail:   "When using Azure provider, 'tenant_id' property must be set",
		})
	}

	if c.Azure.ClientId == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'client_id' must be set",
			Detail:   "When using Azure provider, 'client_id' property must be set",
		})
	}

	if c.Azure.ClientSecret == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'client_secret' must be set",
			Detail:   "When using Azure provider, 'client_secret' property must be set",
		})
	}

	if c.Azure.ResourceGroup == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'resource_group' must be set",
			Detail:   "When using Azure provider, 'resource_group' property must be set",
		})
	}

	if c.Azure.BackupStorageLocation == nil {
		c.Azure.BackupStorageLocation = &AzureBackupStorageLocation{}
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'backup_storage_location' block must exist",
			Detail:   "When using Azure provider, 'backup_storage_location' block must exist",
		})
	}

	if c.Azure.BackupStorageLocation.ResourceGroup == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'resource_group' field in block 'backup_storage_location' must be set",
			Detail:   "When using Azure provider, 'resource_group' field in block 'backup_storage_location' must be set",
		})
	}

	if c.Azure.BackupStorageLocation.StorageAccount == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'storage_account' field in block 'backup_storage_location' must be set",
			Detail:   "When using Azure provider, 'storage_account' field in block 'backup_storage_location' must be set",
		})
	}

	if c.Azure.BackupStorageLocation.Bucket == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'bucket' field in block 'backup_storage_location' must be set",
			Detail:   "When using Azure provider, 'bucket' field in block 'backup_storage_location' must be set",
		})
	}

	// Since nested blocks in hcl2 does not support default values during DecodeBody,
	// we need to set the default value here, rather then adding diagnostics.
	// Once PR https://github.com/hashicorp/hcl2/pull/120 is released, this value can be set in
	// newComponent() and diagnostic can be added.
	defaultApiTimeout := "10m"
	if c.Azure.VolumeSnapshotLocation == nil {
		c.Azure.VolumeSnapshotLocation = &AzureVolumeSnapshotLocation{
			ApiTimeout: defaultApiTimeout,
		}
	}

	if c.Azure.VolumeSnapshotLocation.ApiTimeout == "" {
		c.Azure.VolumeSnapshotLocation.ApiTimeout = defaultApiTimeout
	}

	// Same here for metrics
	if c.Metrics == nil {
		c.Metrics = &Metrics{
			Enabled:        false,
			ServiceMonitor: false,
		}
	}

	return diagnostics
}
