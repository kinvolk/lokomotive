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

package components

import (
	"path/filepath"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/helm"
	"helm.sh/helm/v3/pkg/chart"
)

// Chart is a convenience function which returns a pointer to a chart.Chart representing the
// component named name.
func Chart(name string) (*chart.Chart, error) {
	p := filepath.Join(assets.ComponentsSource, name)

	return helm.ChartFromAssets(p)
}
