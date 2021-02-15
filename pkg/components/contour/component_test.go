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

package contour

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/components/internal/testutil"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

//nolint:funlen
func TestRenderManifest(t *testing.T) {
	tests := []struct {
		desc    string
		hcl     string
		wantErr bool
	}{
		{
			desc: "With monitoring",
			hcl: `
component "contour" {
  enable_monitoring = true
}
			`,
		},
		{
			desc: "With service type",
			hcl: `
component "contour" {
  service_type = "NodePort"
}
			`,
		},
		{
			desc: "Wrong service type",
			hcl: `
component "contour" {
  service_type = "stuff"
}
			`,
			wantErr: true,
		},
		{
			desc: "Non-existent field",
			hcl: `
component "contour" {
  stuff = 3
}
			`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		b, d := util.GetComponentBody(tc.hcl, Name)
		if d != nil {
			t.Errorf("%s - Error getting component body: %v", tc.desc, d)
		}

		c := NewConfig()
		d = c.LoadConfig(b, nil)

		if !tc.wantErr && d.HasErrors() {
			t.Errorf("%s - Valid config should not return error, got: %s", tc.desc, d)
		}

		if tc.wantErr && !d.HasErrors() {
			t.Errorf("%s - Wrong config should have returned an error", tc.desc)
		}

		m, err := c.RenderManifests()
		if err != nil {
			t.Errorf("%s - Rendering manifests with valid config should succeed, got: %s", tc.desc, err)
		}

		if len(m) == 0 {
			t.Errorf("%s - Rendered manifests shouldn't be empty", tc.desc)
		}
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
			name: "default ServiceMonitor",
			inputConfig: `component "contour" {
				enable_monitoring = true
				envoy {
					metrics_scrape_interval = "10s"
				}
			}`,
			expectedManifestName: k8sutil.ObjectMetadata{
				Version: "monitoring.coreos.com/v1", Kind: "ServiceMonitor", Name: "envoy",
			},
			jsonPath: "{.spec.endpoints[0].interval}",
			expected: "10s",
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
