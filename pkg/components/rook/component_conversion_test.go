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

package rook_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

//nolint:funlen
func TestConversion(t *testing.T) {
	tcs := []struct {
		desc           string
		config         string
		wantDeployment string
	}{
		{
			desc:   "Basic",
			config: `component "rook" {}`,
			wantDeployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: rook-ceph-operator
  labels:
    operator: rook
    storage-backend: ceph
    chart: "rook-ceph-v1.4.2"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rook-ceph-operator
  template:
    metadata:
      labels:
        app: rook-ceph-operator
        chart: "rook-ceph-v1.4.2"
    spec:
      containers:
      - name: rook-ceph-operator
        image: "rook/ceph:v1.4.2"
        imagePullPolicy: IfNotPresent
        args: ["ceph", "operator"]
        env:
        - name: ROOK_CURRENT_NAMESPACE_ONLY
          value: "false"
        - name: FLEXVOLUME_DIR_PATH
          value: /var/lib/kubelet/volumeplugins
        - name: ROOK_HOSTPATH_REQUIRES_PRIVILEGED
          value: "false"
        - name: ROOK_LOG_LEVEL
          value: INFO
        - name: ROOK_ENABLE_SELINUX_RELABELING
          value: "true"
        - name: ROOK_DISABLE_DEVICE_HOTPLUG
          value: "false"
        - name: ROOK_CSI_ENABLE_RBD
          value: "true"
        - name: ROOK_CSI_ENABLE_CEPHFS
          value: "true"
        - name: CSI_PLUGIN_PRIORITY_CLASSNAME
          value: 
        - name: CSI_PROVISIONER_PRIORITY_CLASSNAME
          value: 
        - name: ROOK_CSI_ENABLE_GRPC_METRICS
          value: "true"
        - name: CSI_CEPHFS_GRPC_METRICS_PORT
          value: "9092"
        - name: CSI_FORCE_CEPHFS_KERNEL_CLIENT
          value: "true"
        - name: ROOK_ENABLE_FLEX_DRIVER
          value: "false"
        - name: ROOK_ENABLE_DISCOVERY_DAEMON
          value: "true"
        - name: ROOK_OBC_WATCH_OPERATOR_NAMESPACE
          value: "true"
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: ROOK_UNREACHABLE_NODE_TOLERATION_SECONDS
          value: "5"
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 100m
            memory: 256Mi
      serviceAccountName: rook-ceph-system`,
		},

		{
			desc: "All knobs",
			config: `component "rook" {
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
		}`,
			wantDeployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: rook-ceph-operator
  labels:
    operator: rook
    storage-backend: ceph
    chart: "rook-ceph-v1.4.2"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rook-ceph-operator
  template:
    metadata:
      labels:
        app: rook-ceph-operator
        chart: "rook-ceph-v1.4.2"
    spec:
      containers:
      - name: rook-ceph-operator
        image: "rook/ceph:v1.4.2"
        imagePullPolicy: IfNotPresent
        args: ["ceph", "operator"]
        env:
        - name: ROOK_CURRENT_NAMESPACE_ONLY
          value: "false"
        - name: AGENT_TOLERATION
          value: NoSchedule
        - name: AGENT_TOLERATION_KEY
          value: agent_toleration_key
        - name: AGENT_TOLERATIONS
          value: "- effect: NoSchedule\n  key: toleration_key1\n  operator: Equal\n  value: toleration_value1\n- effect: NoSchedule\n  key: toleration_key2\n  operator: Equal\n  value: toleration_value2"
        - name: AGENT_NODE_AFFINITY
          value: storage_node_key1=storage_node_value1; storage_node_key2=storage_node_value2;
        - name: FLEXVOLUME_DIR_PATH
          value: /var/lib/kubelet/volumeplugins
        - name: DISCOVER_TOLERATION
          value: NoSchedule
        - name: DISCOVER_TOLERATION_KEY
          value: discover_toleration_key
        - name: DISCOVER_TOLERATIONS
          value: "- effect: NoSchedule\n  key: toleration_key1\n  operator: Equal\n  value: toleration_value1\n- effect: NoSchedule\n  key: toleration_key2\n  operator: Equal\n  value: toleration_value2"
        - name: DISCOVER_AGENT_NODE_AFFINITY
          value: storage_node_key1=storage_node_value1; storage_node_key2=storage_node_value2;
        - name: ROOK_HOSTPATH_REQUIRES_PRIVILEGED
          value: "false"
        - name: ROOK_LOG_LEVEL
          value: INFO
        - name: ROOK_ENABLE_SELINUX_RELABELING
          value: "true"
        - name: ROOK_DISABLE_DEVICE_HOTPLUG
          value: "false"
        - name: ROOK_CSI_ENABLE_RBD
          value: "true"
        - name: ROOK_CSI_ENABLE_CEPHFS
          value: "true"
        - name: CSI_PLUGIN_PRIORITY_CLASSNAME
          value:
        - name: CSI_PROVISIONER_PRIORITY_CLASSNAME
          value:
        - name: ROOK_CSI_ENABLE_GRPC_METRICS
          value: "true"
        - name: CSI_PROVISIONER_TOLERATIONS
          value: "- effect: NoSchedule\n  key: toleration_key1\n  operator: Equal\n  value: toleration_value1\n- effect: NoSchedule\n  key: toleration_key2\n  operator: Equal\n  value: toleration_value2"
        - name: CSI_PROVISIONER_NODE_AFFINITY
          value: storage_node_key1=storage_node_value1; storage_node_key2=storage_node_value2;
        - name: CSI_PLUGIN_TOLERATIONS
          value: "- effect: NoSchedule\n  key: other_node_key1\n  operator: Equal\n  value: other_node_value1"
        - name: CSI_PLUGIN_NODE_AFFINITY
          value: node_key1=node_value1; node_key2=node_value2;
        - name: CSI_CEPHFS_GRPC_METRICS_PORT
          value: "9092"
        - name: CSI_FORCE_CEPHFS_KERNEL_CLIENT
          value: "true"
        - name: ROOK_ENABLE_FLEX_DRIVER
          value: "false"
        - name: ROOK_ENABLE_DISCOVERY_DAEMON
          value: "true"
        - name: ROOK_OBC_WATCH_OPERATOR_NAMESPACE
          value: "true"
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: ROOK_UNREACHABLE_NODE_TOLERATION_SECONDS
          value: "5"
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 100m
            memory: 256Mi
      nodeSelector:
        storage_node_key1: storage_node_value1
        storage_node_key2: storage_node_value2
      tolerations:
        - effect: NoSchedule
          key: toleration_key1
          operator: Equal
          value: toleration_value1
        - effect: NoSchedule
          key: toleration_key2
          operator: Equal
          value: toleration_value2
      serviceAccountName: rook-ceph-system`,
		},
	}

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.desc, func(t *testing.T) {
			// Using t.Parallel() causes data races on the component's data structure because
			// multiple tests are unmarshaling configuration into the same struct in memory.
			// t.Parallel()

			m := renderManifests(t, tc.config)
			if len(m) == 0 {
				t.Fatalf("Rendered manifests shouldn't be empty")
			}

			got, err := deploymentFromYAML(m["rook-ceph/templates/deployment.yaml"])
			if err != nil {
				t.Fatalf("Unmarshaling deployment: %v", err)
			}

			want, err := deploymentFromYAML(tc.wantDeployment)
			if err != nil {
				t.Fatalf("Unmarshaling deployment: %v", err)
			}

			if diff := cmp.Diff(got, want); diff != "" {
				t.Fatalf("%q: unexpected deployment (-want +got)\n%s", tc.desc, diff)
			}
		})
	}
}

func renderManifests(t *testing.T, configHCL string) map[string]string {
	name := "rook"

	component, err := components.Get(name)
	if err != nil {
		t.Fatalf("Getting component %q: %v", name, err)
	}

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

func int32Ptr(i int32) *int32 {
	return &i
}

func deploymentFromYAML(s string) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{}

	if err := yaml.Unmarshal([]byte(s), deploy); err != nil {
		return nil, fmt.Errorf("Unmarshaling manifest: %v", err)
	}

	return deploy, nil
}
