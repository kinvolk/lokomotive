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

package webui_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

func renderManifests(configHCL string) (map[string]string, error) {
	component, err := components.Get("web-ui")
	if err != nil {
		return nil, err
	}

	body, diagnostics := util.GetComponentBody(configHCL, "web-ui")
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

func checkWebUIDeploymentNotEmpty(t *testing.T, m map[string]string) {
	dStr, ok := m["headlamp/templates/deployment.yaml"]
	if !ok {
		t.Fatalf("deployment config not found")
	}

	i := appsv1.Deployment{}
	if err := yaml.Unmarshal([]byte(dStr), i); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}
}

func ingressFromYAML(s string) (*networkingv1beta1.Ingress, error) {
	i := &networkingv1beta1.Ingress{}
	if err := yaml.Unmarshal([]byte(s), i); err != nil {
		return nil, err
	}

	return i, nil
}

func TestRenderManifest(t *testing.T) { //nolint:funlen
	type testCase struct {
		name          string
		configHCL     string
		expectFailure bool
		expectIngress string
	}

	tcs := []testCase{
		{
			"WithEmptyConfig",
			`component "web-ui" {}`,
			false,
			``,
		},
		{
			"WithEmptyIngress",
			`
			component "web-ui" {
			  ingress {}
			}`,
			true,
			``,
		},
		{
			"WithMinimalIngressBlock",
			`
component "web-ui" {
  ingress {
    host = "web-ui.test.example.com"
  }
}
`,
			false,
			`apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: web-ui
  labels:
    helm.sh/chart: headlamp-0.1.0
    app.kubernetes.io/name: web-ui
    app.kubernetes.io/instance: web-ui
    app.kubernetes.io/version: "0.1.3"
    app.kubernetes.io/managed-by: Helm
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-production
    contour.heptio.com/websocket-routes: /
    kubernetes.io/ingress.class: contour
spec:
  tls:
    - hosts:
        - "web-ui.test.example.com"
      secretName: web-ui.test.example.com-tls
  rules:
    - host: "web-ui.test.example.com"
      http:
        paths:
          - path: /
            backend:
              serviceName: web-ui
              servicePort: 80`,
		},
		{
			"WithAllParameters",
			`
			component "web-ui" {
			  ingress {
			    host = "web-ui.test.example.com"
			    class = "nginx"
			    certmanager_cluster_issuer = "letsencrypt-staging"
			  }
			}
			`,
			false,
			`apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: web-ui
  labels:
    helm.sh/chart: headlamp-0.1.0
    app.kubernetes.io/name: web-ui
    app.kubernetes.io/instance: web-ui
    app.kubernetes.io/version: "0.1.3"
    app.kubernetes.io/managed-by: Helm
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-staging
    contour.heptio.com/websocket-routes: /
    kubernetes.io/ingress.class: nginx
spec:
  tls:
    - hosts:
        - "web-ui.test.example.com"
      secretName: web-ui.test.example.com-tls
  rules:
    - host: "web-ui.test.example.com"
      http:
        paths:
          - path: /
            backend:
              serviceName: web-ui
              servicePort: 80`,
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

		if len(m) == 0 {
			t.Fatalf("Rendered manifests shouldn't be empty with valid config")
		}

		checkWebUIDeploymentNotEmpty(t, m)

		if tc.expectIngress == "" {
			return
		}

		got, err := ingressFromYAML(m["headlamp/templates/ingress.yaml"])
		if err != nil {
			t.Fatalf("Unmarshaling ingress: %v", err)
		}

		want, err := ingressFromYAML(tc.expectIngress)
		if err != nil {
			t.Fatalf("Unmarshaling deployment: %v", err)
		}

		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected deployment (-want +got)\n%s", diff)
		}
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testFunc(t, tc)
		})
	}
}
