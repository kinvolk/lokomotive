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

	"github.com/kinvolk/lokomotive/pkg/components/internal/testutil"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

func TestStorageClassEmptyConfig(t *testing.T) {
	configHCL := `component "aws-ebs-csi-driver" {}`

	component := NewConfig()

	body, diagnostics := util.GetComponentBody(configHCL, Name)
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

	component := NewConfig()

	body, diagnostics := util.GetComponentBody(configHCL, Name)
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

	component := NewConfig()

	body, diagnostics := util.GetComponentBody(configHCL, Name)
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

func TestConversion(t *testing.T) { //nolint:funlen
	testCases := []struct {
		name                 string
		inputConfig          string
		expectedManifestName k8sutil.ObjectMetadata
		expected             string
		jsonPath             string
	}{
		{
			name:        "no_tolerations_node",
			inputConfig: `component "aws-ebs-csi-driver" {}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "apps/v1", Kind: "DaemonSet", Name: "ebs-csi-node",
			},
			jsonPath: "{.spec.template.spec.tolerations[0].operator}",
			expected: "Exists",
		},
		{
			name: "tolerations_node",
			inputConfig: `component "aws-ebs-csi-driver" {
							tolerations {
								key      = "lokomotive.io"
								operator = "Equal"
								value    = "awesome"
								effect   = "NoSchedule"
							}
						}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "apps/v1", Kind: "DaemonSet", Name: "ebs-csi-node",
			},
			jsonPath: "{.spec.template.spec.tolerations[0].value}",
			expected: "awesome",
		},
		{
			name:        "no_tolerations_csi_controller",
			inputConfig: `component "aws-ebs-csi-driver" {}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "apps/v1", Kind: "Deployment", Name: "ebs-csi-controller",
			},
			jsonPath: "{.spec.template.spec.tolerations[0].operator}",
			expected: "Exists",
		},
		{
			name: "tolerations_csi_controller",
			inputConfig: `component "aws-ebs-csi-driver" {
							tolerations {
								key      = "lokomotive.io"
								operator = "Equal"
								value    = "awesome"
								effect   = "NoSchedule"
							}
						}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "apps/v1", Kind: "Deployment", Name: "ebs-csi-controller",
			},
			jsonPath: "{.spec.template.spec.tolerations[0].key}",
			expected: "lokomotive.io",
		},
		{
			name:        "no_tolerations_snapshot_controller",
			inputConfig: `component "aws-ebs-csi-driver" {}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "apps/v1", Kind: "StatefulSet", Name: "ebs-snapshot-controller",
			},
			jsonPath: "{.spec.template.spec.tolerations[0].operator}",
			expected: "Exists",
		},
		{
			name: "tolerations_snapshot_controller",
			inputConfig: `component "aws-ebs-csi-driver" {
							tolerations {
								key      = "lokomotive.io"
								operator = "Equal"
								value    = "awesome"
								effect   = "NoSchedule"
							}
						}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "apps/v1", Kind: "StatefulSet", Name: "ebs-snapshot-controller",
			},
			jsonPath: "{.spec.template.spec.tolerations[0].effect}",
			expected: "NoSchedule",
		},
		{
			name: "affinity_csi_controller",
			inputConfig: `component "aws-ebs-csi-driver" {
							node_affinity {
								key      = "lokomotive.io/role"
								operator = "In"
								values   = ["storage"]
							}
						}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "apps/v1", Kind: "Deployment", Name: "ebs-csi-controller",
			},
			jsonPath: "{.spec.template.spec.affinity.nodeAffinity." +
				"requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key}",
			expected: "lokomotive.io/role",
		},
		{
			name: "affinity_snapshot_controller",
			inputConfig: `component "aws-ebs-csi-driver" {
							node_affinity {
								key      = "lokomotive.io/role"
								operator = "In"
								values   = ["storage"]
							}
						}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "apps/v1", Kind: "StatefulSet", Name: "ebs-snapshot-controller",
			},
			jsonPath: "{.spec.template.spec.affinity.nodeAffinity." +
				"requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]}",
			expected: "storage",
		},
		{
			name:        "storage_class",
			inputConfig: `component "aws-ebs-csi-driver" {}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "ebs-sc",
			},
			jsonPath: "{.reclaimPolicy}",
			expected: "Retain",
		},
		{
			name: "default_storage_class",
			inputConfig: `component "aws-ebs-csi-driver" {
							enable_default_storage_class = true
						}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "ebs-sc",
			},
			jsonPath: `{.metadata.annotations.storageclass\.kubernetes\.io\/is\-default\-class}`,
			expected: "true",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			component := NewConfig()
			m := testutil.RenderManifests(t, component, Name, tc.inputConfig)
			gotConfig := testutil.ConfigFromMap(t, m, tc.expectedManifestName)

			testutil.MatchJSONPathStringValue(t, gotConfig, tc.jsonPath, tc.expected)
		})
	}
}
