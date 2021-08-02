// Copyright 2018 The Prometheus Authors
// Copyright 2021 The Lokomotive Authors

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
	"net/http"
	"time"

	// External dependencies should also be added to assets.go file.
	"github.com/prometheus/alertmanager/pkg/modtimevfs"
	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"

	"github.com/kinvolk/lokomotive/pkg/assets"
)

func main() {
	directoriesToEmbed := map[string]http.FileSystem{
		assets.TerraformModulesSource: http.Dir("../../assets/terraform-modules"),
		assets.ControlPlaneSource:     http.Dir("../../assets/charts/control-plane"),
		assets.ComponentsSource:       http.Dir("../../assets/charts/components"),
	}

	// Reset modification time on files, so after cloning the repository all
	// assets don't get modified.
	u := union.New(directoriesToEmbed)
	fs := modtimevfs.New(u, time.Unix(1, 0))

	options := vfsgen.Options{
		Filename:     "generated_assets.go",
		PackageName:  "assets",
		VariableName: "vfsgenAssets",
	}

	if err := vfsgen.Generate(fs, options); err != nil {
		panic(err)
	}
}
