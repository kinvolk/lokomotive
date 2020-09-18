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

package components

import (
	"fmt"
	"path/filepath"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/helm"
	"helm.sh/helm/v3/pkg/chart"
)

// components is the map of registered components
var components map[string]Component

func init() {
	components = make(map[string]Component)
}

func Register(name string, obj Component) {
	if _, exists := components[name]; exists {
		panic(fmt.Sprintf("component with name %q registered already", name))
	}
	components[name] = obj
}

func ListNames() []string {
	var componentList []string
	for name := range components {
		componentList = append(componentList, name)
	}
	return componentList
}

func Get(name string) (Component, error) {
	component, exists := components[name]
	if !exists {
		return nil, fmt.Errorf("no component with name %q found", name)
	}
	return component, nil
}

// Chart is a convenience function which returns a pointer to a chart.Chart representing the
// component named name.
func Chart(name string) (*chart.Chart, error) {
	p := filepath.Join(assets.ComponentsSource, name)

	return helm.ChartFromAssets(p)
}
