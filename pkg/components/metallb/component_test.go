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
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/components/internal/testutil"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

func TestEmptyConfig(t *testing.T) {
	c := NewConfig()
	emptyConfig := hcl.EmptyBody()
	evalContext := hcl.EvalContext{}
	diagnostics := c.LoadConfig(&emptyConfig, &evalContext)

	if !diagnostics.HasErrors() {
		t.Fatalf("Empty config should return an error")
	}
}

func renderManifest(t *testing.T, configHCL string) map[string]string {
	component := NewConfig()

	body, diagnostics := util.GetComponentBody(configHCL, Name)
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

func getSpeakerDaemonset(t *testing.T, m map[string]string) *appsv1.DaemonSet {
	dsStr := testutil.ConfigFromMap(t, m, k8sutil.ObjectMetadata{
		Version: "apps/v1", Kind: "DaemonSet", Name: "metallb-speaker",
	})

	ds := &appsv1.DaemonSet{}
	if err := yaml.Unmarshal([]byte(dsStr), ds); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	return ds
}

func getDeployController(t *testing.T, m map[string]string) *appsv1.Deployment {
	deployStr := testutil.ConfigFromMap(t, m, k8sutil.ObjectMetadata{
		Version: "apps/v1", Kind: "Deployment", Name: "metallb-controller",
	})

	deploy := &appsv1.Deployment{}
	if err := yaml.Unmarshal([]byte(deployStr), deploy); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	return deploy
}

//nolint:funlen
func TestConversion(t *testing.T) {
	configHCL := `
component "metallb" {
  address_pools = {
	default = ["1.1.1.1/32", "2.2.2.2/32"]
  }
  speaker_toleration {
	key      = "speaker_key1"
	operator = "Equal"
	value    = "value1"
  }
  speaker_toleration {
	key      = "speaker_key2"
    operator = "Equal"
	value    = "value2"
  }
  speaker_node_selectors = {
    "speaker_node_key1" = "speaker_node_value1"
    "speaker_node_key2" = "speaker_node_value2"
  }
  controller_toleration {
	key 	 = "controller_key1"
	operator = "Equal"
	value 	 = "value1"
  }
  controller_toleration {
	key 	 = "controller_key2"
	operator = "Equal"
	value 	 = "value2"
  }
  controller_node_selectors = {
    "controller_node_key1" = "controller_node_value1"
    "controller_node_key2" = "controller_node_value2"
  }
  service_monitor = true
}`

	m := renderManifest(t, configHCL)
	if len(m) == 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}

	ds := getSpeakerDaemonset(t, m)
	deploy := getDeployController(t, m)

	tcs := []struct {
		Name string
		Test func(*testing.T, map[string]string)
	}{
		{
			"SpeakerConversions", func(t *testing.T, m map[string]string) {
				expected := []corev1.Toleration{
					{Key: "speaker_key1", Operator: "Equal", Value: "value1"},
					{Key: "speaker_key2", Operator: "Equal", Value: "value2"},
				}
				if !reflect.DeepEqual(expected, ds.Spec.Template.Spec.Tolerations) {
					t.Fatalf("expected: %#v, got: %#v", expected, ds.Spec.Template.Spec.Tolerations)
				}
			},
		},
		{
			"SpeakerNodeSelectors", func(t *testing.T, m map[string]string) {
				expected := map[string]string{
					"kubernetes.io/os":  "linux",
					"speaker_node_key1": "speaker_node_value1",
					"speaker_node_key2": "speaker_node_value2",
				}
				if !reflect.DeepEqual(expected, ds.Spec.Template.Spec.NodeSelector) {
					t.Fatalf("expected: %v, got: %v", expected, ds.Spec.Template.Spec.NodeSelector)
				}
			},
		},
		{
			"ControllerTolerations", func(t *testing.T, m map[string]string) {
				expected := []corev1.Toleration{
					{Key: "controller_key1", Operator: "Equal", Value: "value1"},
					{Key: "controller_key2", Operator: "Equal", Value: "value2"},
					{Key: "node-role.kubernetes.io/master", Effect: "NoSchedule"},
				}
				if !reflect.DeepEqual(expected, deploy.Spec.Template.Spec.Tolerations) {
					t.Fatalf("expected: %+v\ngot: %+v", expected, deploy.Spec.Template.Spec.Tolerations)
				}
			},
		},
		{
			"ControllerNodeSelectors", func(t *testing.T, m map[string]string) {
				expected := map[string]string{
					"kubernetes.io/os":          "linux",
					"controller_node_key1":      "controller_node_value1",
					"controller_node_key2":      "controller_node_value2",
					"node.kubernetes.io/master": "",
				}
				if !reflect.DeepEqual(expected, deploy.Spec.Template.Spec.NodeSelector) {
					t.Fatalf("expected: %v\ngot: %v", expected, deploy.Spec.Template.Spec.NodeSelector)
				}
			},
		},
		{
			"MonitoringConfig", func(t *testing.T, m map[string]string) {
				expectedConfig := []string{
					"metallb/templates/service.yaml",
					"metallb/templates/servicemonitor.yaml",
					"metallb/templates/grafana.yaml",
					"metallb/templates/prometheusrules-lokomotive.yaml",
				}

				for _, ec := range expectedConfig {
					if _, ok := m[ec]; !ok {
						t.Fatalf("expected %s to be generated but it is not available", ec)
					}
				}
			},
		},
		{
			"EIPConfig", func(t *testing.T, m map[string]string) {
				expectedCM := `apiVersion: v1
data:
  config: |
    address-pools:
    - addresses:
      - 1.1.1.1/32
      - 2.2.2.2/32
      name: default
      protocol: bgp
    peer-autodiscovery:
      from-annotations:
      - my-asn: metallb.lokomotive.io/my-asn
        peer-address: metallb.lokomotive.io/peer-address
        peer-asn: metallb.lokomotive.io/peer-asn
      from-labels:
      - hold-time: metallb.lokomotive.io/hold-time
        my-asn: metallb.lokomotive.io/my-asn
        peer-address: metallb.lokomotive.io/peer-address
        peer-asn: metallb.lokomotive.io/peer-asn
        peer-port: metallb.lokomotive.io/peer-port
        router-id: metallb.lokomotive.io/router-id
        source-address: metallb.lokomotive.io/src-address
kind: ConfigMap
metadata:
  name: metallb
  namespace: metallb-system
`

				gotCM := testutil.ConfigFromMap(t, m, k8sutil.ObjectMetadata{
					Version: "v1", Kind: "ConfigMap", Name: "metallb",
				})

				if gotCM != expectedCM {
					t.Fatalf("expected: %s, got: %s", expectedCM, gotCM)
				}
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Test(t, m)
		})
	}
}
