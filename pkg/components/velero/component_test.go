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

package velero //nolint:testpackage

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/pkg/components/util"
)

func TestEmptyConfig(t *testing.T) {
	c := newComponent()

	emptyConfig := hcl.EmptyBody()
	evalContext := hcl.EvalContext{}
	diagnostics := c.LoadConfig(&emptyConfig, &evalContext)

	if !diagnostics.HasErrors() {
		t.Errorf("Empty config should return error")
	}
}

func TestRenderManifestAzure(t *testing.T) {
	configHCL := `
component "velero" {
  azure {
    subscription_id  = "foo"
    tenant_id        = "foo"
    client_id        = "foo"
    client_secret    = "foo"
    resource_group   = "foo"

    backup_storage_location {
      resource_group  = "foo"
      storage_account = "foo"
      bucket          = "foo"
    }
  }
}
`

	component := newComponent()

	body, diagnostics := util.GetComponentBody(configHCL, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}

	m, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}

	if len(m) == 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}

func TestRenderManifestOpenEBS(t *testing.T) {
	configHCL := `
component "velero" {
  openebs {
    credentials = "foo"
    provider    = "aws"

    backup_storage_location {
      provider = "aws"
      bucket   = "foo"
      region   = "foo"
    }

    volume_snapshot_location {
      provider = "aws"
      bucket   = "foo"
      region   = "foo"
    }
  }
}
`

	component := newComponent()

	body, diagnostics := util.GetComponentBody(configHCL, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}

	m, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}

	if len(m) == 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}

func TestRenderManifestConflictingProviders(t *testing.T) {
	configHCL := `
component "velero" {
  azure {}
  openebs {}
}
`

	component := newComponent()

	body, d := util.GetComponentBody(configHCL, name)
	if d != nil {
		t.Fatalf("Error getting component body: %v", d)
	}

	if d := component.LoadConfig(body, &hcl.EvalContext{}); !d.HasErrors() {
		t.Fatalf("Loading configuration should fail if there is more than one provider configured")
	}
}

func TestRenderManifestNoProviderConfigured(t *testing.T) {
	configHCL := `
component "velero" {}
`

	component := newComponent()

	body, d := util.GetComponentBody(configHCL, name)
	if d != nil {
		t.Fatalf("Error getting component body: %v", d)
	}

	if d := component.LoadConfig(body, &hcl.EvalContext{}); !d.HasErrors() {
		t.Fatalf("Loading configuration should fail if there is no provider configured")
	}
}
