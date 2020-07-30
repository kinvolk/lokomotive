// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build ignore

package main

import (
	"log"

	"github.com/kinvolk/lokomotive/pkg/assets"
)

func main() {
	dirs := map[string]string{
		assets.TerraformModulesSource: "../../assets/lokomotive-kubernetes",
		assets.ControlPlaneSource:     "../../assets/charts/control-plane",
		assets.ComponentsSource:       "../../assets/charts/components",
		// This assets path is deprecated and should not be used for new components. It contains
		// manifests for components which haven't yet been converted to Helm charts.
		"/components": "../../assets/components",
	}
	err := assets.Generate("generated_assets.go", "assets", "vfsgenAssets", dirs)
	if err != nil {
		log.Fatalln(err)
	}
}
