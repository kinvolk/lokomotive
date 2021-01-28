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

// Package aws deals with configuring Velero aws plugin.
package aws

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/internal"
)

const indentation = 6

// Configuration contains AWS specific parameters.
type Configuration struct {
	Credentials            string                  `hcl:"credentials"`
	BackupStorageLocation  *BackupStorageLocation  `hcl:"backup_storage_location,block"`
	VolumeSnapshotLocation *VolumeSnapshotLocation `hcl:"volume_snapshot_location,block"`
}

// BackupStorageLocation configures the backup storage location for AWS plugin.
type BackupStorageLocation struct {
	Region           string `hcl:"region"`
	Bucket           string `hcl:"bucket"`
	Name             string `hcl:"name,optional"`
	Prefix           string `hcl:"prefix,optional"`
	S3ForcePathStyle bool   `hcl:"s3_force_path_style,optional"`
	S3URL            string `hcl:"s3_url,optional"`
	PublicURL        string `hcl:"public_url,optional"`
}

// validate validates BackupStorageLocation struct fields.
func (b *BackupStorageLocation) validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if b == nil {
		b = &BackupStorageLocation{}

		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'csi.aws.backup_storage_location' block must be specified",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if b.Bucket == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'csi.aws.backup_storage_location.bucket' cannot be empty",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if b.Region == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'csi.aws.backup_storage_location.region' cannot be empty",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	return diagnostics
}

// VolumeSnapshotLocation configures the volume snapshot location for the AWS plugin.
type VolumeSnapshotLocation struct {
	Region string `hcl:"region"`
	Name   string `hcl:"name,optional"`
}

// validate validates VolumeSnapshotLocation struct fields.
func (v *VolumeSnapshotLocation) validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if v == nil {
		v = &VolumeSnapshotLocation{}

		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'csi.aws.volume_snapshot_location' block must be specified",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if v.Region == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'csi.aws.volume_snapshot_location.region' cannot be empty",
			Detail:   "Make sure to set the field to valid non-empty value",
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

// Validate validates AWS plugin specific parts in the configuration.
func (c *Configuration) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if c.Credentials == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'credentials' cannot be empty",
			Detail:   "No credentials found",
		})
	}

	diagnostics = append(diagnostics, c.BackupStorageLocation.validate()...)
	diagnostics = append(diagnostics, c.VolumeSnapshotLocation.validate()...)

	return diagnostics
}
