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

package openebsstorageclass

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
	if diagnostics.HasErrors() {
		t.Fatal("Empty config should not return errors")
	}
}

func TestDefaultStorageClass(t *testing.T) {
	c := defaultStorageClass()

	if c.ReplicaCount != 3 {
		t.Fatal("Default value of replica count should be 3")
	}
	if !c.Default {
		t.Fatal("Default value should be true")
	}
	if len(c.Disks) != 0 {
		t.Fatal("Default list of disks should be empty")
	}
}

func TestUserInputValues(t *testing.T) {

	storageClasses := `
	component "openebs-storage-class" {
		storage-class "replica1-no-disk-selected" {
			replica_count = 1
		}
		storage-class "replica1" {
			disks = ["disk1"]
			replica_count = 1
		}
		storage-class "replica3" {
			replica_count = 3
			default = true
			disks = ["disk2","disk3","disk4"]
		}
	}
	`
	testRenderManifest(t, storageClasses)
}

func testRenderManifest(t *testing.T, configHCL string) {
	component, diagnostics := util.LoadComponentFromHCLString(configHCL, name)
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}
	m, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}
	if len(m) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}
