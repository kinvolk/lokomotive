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

//nolint:testpackage
package awsebscsidriver

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

func TestStorageClassEmptyConfig(t *testing.T) {
	configHCL := `component "aws-ebs-csi-driver" {}`

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

	storageClassFound := false
	for _, v := range m {
		storageClassFound = strings.Contains(v, "storageclass.kubernetes.io/is-default-class: \"true\"")
		if storageClassFound {
			break
		}
	}

	if !storageClassFound {
		t.Fatalf("Empty config should apply default storage class")
	}
}

func TestStorageClassEnabled(t *testing.T) {
	configHCL := `component "aws-ebs-csi-driver" {
		enable_default_storage_class = true
	}`

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

	storageClassFound := false
	for _, v := range m {
		storageClassFound = strings.Contains(v, "storageclass.kubernetes.io/is-default-class: \"true\"")
		if storageClassFound {
			break
		}
	}

	if !storageClassFound {
		t.Fatalf("Default storage class should be set")
	}
}

func TestStorageClassDisabled(t *testing.T) {
	configHCL := `component "aws-ebs-csi-driver" {
		enable_default_storage_class = false
	}`

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

	storageClassFound := true
	for _, v := range m {
		storageClassFound = strings.Contains(v, "storageclass.kubernetes.io/is-default-class: \"true\"")
		if storageClassFound {
			break
		}
	}

	if storageClassFound {
		t.Fatalf("Default storage class should not be set")
	}
}
