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

package externaldns

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"testing"
)

func TestEmptyConfig(t *testing.T) {
	c := newComponent()
	emptyConfig := hcl.EmptyBody()
	evalContext := hcl.EvalContext{}
	diagnostics := c.LoadConfig(&emptyConfig, &evalContext)
	if !diagnostics.HasErrors() {
		t.Fatal("Empty config should return errors as AWS block is required.")
	}
}

func TestEmptyBody(t *testing.T) {
	configHCL := `component "external-dns" {}`
	_, diagnostics := util.LoadComponentFromHCLString(configHCL, name)
	if !diagnostics.HasErrors() {
		t.Fatal("Empty config should return errors as there are required fields.")
	}
}
func TestDefaultValues(t *testing.T) {
	c := newComponent()
	if c.Namespace != "external-dns" {
		t.Fatal("Default namespace for installation should be external-dns.")
	}
	if c.Metrics {
		t.Fatal("Default Metrics value should be false.")
	}
	if len(c.Sources) != 1 || c.Sources[0] != "service" {
		t.Fatal("Default sources should be service only.")
	}
	if c.Policy != "upsert-only" {
		t.Fatal("Default policy should be upsert-only.")
	}
	if c.AwsConfig.ZoneType != "public" {
		t.Fatal("Default zone type in AWS should be public.")
	}
}

func TestAwsConfigBySettingConfigFields(t *testing.T) {
	configHCL := `
  component "external-dns" {
    sources = ["service"]
    metrics =  false
    policy = "upsert-only"
    owner_id = "test-owner"
    aws {
      zone_id = "TESTZONEID"
      zone_type = "public"
      aws_access_key_id = "TESTACCESSKEY"
      aws_secret_access_key = "TESTSECRETACCESSKEY"
    }
  }
  `
	component, diagnostics := util.LoadComponentFromHCLString(configHCL, name)
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}
	m, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests should not produce error as config fields were set, got: %s", err)
	}
	if len(m) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}
