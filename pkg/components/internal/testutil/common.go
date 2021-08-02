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

package testutil

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

// ConfigFromMap takes a map and a key. The function returns the YAML object associated with the
// key. If the key does not exist in that map, the function fails.
func ConfigFromMap(t *testing.T, m map[string]string, key k8sutil.ObjectMetadata) string {
	for _, v := range m {
		splittedYAML, err := k8sutil.SplitYAMLDocuments(v)
		if err != nil {
			t.Fatalf("Splitting YAML doc separated by '---': %v", err)
		}

		for _, val := range splittedYAML {
			obj, err := k8sutil.YAMLToObjectMetadata(val)
			if err != nil {
				t.Fatalf("Converting YAML to ObjectMetadata: %v", err)
			}

			if obj == key {
				return val
			}
		}
	}

	t.Fatalf("Given object not found: %+v", key)

	return ""
}

// RenderManifests converts a component into YAML manifests.
func RenderManifests(
	t *testing.T, component components.Component, componentName string, hclConfig string,
) map[string]string {
	body, diagnostics := util.GetComponentBody(hclConfig, componentName)
	if diagnostics.HasErrors() {
		t.Fatalf("Getting component body: %v", diagnostics.Errs())
	}

	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Loading configuration: %v", diagnostics)
	}

	ret, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests: %v", err)
	}

	if len(ret) == 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}

	return ret
}
