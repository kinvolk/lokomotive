// Copyright 2021 The Lokomotive Authors
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

// Package openebs deals with configuring Velero openebs plugin.
package openebs

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/internal"
)

const indentation = 6

// Configuration contains OpenEBS specific parameters.
type Configuration struct {
	Credentials            string                  `hcl:"credentials"`
	Provider               string                  `hcl:"provider,optional"`
	BackupStorageLocation  *BackupStorageLocation  `hcl:"backup_storage_location,block"`
	VolumeSnapshotLocation *VolumeSnapshotLocation `hcl:"volume_snapshot_location,block"`
}

// BackupStorageLocation configures the backup storage location for OpenEBS plugin.
type BackupStorageLocation struct {
	Region   string `hcl:"region"`
	Bucket   string `hcl:"bucket"`
	Provider string `hcl:"provider,optional"`
	Name     string `hcl:"name,optional"`
}

// validate validates BackupStorageLocation struct fields.
func (b *BackupStorageLocation) validate(defaultProvider string) hcl.Diagnostics { // nolint:dupl
	var diagnostics hcl.Diagnostics

	if b == nil {
		b = &BackupStorageLocation{}

		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'openebs.backup_storage_location' block must be specified",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if b.Bucket == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'openebs.backup_storage_location.bucket' cannot be empty",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if b.Region == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'openebs.backup_storage_location.region' cannot be empty",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if b.Provider == "" && defaultProvider == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "either 'openebs.provider' or 'openebs.backup_storage_location.provider' must be set",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if !isSupportedProvider(b.Provider) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: fmt.Sprintf("openebs.backup_storage_location.provider must be one of: '%s'",
				openEBSSupportedProviders()),
			Detail: "Make sure to set provider to one of supported values",
		})
	}

	return diagnostics
}

// VolumeSnapshotLocation configures the volume snapshot location for OpenEBS plugin.
type VolumeSnapshotLocation struct {
	Bucket           string `hcl:"bucket"`
	Region           string `hcl:"region"`
	Provider         string `hcl:"provider,optional"`
	Name             string `hcl:"name,optional"`
	Prefix           string `hcl:"prefix,optional"`
	OpenEBSNamespace string `hcl:"openebs_namespace,optional"`
	S3URL            string `hcl:"s3_url,optional"`
	Local            bool   `hcl:"local,optional"`
}

// validate validates VolumeSnapshotLocation struct fields.
func (v *VolumeSnapshotLocation) validate(defaultProvider string) hcl.Diagnostics { // nolint:dupl
	var diagnostics hcl.Diagnostics

	if v == nil {
		v = &VolumeSnapshotLocation{}

		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'openebs.volume_snapshot_location' block must be specified",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if v.Bucket == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'openebs.volume_snapshot_location.bucket' cannot be empty",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if v.Region == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'openebs.volume_snapshot_location.region' cannot be empty",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if v.Provider == "" && defaultProvider == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "either 'openebs.provider' or 'openebs.volume_snapshot_location.provider' must be set",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if !isSupportedProvider(v.Provider) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: fmt.Sprintf("openebs.volume_snapshot_location.provider must be one of: '%s'",
				openEBSSupportedProviders()),
			Detail: "Make sure to set provider to one of supported values",
		})
	}

	return diagnostics
}

// Values returns the plugin specific values for Velero Helm chart.
func (c *Configuration) Values() (string, error) {
	t := template.Must(template.New("values").Parse(chartValuesTmpl))

	var buf bytes.Buffer

	v := struct {
		Configuration       *Configuration
		CredentialsIndented string
	}{
		Configuration:       c,
		CredentialsIndented: internal.Indent(c.Credentials, indentation),
	}

	if err := t.Execute(&buf, v); err != nil {
		return "", fmt.Errorf("executing values template: %w", err)
	}

	return buf.String(), nil
}

// Validate validates OpenEBS specific parts in the configuration.
func (c *Configuration) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if c.Credentials == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'credentials' cannot be empty",
			Detail:   "No credentials found",
		})
	}

	if !isSupportedProvider(c.Provider) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: fmt.Sprintf("openebs.provider must be one of: '%s'",
				openEBSSupportedProviders()),
			Detail: "Make sure to set provider to one of supported values",
		})
	}

	diagnostics = append(diagnostics, c.BackupStorageLocation.validate(c.Provider)...)
	diagnostics = append(diagnostics, c.VolumeSnapshotLocation.validate(c.Provider)...)

	return diagnostics
}

// isSupportedProvider checks if the provider is supported or not.
func isSupportedProvider(provider string) bool {
	for _, p := range openEBSSupportedProviders() {
		if provider == p {
			return true
		}
	}

	return false
}

// openEBSSupportedProviders returns the list of supported providers.
func openEBSSupportedProviders() []string {
	return []string{"aws", "gcp"}
}
