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

package clusterautoscaler

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/pkg/components/util"
)

func TestEmptyConfig(t *testing.T) {
	c := NewConfig()

	emptyConfig := hcl.EmptyBody()

	evalContext := hcl.EvalContext{}

	diagnostics := c.LoadConfig(&emptyConfig, &evalContext)
	if !diagnostics.HasErrors() {
		t.Fatal("Empty config should return errors, as there are required fields.")
	}
}

func TestEmptyBody(t *testing.T) {
	c := NewConfig()

	config := `component "cluster-autoscaler" {}`

	body, diagnostics := util.GetComponentBody(config, Name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	if diagnostics := c.LoadConfig(body, &hcl.EvalContext{}); !diagnostics.HasErrors() {
		t.Fatal("Empty config should return errors as there are required fields.")
	}
}

func TestRender(t *testing.T) {
	c := NewConfig()

	config := `
  component "cluster-autoscaler" {
		cluster_name = "foo"

		worker_pool = "bar"

		packet {
			project_id = "foo"
			facility = "sjc1"
		}
	}
  `

	body, diagnostics := util.GetComponentBody(config, Name)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	if diagnostics := c.LoadConfig(body, &hcl.EvalContext{}); diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}
}
