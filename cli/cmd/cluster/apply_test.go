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

package cluster //nolint:testpackage

import (
	"reflect"
	"testing"

	"github.com/kinvolk/lokomotive/pkg/helm"
)

func Test_removeKubeletChart(t *testing.T) {
	tests := []struct {
		name   string
		charts []helm.LokomotiveChart
		want   []helm.LokomotiveChart
	}{
		{
			name: "disable_kubelet_upgrade",
			charts: []helm.LokomotiveChart{
				{Name: "foobar"}, {Name: "kubelet"}, {Name: "barbaz"},
			},
			want: []helm.LokomotiveChart{
				{Name: "foobar"}, {Name: "barbaz"},
			},
		},
		{
			// This is the possible condition when user sets the `disable_self_hosted_kubelet = true`
			// but asks to upgrade kubelet anyway.
			name: "upgrade_kubelet_but_there_is_no_kubelet",
			charts: []helm.LokomotiveChart{
				{Name: "foobar"}, {Name: "barbaz"},
			},
			want: []helm.LokomotiveChart{
				{Name: "foobar"}, {Name: "barbaz"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := removeKubeletChart(tt.charts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeKubeletChart() = %v, want %v", got, tt.want)
			}
		})
	}
}
