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

package metallb

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
		t.Fatalf("Empty config should return an error")
	}
}

func renderManifest(t *testing.T, configHCL string) map[string]string {
	component := newComponent()

	body, diagnostics := util.GetComponentBody(configHCL, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}

	ret, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}

	return ret
}

func testRenderManifest(t *testing.T, configHCL string) {
	m := renderManifest(t, configHCL)
	if len(m) == 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}

func TestRenderManifestWithTolerations(t *testing.T) {
	configHCL := `
component "metallb" {
  address_pools = {
	default = ["1.1.1.1/32"]
  }
  speaker_toleration {
    key = "speaker_key1"
    operator = "Equal"
    value = "value1"
  }
  speaker_toleration {
    key = "speaker_key2"
  operator = "Equal"
    value = "value2"
  }

  controller_toleration {
    key = "controller_key1"
    operator = "Equal"
    value = "value1"
  }
  controller_toleration {
    key = "controller_key2"
    operator = "Equal"
    value = "value2"
  }
}
`
	testRenderManifest(t, configHCL)
}

func TestRenderManifestWithServiceMonitor(t *testing.T) {
	configHCL := `
component "metallb" {
  address_pools = {
    default = ["1.1.1.1/32"]
  }
  service_monitor = true
}
`
	testRenderManifest(t, configHCL)
}
