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

// Package csi deals with configuring Velero CSI plugin.
package csi

import (
	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/pkg/components/velero/csi/drivers/aws"
)

// Configuration contains various CSI driver specific sub block.
type Configuration struct {
	// AWSDriver represents the AWS EBS CSI driver component.
	AWSDriver *aws.Configuration `hcl:"aws,block"`
}

// Values returns the plugin specific values for Velero Helm chart.
func (c *Configuration) Values() (string, error) {
	return c.AWSDriver.Values()
}

// Validate validates CSI driver specific parts in the configuration.
func (c *Configuration) Validate() hcl.Diagnostics {
	return c.validate()
}

// validate validates component configuration.
func (c *Configuration) validate() hcl.Diagnostics {
	diagnostics := hcl.Diagnostics{}
	// Since the only supported driver currently is AWS EBS CSI driver, currently we
	// don't deal with checks such as only one driver should be configured, based on
	// the driver configured, its configuration should be validated.
	// This absraction would be needed once more CSI drivers are supported.
	if c.AWSDriver == nil {
		return append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "CSI driver for AWS not configured",
			Detail:   "Make sure to configure the `aws` sub block under `csi`",
		})
	}

	return append(diagnostics, c.AWSDriver.Validate()...)
}
