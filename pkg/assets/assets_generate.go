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
	"net/http"
	"time"

	"github.com/prometheus/alertmanager/pkg/modtimevfs"
	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"
)

func main() {
	ufs := union.New(map[string]http.FileSystem{
		"/lokomotive-kubernetes": http.Dir("../../assets/lokomotive-kubernetes"),
		"/components":            http.Dir("../../assets/components"),
	})
	fs := modtimevfs.New(ufs, time.Unix(1, 0))
	err := vfsgen.Generate(fs, vfsgen.Options{
		Filename:     "generated_assets.go",
		PackageName:  "assets",
		VariableName: "vfsgenAssets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
