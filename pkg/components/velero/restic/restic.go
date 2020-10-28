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

// Package restic deals with configuring restic plugin.
package restic

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/internal"
)

const indentation = 6

// Configuration contains Restic specific parameters.
type Configuration struct {
	Credentials             string                 `hcl:"credentials"`
	RequireVolumeAnnotation bool                   `hcl:"require_volume_annotation,optional"`
	BackupStorageLocation   *BackupStorageLocation `hcl:"backup_storage_location,block"`
}

// BackupStorageLocation configures the backup storage location.
type BackupStorageLocation struct {
	Provider string `hcl:"provider"`
	Bucket   string `hcl:"bucket"`
	Name     string `hcl:"name,optional"`
	Region   string `hcl:"region,optional"`
}

// NewConfiguration returns the default restic configuration.
func NewConfiguration() *Configuration {
	return &Configuration{
		BackupStorageLocation: &BackupStorageLocation{},
	}
}

// Values returns plugin-specific value for Velero Helm chart.
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

// Validate validates the restic configuration.
func (c *Configuration) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if c.BackupStorageLocation == nil {
		c.BackupStorageLocation = &BackupStorageLocation{}

		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "'restic.backup_storage_location' block must be specified",
			Detail:   "Make sure to set the field to valid non-empty value",
		})
	}

	if c.BackupStorageLocation.Bucket == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "restic.backup_storage_location.bucket must not be empty",
			Detail:   "Make sure `bucket` value is set",
		})
	}

	if !isSupportedProvider(c.BackupStorageLocation.Provider) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary: fmt.Sprintf("restic.backup_storage_location.provider must be one of: '%s'",
				resticSupportedProviders()),
			Detail: "Make sure to set provider to one of supported values",
		})
	}

	return diagnostics
}

// isSupportedProvider checks if the provider is supported or not.
func isSupportedProvider(provider string) bool {
	for _, p := range resticSupportedProviders() {
		if provider == p {
			return true
		}
	}

	return false
}

// resticSupportedProviders returns the list of supported providers.
func resticSupportedProviders() []string {
	return []string{"aws", "gcp", "azure"}
}
