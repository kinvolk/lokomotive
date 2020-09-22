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

// Package azure deals with configuring Velero azure plugin.
package azure

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/hashicorp/hcl/v2"
)

// Configuration contains azure specific parameters
type Configuration struct {
	SubscriptionID         string                  `hcl:"subscription_id,optional"`
	TenantID               string                  `hcl:"tenant_id,optional"`
	ClientID               string                  `hcl:"client_id,optional"`
	ClientSecret           string                  `hcl:"client_secret,optional"`
	ResourceGroup          string                  `hcl:"resource_group,optional"`
	BackupStorageLocation  *BackupStorageLocation  `hcl:"backup_storage_location,block"`
	VolumeSnapshotLocation *VolumeSnapshotLocation `hcl:"volume_snapshot_location,block"`
}

// BackupStorageLocation stores information about storage account used for backups on Azure
type BackupStorageLocation struct {
	Name           string `hcl:"name,optional"`
	ResourceGroup  string `hcl:"resource_group,optional"`
	StorageAccount string `hcl:"storage_account,optional"`
	Bucket         string `hcl:"bucket,optional"`
}

// VolumeSnapshotLocation stores information where disk snapshots will be stored on Azure
type VolumeSnapshotLocation struct {
	Name          string `hcl:"name,optional"`
	ResourceGroup string `hcl:"resource_group,optional"`
	APITimeout    string `hcl:"api_timeout,optional"`
}

// Values returns Azure-specific values for Velero Helm chart.
func (c *Configuration) Values() (string, error) {
	t := template.Must(template.New("values").Parse(chartValuesTmpl))

	var buf bytes.Buffer

	if err := t.Execute(&buf, c); err != nil {
		return "", fmt.Errorf("rendering azure values: %w", err)
	}

	return buf.String(), nil
}

// Validate validates azure specific parts in the configuration
func (c *Configuration) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if c == nil {
		c = &Configuration{}
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'azure' block must exist",
			Detail:   "When using Azure provider, 'azure' block must exist",
		})
	}
	if c.SubscriptionID == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'subscription_id' must be set",
			Detail:   "When using Azure provider, 'subscription_id' field must be set in 'azure' block",
		})
	}

	if c.TenantID == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'tenant_id' must be set",
			Detail:   "When using Azure provider, 'tenant_id' field must be set in 'azure' block",
		})
	}

	if c.ClientID == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'client_id' must be set",
			Detail:   "When using Azure provider, 'client_id' field must be set in 'azure' block",
		})
	}

	if c.ClientSecret == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'client_secret' must be set",
			Detail:   "When using Azure provider, 'client_secret' field must be set in 'azure' block",
		})
	}

	if c.ResourceGroup == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'resource_group' must be set",
			Detail:   "When using Azure provider, 'resource_group' field must be set in 'azure' block",
		})
	}

	if c.BackupStorageLocation == nil {
		c.BackupStorageLocation = &BackupStorageLocation{}
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'backup_storage_location' block must exist",
			Detail:   "When using Azure provider, 'backup_storage_location' block must exist",
		})
	}

	if c.BackupStorageLocation.ResourceGroup == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'resource_group' field in block 'backup_storage_location' must be set",
			Detail:   "When using Azure provider, 'resource_group' field in block 'backup_storage_location' must be set",
		})
	}

	if c.BackupStorageLocation.StorageAccount == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'storage_account' field in block 'backup_storage_location' must be set",
			Detail:   "When using Azure provider, 'storage_account' field in block 'backup_storage_location' must be set",
		})
	}

	if c.BackupStorageLocation.Bucket == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'bucket' field in block 'backup_storage_location' must be set",
			Detail:   "When using Azure provider, 'bucket' field in block 'backup_storage_location' must be set",
		})
	}

	// Since nested blocks in hcl2 does not support default values during DecodeBody,
	// we need to set the default value here, rather than adding diagnostics.
	// Once PR https://github.com/hashicorp/hcl2/pull/120 is released, this value can be set in
	// newComponent() and diagnostic can be added.
	defaultAPITimeout := "10m"
	if c.VolumeSnapshotLocation == nil {
		c.VolumeSnapshotLocation = &VolumeSnapshotLocation{
			APITimeout: defaultAPITimeout,
		}
	}

	if c.VolumeSnapshotLocation.APITimeout == "" {
		c.VolumeSnapshotLocation.APITimeout = defaultAPITimeout
	}

	return diagnostics
}
