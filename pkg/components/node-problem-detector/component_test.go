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

package nodeproblemdetector_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/components/internal/testutil"
	nodeproblemdetector "github.com/kinvolk/lokomotive/pkg/components/node-problem-detector"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

const name = "node-problem-detector"

func TestRenderManifest(t *testing.T) {
	tests := []struct {
		desc    string
		hcl     string
		wantErr bool
	}{
		{
			desc: "Valid config",
			hcl: `
component "node-problem-detector" {
	custom_monitors = ["testdata"]
}
			`,
		},
	}

	for _, tc := range tests {
		b, d := util.GetComponentBody(tc.hcl, name)
		if d != nil {
			t.Errorf("%s - Error getting component body: %v", tc.desc, d)
		}

		c := nodeproblemdetector.NewConfig()

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

func TestRenderManifestConfigMapCustomMonitor(t *testing.T) {
	configHCL := `
component "node-problem-detector" {
	custom_monitors = ["testdata"]
}
`

	component := nodeproblemdetector.NewConfig()

	body, d := util.GetComponentBody(configHCL, name)
	if d.HasErrors() {
		t.Fatalf("Error getting component body: %v", d)
	}

	if d = component.LoadConfig(body, &hcl.EvalContext{}); d.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %v", d)
	}

	m := testutil.RenderManifests(t, component, name, configHCL)
	jsonPath := "{.data.custom-monitor-0\\.json}"
	expected := "testdata"

	gotConfig := testutil.ConfigFromMap(t, m, "node-problem-detector/templates/custom-config-configmap.yaml")

	testutil.MatchJSONPathStringValue(t, gotConfig, jsonPath, expected)
}
