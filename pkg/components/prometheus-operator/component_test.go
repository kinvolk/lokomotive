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

package prometheus

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/components/internal/testutil"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

//nolint:funlen
func TestRenderManifest(t *testing.T) {
	tests := []struct {
		desc    string
		hcl     string
		wantErr bool
	}{
		{
			desc: "essential values only",
			hcl: `
component "prometheus-operator" {
  grafana {
    admin_password = "foobar"
  }
  namespace = "monitoring"
}`,
		},
		{
			desc: "no values",
			hcl:  `component "prometheus-operator" {}`,
		},
		{
			desc: "ingress and host given",
			hcl: `
component "prometheus-operator" {
  grafana {
    ingress {
	  host = "foobar"
	}
  }
}`,
		},
		{
			desc: "ingress and no host given",
			hcl: `
component "prometheus-operator" {
  grafana {
    ingress {}
  }
}`,
			wantErr: true,
		},
		{
			desc: "prometheus ingress and external_url given and are different",
			hcl: `
component "prometheus-operator" {
  prometheus {
	external_url = "https://prometheus.notmydomain.net"
    ingress {
      host = "prometheus.mydomain.net"
    }
  }
}
`,
			wantErr: true,
		},
		{
			desc: "prometheus ingress and external_url given and are same",
			hcl: `
component "prometheus-operator" {
  prometheus {
	external_url = "https://prometheus.mydomain.net"
    ingress {
      host = "prometheus.mydomain.net"
    }
  }
}
`,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			b, d := util.GetComponentBody(tc.hcl, name)
			if d != nil {
				t.Fatalf("error getting component body: %v", d)
			}

			c := newComponent()
			d = c.LoadConfig(b, nil)

			if !tc.wantErr && d.HasErrors() {
				t.Fatalf("valid config should not return error, got: %s", d)
			}

			if tc.wantErr && !d.HasErrors() {
				t.Fatal("wrong config should have returned an error")
			} else if tc.wantErr && d.HasErrors() {
				// This means that test has passed and there is no need to go forward, we can safely
				// return.
				return
			}

			m, err := c.RenderManifests()
			if err != nil {
				t.Fatalf("rendering manifests with valid config should succeed, got: %s", err)
			}

			if len(m) == 0 {
				t.Fatal("rendered manifests shouldn't be empty")
			}
		})
	}
}

//nolint:funlen
func TestConversion(t *testing.T) {
	testCases := []struct {
		name                 string
		inputConfig          string
		expectedManifestName string
		expected             string
		jsonPath             string
	}{
		{
			name: "use external_url param",
			inputConfig: `
component "prometheus-operator" {
  prometheus {
    external_url = "https://prometheus.externalurl.net"
  }
}
`,
			expectedManifestName: "kube-prometheus-stack/templates/prometheus/prometheus.yaml",
			expected:             "https://prometheus.externalurl.net",
			jsonPath:             "{.spec.externalUrl}",
		},
		{
			name: "no external_url param",
			inputConfig: `
		component "prometheus-operator" {
		  prometheus {
		    ingress {
		      host                       = "prometheus.mydomain.net"
		      class                      = "contour"
		      certmanager_cluster_issuer = "letsencrypt-production"
		    }
		  }
		}
		`,
			expectedManifestName: "kube-prometheus-stack/templates/prometheus/prometheus.yaml",
			expected:             "https://prometheus.mydomain.net",
			jsonPath:             "{.spec.externalUrl}",
		},
		{
			name: "ingress creation for prometheus",
			inputConfig: `
		component "prometheus-operator" {
		  prometheus {
		    ingress {
		      host                       = "prometheus.mydomain.net"
		      class                      = "contour"
		      certmanager_cluster_issuer = "letsencrypt-production"
		    }
		  }
		}
		`,
			expectedManifestName: "kube-prometheus-stack/templates/prometheus/ingress.yaml",
			expected:             "prometheus.mydomain.net",
			jsonPath:             "{.spec.rules[0].host}",
		},
		{
			name:                 "verify foldersFromFilesStructure in configmap",
			inputConfig:          `component "prometheus-operator" {}`,
			expectedManifestName: "kube-prometheus-stack/charts/grafana/templates/configmap-dashboard-provider.yaml",
			expected: `apiVersion: 1
providers:
- name: 'sidecarProvider'
  orgId: 1
  folder: ''
  type: file
  disableDeletion: false
  allowUiUpdates: false
  options:
    foldersFromFilesStructure: true
    path: /tmp/dashboards`,
			jsonPath: "{.data.provider\\.yaml}",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			component := newComponent()
			m := testutil.RenderManifests(t, component, name, tc.inputConfig)
			gotConfig := testutil.ConfigFromMap(t, m, tc.expectedManifestName)

			testutil.MatchJSONPathStringValue(t, gotConfig, tc.jsonPath, tc.expected)
		})
	}
}
