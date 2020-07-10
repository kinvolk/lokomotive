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
	"os"
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
		t.Fatal("Empty config should return errors as AWS block is required.")
	}
}

func TestEmptyBody(t *testing.T) {
	c := newComponent()
	config := `component "external-dns" {}`
	body, diagnostics := util.GetComponentBody(config, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}
	if diagnostics := c.LoadConfig(body, &hcl.EvalContext{}); !diagnostics.HasErrors() {
		t.Fatal("Empty config should return errors as AWS block is required.")
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

	if len(c.Sources) != 1 || c.Sources[0] != "ingress" {
		t.Fatal("Default sources should be ingress only.")
	}
	if c.Policy != "upsert-only" {
		t.Fatal("Default policy should be upsert-only.")
	}
	if c.AwsConfig.ZoneType != "public" {
		t.Fatal("Default zone type in AWS should be public.")
	}
}

func TestAwsConfigWithoutProvidingCredentials(t *testing.T) {
	c := newComponent()
	config := `
 component "external-dns" {
   sources = ["ingress"]
   metrics =  false
   policy = "upsert-only"
   owner_id = "test-owner"
   aws {
     zone_id = "TESTZONEID"
     zone_type = "public"
   }
 }
 `
	// Unset AWS environment variables
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")

	body, diagnostics := util.GetComponentBody(config, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}
	if diagnostics := c.LoadConfig(body, &hcl.EvalContext{}); diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}
	if _, err := c.RenderManifests(); err == nil {
		t.Fatalf("Rendering manifests should produce error as AWS credentials were not passed")
	}
}

func TestAwsConfigBySettingEnvVariables(t *testing.T) {
	c := newComponent()
	config := `
  component "external-dns" {
    sources = ["ingress"]
    metrics =  false
    policy = "upsert-only"
    owner_id = "test-owner"
    aws {
      zone_id = "TESTZONEID"
      zone_type = "public"
    }
  }
  `
	// Set env variables.
	if err := os.Setenv("AWS_ACCESS_KEY_ID", "TESTACCESSKEY"); err != nil {
		t.Fatalf("Error setting env variable: %s", err)
	}
	if err := os.Setenv("AWS_SECRET_ACCESS_KEY", "TESTSECRETACCESSKEY"); err != nil {
		t.Fatalf("Error setting env variable: %s", err)
	}
	body, diagnostics := util.GetComponentBody(config, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}
	if diagnostics := c.LoadConfig(body, &hcl.EvalContext{}); diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}
	m, err := c.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests should not produce error as env variables were set, got: %s", err)
	}
	if len(m.Chart.Raw) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}

func TestAwsConfigBySettingEmptyEnvVariables(t *testing.T) {
	c := newComponent()
	config := `
  component "external-dns" {
    sources = ["ingress"]
    metrics =  false
    policy = "upsert-only"
    owner_id = "test-owner"
    aws {
      zone_id = "TESTZONEID"
      zone_type = "public"
    }
  }
  `
	// Set env variables.
	err := os.Setenv("AWS_ACCESS_KEY_ID", "")
	if err != nil {
		t.Fatalf("Error setting env variable: %s", err)
	}
	err = os.Setenv("AWS_SECRET_ACCESS_KEY", "")
	if err != nil {
		t.Fatalf("Error setting env variable: %s", err)
	}
	body, diagnostics := util.GetComponentBody(config, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}
	if diagnostics := c.LoadConfig(body, &hcl.EvalContext{}); diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}
	_, err = c.RenderManifests()
	if err == nil {
		t.Fatalf("Rendering manifests should produce error as AWS credentials were passed empty")
	}
}

func TestAwsConfigBySettingConfigFields(t *testing.T) {
	c := newComponent()
	config := `
  component "external-dns" {
    sources = ["ingress"]
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
	body, diagnostics := util.GetComponentBody(config, name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}
	if diagnostics := c.LoadConfig(body, &hcl.EvalContext{}); diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}
	m, err := c.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests should not produce error as config fields were set, got: %s", err)
	}
	if len(m.Chart.Raw) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}
