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
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/pkg/components/internal/testutil"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

func TestEmptyConfig(t *testing.T) {
	c := NewConfig()
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
	if len(m) <= 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}
}

func TestConversion(t *testing.T) {
	testCases := []struct {
		name                 string
		inputConfig          string
		expectedManifestName k8sutil.ObjectMetadata
		expected             string
		jsonPath             string
	}{
		{
			name:        "default_storage_class",
			inputConfig: `component "openebs-storage-class" {}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				// openebs-cstor-disk-replica-3 is the name of default SC created by this component.
				Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "openebs-cstor-disk-replica-3",
			},
			jsonPath: "{.reclaimPolicy}",
			expected: "Retain",
		},
		{
			name: "default_reclaim_policy",
			inputConfig: `component "openebs-storage-class" {
				storage-class "test" {}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "test",
			},
			jsonPath: "{.reclaimPolicy}",
			expected: "Retain",
		},
		{
			name: "overridden_reclaim_policy",
			inputConfig: `component "openebs-storage-class" {
				storage-class "test" {
					reclaim_policy = "Delete"
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "test",
			},
			jsonPath: "{.reclaimPolicy}",
			expected: "Delete",
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

func TestFullConversion(t *testing.T) { //nolint:funlen
	config := `component "openebs-storage-class" {
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
}`
	component := NewConfig()
	m := testutil.RenderManifests(t, component, Name, config)

	testCases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "replica1-no-disk-selected-sc",
			fn: func(t *testing.T) {
				got := testutil.ConfigFromMap(t, m, k8sutil.ObjectMetadata{
					Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "replica1-no-disk-selected",
				})

				expected := `- name: StoragePoolClaim
  value: "cstor-pool-replica1-no-disk-selected"
- name: ReplicaCount
  value: "1"
`
				testutil.MatchJSONPathStringValue(t, got, "{.metadata.annotations.cas\\.openebs\\.io/config}", expected)
			},
		},
		{
			name: "replica1-no-disk-selected-spc",
			fn: func(t *testing.T) {
				got := testutil.ConfigFromMap(t, m, k8sutil.ObjectMetadata{
					Version: "openebs.io/v1alpha1", Kind: "StoragePoolClaim", Name: "cstor-pool-replica1-no-disk-selected",
				})

				testutil.JSONPathExists(t, got, "{.spec.blockDevices}", "blockDevices is not found")
			},
		},
		{
			name: "replica1-verify-disks",
			fn: func(t *testing.T) {
				got := testutil.ConfigFromMap(t, m, k8sutil.ObjectMetadata{
					Version: "openebs.io/v1alpha1", Kind: "StoragePoolClaim", Name: "cstor-pool-replica1",
				})

				expected := "disk1"
				testutil.MatchJSONPathStringValue(t, got, "{.spec.blockDevices.blockDeviceList[0]}", expected)
			},
		},
		{
			name: "replica3-verify-disks",
			fn: func(t *testing.T) {
				got := testutil.ConfigFromMap(t, m, k8sutil.ObjectMetadata{
					Version: "openebs.io/v1alpha1", Kind: "StoragePoolClaim", Name: "cstor-pool-replica3",
				})

				expected := []string{"disk2", "disk3", "disk4"}

				for idx, exp := range expected {
					jpath := fmt.Sprintf("{.spec.blockDevices.blockDeviceList[%d]}", idx)
					testutil.MatchJSONPathStringValue(t, got, jpath, exp)
				}
			},
		},
		{
			name: "replica3-sc",
			fn: func(t *testing.T) {
				got := testutil.ConfigFromMap(t, m, k8sutil.ObjectMetadata{
					Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "replica3",
				})

				expected := "true"
				jpath := "{.metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class}"
				testutil.MatchJSONPathStringValue(t, got, jpath, expected)

				expected = `- name: StoragePoolClaim
  value: "cstor-pool-replica3"
- name: ReplicaCount
  value: "3"
`
				testutil.MatchJSONPathStringValue(t, got, "{.metadata.annotations.cas\\.openebs\\.io/config}", expected)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, tc.fn)
	}
}
