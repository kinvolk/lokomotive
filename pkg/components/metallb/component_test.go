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

	"github.com/kinvolk/lokomotive/pkg/components/util"
)

func TestEmptyConfig(t *testing.T) {
	c := newComponent()
	emptyConfig := hcl.EmptyBody()
	evalContext := hcl.EvalContext{}
	diagnostics := c.LoadConfig(&emptyConfig, &evalContext)
	if !diagnostics.HasErrors() {
		t.Fatalf("Empty config should return an error")
	}
}

func renderManifest(t *testing.T, configHCL string) map[string]string {
	component := newComponent()

	body, diagnostics := util.GetComponentBody(configHCL, name)
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
	dsStr, ok := m["daemonset-speaker.yaml"]
	if !ok {
		t.Fatalf("speaker daemonset config not found")
	}

	ds := &appsv1.DaemonSet{}
	if err := yaml.Unmarshal([]byte(dsStr), ds); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	return ds
}

func getDeployController(t *testing.T, m map[string]string) *appsv1.Deployment {
	deployStr, ok := m["deployment-controller.yaml"]
	if !ok {
		t.Fatalf("controller deployment config not found")
	}

	deploy := &appsv1.Deployment{}
	if err := yaml.Unmarshal([]byte(deployStr), deploy); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	return deploy
}

// nolint:funlen
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

	tcs := []struct {
		Name string
		Test func(*testing.T, map[string]string)
	}{
		{
			"SpeakerConversions", func(t *testing.T, m map[string]string) {
				ds := getSpeakerDaemonset(t, m)
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
				ds := getSpeakerDaemonset(t, m)
				expected := map[string]string{
					"beta.kubernetes.io/os": "linux",
					"speaker_node_key1":     "speaker_node_value1",
					"speaker_node_key2":     "speaker_node_value2",
				}
				if !reflect.DeepEqual(expected, ds.Spec.Template.Spec.NodeSelector) {
					t.Fatalf("expected: %v, got: %v", expected, ds.Spec.Template.Spec.NodeSelector)
				}
			},
		},
		{
			"ControllerTolerations", func(t *testing.T, m map[string]string) {
				deploy := getDeployController(t, m)
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
				deploy := getDeployController(t, m)
				expected := map[string]string{
					"beta.kubernetes.io/os":     "linux",
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
					"service.yaml",
					"service-monitor.yaml",
					"grafana-dashboard.yaml",
					"grafana-alertmanager-rule.yaml",
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
				expectedCM := `peer-autodiscovery:
  from-labels:
    my-asn: metallb.lokomotive.io/my-asn
    peer-asn: metallb.lokomotive.io/peer-asn
    peer-address: metallb.lokomotive.io/peer-address
address-pools:
- name: default
  protocol: bgp
  addresses:
  - 1.1.1.1/32
  - 2.2.2.2/32
`

				cmStr, ok := m["configmap.yaml"]
				if !ok {
					t.Fatalf("metallb configmap not found")
				}

				cm := &corev1.ConfigMap{}
				if err := yaml.Unmarshal([]byte(cmStr), cm); err != nil {
					t.Fatalf("failed unmarshalling manifest: %v", err)
				}

				gotCM, ok := cm.Data["config"]
				if !ok {
					t.Fatalf("metallb configmap is missing 'config' key")
				}

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
