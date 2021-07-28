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

package rookceph

import (
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
		t.Fatalf("Empty config should not return errors")
	}
}

func TestRenderManifest(t *testing.T) {
	configHCL := `
component "rook-ceph" {
  namespace = "rook-test"

  monitor_count = 3

  node_affinity {
    key      = "node-role.kubernetes.io/storage"
    operator = "Exists"
  }

  node_affinity {
    key      = "storage.lokomotive.io"
    operator = "In"

    values = [
      "foo",
    ]
  }

  toleration {
    key      = "storage.lokomotive.io"
    operator = "Equal"
    value    = "rook-ceph"
    effect   = "NoSchedule"
  }
}
`

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
		fn                   func(*testing.T, string, string, string)
	}{
		{
			name: "default_reclaim_policy",
			inputConfig: `component "rook-ceph" {
				storage_class {
					enable = true
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "rook-ceph-block",
			},
			jsonPath: "{.reclaimPolicy}",
			expected: "Retain",
			fn:       testutil.MatchJSONPathStringValue,
		},
		{
			name: "overridden_reclaim_policy",
			inputConfig: `component "rook-ceph" {
				storage_class {
					enable         = true
					reclaim_policy = "Delete"
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "storage.k8s.io/v1", Kind: "StorageClass", Name: "rook-ceph-block",
			},
			jsonPath: "{.reclaimPolicy}",
			expected: "Delete",
			fn:       testutil.MatchJSONPathStringValue,
		},
		{
			name: "resources_osd_requests_cpu",
			inputConfig: `component "rook-ceph" {
				resources {
					osd {
						requests {
							cpu = "5"
						}
					}
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.resources.osd}",
			expected: "{\"requests\":{\"cpu\":\"5\"}}",
			fn:       testutil.MatchJSONPathJSONValue,
		},
		{
			name: "resources_mgr_requests_memory",
			inputConfig: `component "rook-ceph" {
				resources {
					mgr {
						requests {
							memory = "10Gi"
						}
					}
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.resources.mgr}",
			expected: "{\"requests\":{\"memory\":\"10Gi\"}}",
			fn:       testutil.MatchJSONPathJSONValue,
		},
		{
			name: "resources_mon_limits_cpu",
			inputConfig: `component "rook-ceph" {
				resources {
					mon {
						limits {
							cpu = "3"
						}
					}
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.resources.mon}",
			expected: "{\"limits\":{\"cpu\":\"3\"}}",
			fn:       testutil.MatchJSONPathJSONValue,
		},
		{
			name: "resources_mds_limits_memory",
			inputConfig: `component "rook-ceph" {
				resources {
					mds {
						limits {
							memory = "2Gi"
						}
					}
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.resources.mds}",
			expected: "{\"limits\":{\"memory\":\"2Gi\"}}",
			fn:       testutil.MatchJSONPathJSONValue,
		},
		{
			name: "resources_prepareosd_requests",
			inputConfig: `component "rook-ceph" {
				resources {
					prepareosd {
						requests {
							memory = "9Gi"
							cpu    = "9"
						}
					}
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.resources.prepareosd}",
			expected: "{\"requests\":{\"cpu\":\"9\",\"memory\":\"9Gi\"}}",
			fn:       testutil.MatchJSONPathJSONValue,
		},
		{
			name: "resources_crashcollector_limits",
			inputConfig: `component "rook-ceph" {
				resources {
					crashcollector {
						limits {
							cpu    = "7"
							memory = "7Gi"
						}
					}
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.resources.crashcollector}",
			expected: "{\"limits\":{\"cpu\":\"7\",\"memory\":\"7Gi\"}}",
			fn:       testutil.MatchJSONPathJSONValue,
		},
		{
			name: "resources_mgr_sidecar_all",
			inputConfig: `component "rook-ceph" {
				resources {
					mgr_sidecar {
						limits {
							cpu = "6"
						}

						requests {
							memory = "1Gi"
							cpu    = "1"
						}
					}
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.resources.mgr-sidecar}",
			expected: "{\"limits\":{\"cpu\":\"6\"},\"requests\":{\"cpu\":\"1\",\"memory\":\"1Gi\"}}",
			fn:       testutil.MatchJSONPathJSONValue,
		},
		{
			name: "metadata_device",
			inputConfig: `component "rook-ceph" {
				metadata_device = "foobar"
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.storage.config.metadataDevice}",
			expected: "foobar",
			fn:       testutil.MatchJSONPathStringValue,
		},
		{
			name: "node_affinity_key",
			inputConfig: `component "rook-ceph" {
				node_affinity {
				  key      = "storage.lokomotive.io"
				  operator = "Exists"
				  values   = ["one", "two"]
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.placement.all.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution" +
				".nodeSelectorTerms[0].matchExpressions[0].key}",
			expected: "storage.lokomotive.io",
			fn:       testutil.MatchJSONPathStringValue,
		},
		{
			name: "node_affinity_value",
			inputConfig: `component "rook-ceph" {
				node_affinity {
				  key      = "storage.lokomotive.io"
				  operator = "Exists"
				  values   = ["one", "two"]
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.placement.all.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution" +
				".nodeSelectorTerms[0].matchExpressions[0].values[0]}",
			expected: "one",
			fn:       testutil.MatchJSONPathStringValue,
		},
		{
			name: "tolerations_key",
			inputConfig: `component "rook-ceph" {
				toleration {
				  key      = "storage.lokomotive.io"
				  operator = "Equal"
				  value    = "ceph"
				  effect   = "NoSchedule"
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.placement.all.tolerations[0].key}",
			expected: "storage.lokomotive.io",
			fn:       testutil.MatchJSONPathStringValue,
		},
		{
			name: "tolerations_operator",
			inputConfig: `component "rook-ceph" {
				toleration {
				  key      = "storage.lokomotive.io"
				  operator = "Equal"
				  value    = "ceph"
				  effect   = "NoSchedule"
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "ceph.rook.io/v1", Kind: "CephCluster", Name: "rook-ceph",
			},
			jsonPath: "{.spec.placement.all.tolerations[0].operator}",
			expected: "Equal",
			fn:       testutil.MatchJSONPathStringValue,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			component := NewConfig()
			m := testutil.RenderManifests(t, component, Name, tc.inputConfig)
			gotConfig := testutil.ConfigFromMap(t, m, tc.expectedManifestName)

			tc.fn(t, gotConfig, tc.jsonPath, tc.expected)
		})
	}
}
