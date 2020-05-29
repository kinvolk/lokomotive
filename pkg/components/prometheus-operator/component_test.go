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

	"github.com/kinvolk/lokomotive/pkg/components/util"
)

// nolint:funlen
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
