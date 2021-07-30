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

package azurearconboarding_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	azurearconboarding "github.com/kinvolk/lokomotive/pkg/components/azure-arc-onboarding"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

func TestConfig(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		desc    string
		config  string
		wantErr bool
	}{
		{
			desc: "Valid_config",
			config: `
component azure-arc-onboarding {
  application_client_id = "test-application-id"
  tenant_id             = "test-tenant-id"
  application_password  = "test-application-password"
  resource_group        = "test-resource-group"
  cluster_name          = "foobar"
}
			`,
		},
		{
			desc: "Empty_config",
			config: `
component azure-arc-onboarding {}
			`,
			wantErr: true,
		},
		{
			desc: "All_fields_set_to_empty_string",
			config: `
component azure-arc-onboarding {
  application_client_id = ""
  tenant_id             = ""
  application_password  = ""
  resource_group        = ""
  cluster_name          = ""
}
			`,
			wantErr: true,
		},
		{
			desc: "Empty_application_client_id",
			config: `
component azure-arc-onboarding {
  application_client_id = ""
  tenant_id             = "test-tenant-id"
  application_password  = "test-application-password"
  resource_group        = "test-resource-group"
  cluster_name          = "foobar"
}
			`,
			wantErr: true,
		},
		{
			desc: "Empty_tenant_id",
			config: `
component azure-arc-onboarding {
  application_client_id = "test-application-id"
  tenant_id             = ""
  application_password  = "test-application-password"
  resource_group        = "test-resource-group"
  cluster_name          = "foobar"
}
			`,
			wantErr: true,
		},
		{
			desc: "Empty_application_password",
			config: `
component azure-arc-onboarding {
  application_client_id = "test-application-id"
  tenant_id             = "tenant-id"
  application_password  = ""
  resource_group        = "test-resource-group"
  cluster_name          = "foobar"
}
			`,
			wantErr: true,
		},
		{
			desc: "Empty_resource_group",
			config: `
component azure-arc-onboarding {
  application_client_id = "test-application-id"
  tenant_id             = "test-tenant-id"
  application_password  = "test-application-password"
  resource_group        = ""
  cluster_name          = "foobar"
}
			`,
			wantErr: true,
		},
		{
			desc: "Empty_cluster_name",
			config: `
component azure-arc-onboarding {
  application_client_id = "test-application-id"
  tenant_id             = "test-tenant-id"
  application_password  = "test-application-password"
  resource_group        = "test-resource-group"
  cluster_name          = ""
}
			`,
			wantErr: true,
		},
		{
			desc: "Expect_no_error",
			config: `
component azure-arc-onboarding {
  application_client_id = "test-application-id"
  tenant_id             = "test-tenant-id"
  application_password  = "test-application-password"
  resource_group        = "test-resource-group"
  cluster_name          = "cluster-name"
}
			`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			body, diagnostics := util.GetComponentBody(test.config, azurearconboarding.Name)
			if diagnostics.HasErrors() {
				t.Fatalf("Error getting component body: %v", diagnostics)
			}

			c := azurearconboarding.NewConfig()

			diagnostics = c.LoadConfig(body, &hcl.EvalContext{})
			if test.wantErr && !diagnostics.HasErrors() {
				t.Errorf(" Failed: expected error got none")
			}

			if !test.wantErr && diagnostics.HasErrors() {
				t.Errorf(" Failed: expected no error, got: %v", diagnostics)
			}
		})
	}
}
