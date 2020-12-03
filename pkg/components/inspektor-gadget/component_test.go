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

package inspektorgadget_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"

	inspektorgadget "github.com/kinvolk/lokomotive/pkg/components/inspektor-gadget"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

func renderManifests(configHCL string) (map[string]string, error) {
	component := inspektorgadget.NewConfig()

	body, diagnostics := util.GetComponentBody(configHCL, "inspektor-gadget")
	if diagnostics != nil {
		return nil, fmt.Errorf("Getting component body: %v", diagnostics)
	}

	diagnostics = component.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("Valid config should not return an error, got: %s", diagnostics)
	}

	ret, err := component.RenderManifests()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func daemonSetFromYAML(s string) (*appsv1.DaemonSet, error) {
	i := &appsv1.DaemonSet{}
	if err := yaml.Unmarshal([]byte(s), i); err != nil {
		return nil, err
	}

	return i, nil
}

func TestRenderManifest(t *testing.T) { //nolint:funlen
	type testCase struct {
		name            string
		configHCL       string
		expectFailure   bool
		expectDaemonSet string
	}

	tcs := []testCase{
		{
			"WithEmptyConfig",
			`component "inspektor-gadget" {}`,
			false,
			`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: inspektor-gadget
  labels:
    helm.sh/chart: inspektor-gadget-0.1.0
    app.kubernetes.io/name: inspektor-gadget
    app.kubernetes.io/instance: inspektor-gadget
    app.kubernetes.io/version: "0.2.0"
    app.kubernetes.io/managed-by: Helm
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: inspektor-gadget
      app.kubernetes.io/instance: inspektor-gadget
  template:
    metadata:
      annotations:
        inspektor-gadget.kinvolk.io/option-runc-hooks: auto
        inspektor-gadget.kinvolk.io/option-traceloop: "true"
      labels:
        k8s-app: gadget # headlamp's traceloop plugin expects this
        app.kubernetes.io/name: inspektor-gadget
        app.kubernetes.io/instance: inspektor-gadget
    spec:
      serviceAccountName: inspektor-gadget
      securityContext:
        null
      hostPID: true
      hostNetwork: true
      containers:
      - name: gadget
        securityContext:
          privileged: true
        image: "kinvolk/gadget:202007010134320f732c"
        imagePullPolicy: Always
        resources:
            {}
        command: [ "/entrypoint.sh" ]
        lifecycle:
          preStop:
            exec:
              command:
              - "/cleanup.sh"
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: GADGET_POD_UID
          valueFrom:
            fieldRef:
              fieldPath: metadata.uid
        - name: TRACELOOP_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: TRACELOOP_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: TRACELOOP_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: TRACELOOP_IMAGE
          value: kinvolk/gadget:202007010134320f732c
        - name: INSPEKTOR_GADGET_VERSION
          value: 0.2.0
        - name: INSPEKTOR_GADGET_OPTION_TRACELOOP
          value: "true"
        - name: INSPEKTOR_GADGET_OPTION_TRACELOOP_LOGLEVEL
          value: info,json
        - name: INSPEKTOR_GADGET_OPTION_RUNC_HOOKS_MODE
          value: "auto"
        volumeMounts:
        - name: host
          mountPath: /host
        - name: run
          mountPath: /run
          mountPropagation: Bidirectional
        - name: modules
          mountPath: /lib/modules
        - name: debugfs
          mountPath: /sys/kernel/debug
        - name: cgroup
          mountPath: /sys/fs/cgroup
        - name: bpffs
          mountPath: /sys/fs/bpf
        - name: localtime
          mountPath: /etc/localtime
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
      - name: host
        hostPath:
          path: /
      - name: run
        hostPath:
          path: /run
      - name: cgroup
        hostPath:
          path: /sys/fs/cgroup
      - name: modules
        hostPath:
          path: /lib/modules
      - name: bpffs
        hostPath:
          path: /sys/fs/bpf
      - name: debugfs
        hostPath:
          path: /sys/kernel/debug
      - name: localtime
        hostPath:
          path: /etc/localtime`,
		},
		{
			"WithoutTraceloop",
			`component "inspektor-gadget" {
			  enable_traceloop = false
			}`,
			false,
			`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: inspektor-gadget
  labels:
    helm.sh/chart: inspektor-gadget-0.1.0
    app.kubernetes.io/name: inspektor-gadget
    app.kubernetes.io/instance: inspektor-gadget
    app.kubernetes.io/version: "0.2.0"
    app.kubernetes.io/managed-by: Helm
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: inspektor-gadget
      app.kubernetes.io/instance: inspektor-gadget
  template:
    metadata:
      labels:
        k8s-app: gadget # headlamp's traceloop plugin expects this
        app.kubernetes.io/name: inspektor-gadget
        app.kubernetes.io/instance: inspektor-gadget
    spec:
      serviceAccountName: inspektor-gadget
      securityContext:
        null
      hostPID: true
      hostNetwork: true
      containers:
      - name: gadget
        securityContext:
          privileged: true
        image: "kinvolk/gadget:202007010134320f732c"
        imagePullPolicy: Always
        resources:
            {}
        command: [ "/entrypoint.sh" ]
        lifecycle:
          preStop:
            exec:
              command:
              - "/cleanup.sh"
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: GADGET_POD_UID
          valueFrom:
            fieldRef:
              fieldPath: metadata.uid
        - name: INSPEKTOR_GADGET_VERSION
          value: 0.2.0
        - name: INSPEKTOR_GADGET_OPTION_TRACELOOP
          value: "false"
        volumeMounts:
        - name: host
          mountPath: /host
        - name: run
          mountPath: /run
          mountPropagation: Bidirectional
        - name: modules
          mountPath: /lib/modules
        - name: debugfs
          mountPath: /sys/kernel/debug
        - name: cgroup
          mountPath: /sys/fs/cgroup
        - name: bpffs
          mountPath: /sys/fs/bpf
        - name: localtime
          mountPath: /etc/localtime
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
      - name: host
        hostPath:
          path: /
      - name: run
        hostPath:
          path: /run
      - name: cgroup
        hostPath:
          path: /sys/fs/cgroup
      - name: modules
        hostPath:
          path: /lib/modules
      - name: bpffs
        hostPath:
          path: /sys/fs/bpf
      - name: debugfs
        hostPath:
          path: /sys/kernel/debug
      - name: localtime
        hostPath:
          path: /etc/localtime`,
		},
	}

	testFunc := func(t *testing.T, tc testCase) {
		m, err := renderManifests(tc.configHCL)
		if err != nil {
			if tc.expectFailure {
				return
			}

			t.Fatalf("Rendering manifests: %v", err)
		}

		got, err := daemonSetFromYAML(m["inspektor-gadget/templates/daemonset.yaml"])
		if err != nil {
			t.Fatalf("Unmarshaling ingress: %v", err)
		}

		want, err := daemonSetFromYAML(tc.expectDaemonSet)
		if err != nil {
			t.Fatalf("Unmarshaling daemonset: %v", err)
		}

		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected daemonset -want +got)\n%s", diff)
		}
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testFunc(t, tc)
		})
	}
}
