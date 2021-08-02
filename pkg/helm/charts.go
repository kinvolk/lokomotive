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

// Package helm handles Helm-related operations.
package helm

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// LokomotiveChart represents a Helm chart belonging to a Lokomotive component or control plane
// element.
type LokomotiveChart struct {
	// The name of the chart.
	Name string
	// The namespace into which the chart should be deployed.
	Namespace string
}

// ChartFromAssets finds a Helm chart in the assets at location, creates a chart.Chart struct from
// the chart and returns a pointer to it.
func ChartFromAssets(location string) (*chart.Chart, error) {
	tmpDir, err := ioutil.TempDir("", "lokoctl-chart-")
	if err != nil {
		return nil, fmt.Errorf("creating temporary directory: %w", err)
	}

	// TODO: os.RemoveAll() returns an error which we currently don't handle. Handling the error
	// isn't trivial since we call os.RemoveAll() in a defer statement.
	defer os.RemoveAll(tmpDir) //nolint: errcheck

	if err := assets.Extract(location, tmpDir); err != nil {
		return nil, fmt.Errorf("traversing assets: %w", err)
	}

	return loader.Load(tmpDir)
}
