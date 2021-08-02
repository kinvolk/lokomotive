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

package rook_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/components/rook"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

//nolint:funlen
func TestConversion(t *testing.T) {
	configHCL := `
component "rook" {
  enable_monitoring = true

  node_selector = {
    "storage_node_key1" = "storage_node_value1"
    "storage_node_key2" = "storage_node_value2"
  }

  toleration {
    key      = "toleration_key1"
    operator = "Equal"
    value    = "toleration_value1"
    effect   = "NoSchedule"
  }

  toleration {
    key      = "toleration_key2"
    operator = "Equal"
    value    = "toleration_value2"
    effect   = "NoSchedule"
  }

  agent_toleration_key    = "agent_toleration_key"
  agent_toleration_effect = "NoSchedule"

  discover_toleration_key    = "discover_toleration_key"
  discover_toleration_effect = "NoSchedule"

  csi_plugin_node_selector = {
    "node_key1" = "node_value1"
    "node_key2" = "node_value2"
  }

  csi_plugin_toleration {
    key = "other_node_key1"
    operator = "Equal"
    value = "other_node_value1"
    effect = "NoSchedule"
  }
}
`

	m := renderManifests(t, configHCL)
	if len(m) == 0 {
		t.Fatalf("Rendered manifests shouldn't be empty")
	}

	deploy := getOperatorDeployment(t, m)

	tcs := []struct {
		Name string
		Test func(*testing.T, map[string]string, *appsv1.Deployment)
	}{
		{"MemoryResources", memoryResources},
		{"NodeSelector", nodeSelector},
		{"Monitoring", monitoring},
		{"OperatorTolerations", operatorTolerations},
		{"EnvVars", envVars},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			tc.Test(t, m, deploy)
		})
	}
}

func renderManifests(t *testing.T, configHCL string) map[string]string {
	name := "rook"

	component := rook.NewConfig()

	body, diagnostics := util.GetComponentBody(configHCL, name)
	if diagnostics != nil {
		t.Fatalf("Getting component body: %v", diagnostics)
	}

	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return an error, got: %s", diagnostics)
	}

	ret, err := component.RenderManifests()
	if err != nil {
		t.Fatalf("Rendering manifests with valid config should succeed, got: %s", err)
	}

	return ret
}

func getOperatorDeployment(t *testing.T, m map[string]string) *appsv1.Deployment {
	deployStr, ok := m["rook-ceph/templates/deployment.yaml"]
	if !ok {
		t.Fatalf("Operator deployment config not found")
	}

	deploy := &appsv1.Deployment{}
	if err := yaml.Unmarshal([]byte(deployStr), deploy); err != nil {
		t.Fatalf("Unmarshaling manifest: %v", err)
	}

	return deploy
}

func memoryResources(t *testing.T, m map[string]string, deploy *appsv1.Deployment) {
	expected := "512Mi"

	got := deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()
	if expected != got {
		t.Fatalf("Expected: %s, got: %s", expected, got)
	}
}

func nodeSelector(t *testing.T, m map[string]string, deploy *appsv1.Deployment) {
	expected := map[string]string{
		"storage_node_key1": "storage_node_value1",
		"storage_node_key2": "storage_node_value2",
	}
	got := deploy.Spec.Template.Spec.NodeSelector

	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("Expected: %v, got: %v", expected, got)
	}
}

func monitoring(t *testing.T, m map[string]string, deploy *appsv1.Deployment) {
	expectedConfig := []string{
		"rook-ceph/templates/ceph-cluster.yaml",
		"rook-ceph/templates/ceph-osd.yaml",
		"rook-ceph/templates/ceph-pools.yaml",
		"rook-ceph/templates/csi-metrics-service-monitor.yaml",
		"rook-ceph/templates/prometheus-ceph-v14-rules.yaml",
		"rook-ceph/templates/service-monitor.yaml",
	}

	for _, ec := range expectedConfig {
		if _, ok := m[ec]; !ok {
			t.Fatalf("Expected %q to be generated but it is not available", ec)
		}
	}
}

func operatorTolerations(t *testing.T, m map[string]string, deploy *appsv1.Deployment) {
	expected := []corev1.Toleration{
		{Key: "toleration_key1", Operator: "Equal", Value: "toleration_value1", Effect: "NoSchedule"},
		{Key: "toleration_key2", Operator: "Equal", Value: "toleration_value2", Effect: "NoSchedule"},
	}
	got := deploy.Spec.Template.Spec.Tolerations

	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("Expected: %v, got: %v", expected, got)
	}
}

func envVars(t *testing.T, m map[string]string, deploy *appsv1.Deployment) {
	//nolint:lll
	tcs := map[string]string{
		"AGENT_TOLERATION_KEY":          "agent_toleration_key",
		"AGENT_TOLERATION":              "NoSchedule",
		"AGENT_TOLERATIONS":             "- effect: NoSchedule\n  key: toleration_key1\n  operator: Equal\n  value: toleration_value1\n- effect: NoSchedule\n  key: toleration_key2\n  operator: Equal\n  value: toleration_value2",
		"DISCOVER_TOLERATION_KEY":       "discover_toleration_key",
		"DISCOVER_TOLERATION":           "NoSchedule",
		"DISCOVER_TOLERATIONS":          "- effect: NoSchedule\n  key: toleration_key1\n  operator: Equal\n  value: toleration_value1\n- effect: NoSchedule\n  key: toleration_key2\n  operator: Equal\n  value: toleration_value2",
		"CSI_PROVISIONER_TOLERATIONS":   "- effect: NoSchedule\n  key: toleration_key1\n  operator: Equal\n  value: toleration_value1\n- effect: NoSchedule\n  key: toleration_key2\n  operator: Equal\n  value: toleration_value2",
		"CSI_PLUGIN_TOLERATIONS":        "- effect: NoSchedule\n  key: other_node_key1\n  operator: Equal\n  value: other_node_value1",
		"AGENT_NODE_AFFINITY":           "storage_node_key1=storage_node_value1; storage_node_key2=storage_node_value2;",
		"DISCOVER_AGENT_NODE_AFFINITY":  "storage_node_key1=storage_node_value1; storage_node_key2=storage_node_value2;",
		"CSI_PROVISIONER_NODE_AFFINITY": "storage_node_key1=storage_node_value1; storage_node_key2=storage_node_value2;",
		"CSI_PLUGIN_NODE_AFFINITY":      "node_key1=node_value1; node_key2=node_value2;",
	}

	envs := deploy.Spec.Template.Spec.Containers[0].Env

	// Convert the envs into a map so that comparison becomes easier.
	envVars := map[string]string{}
	for _, en := range envs {
		envVars[en.Name] = en.Value
	}

	// For every expected key-value pair, ensure it exists in the converted output.
	for k, v := range tcs {
		k, v := k, v
		t.Run(k, func(t *testing.T) {
			gotVal, ok := envVars[k]
			if !ok {
				t.Fatalf("Expected: %q env", k)
			}

			if gotVal != v {
				t.Fatalf("Expected env var %q:%q, got %q:%q", k, v, k, gotVal)
			}
		})
	}
}
