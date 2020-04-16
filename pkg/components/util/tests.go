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

package util

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/config"
)

// LoadComponentFromHCLString loads a Component instance from a
// given HCL string. This function is mainly used in tests.
func LoadComponentFromHCLString(configHCL, name string) (components.Component, hcl.Diagnostics) {
	cfgMap := map[string][]byte{
		"test.lokocfg": []byte(configHCL),
	}
	varconfigMap := map[string][]byte{}

	cfg, diags := config.ParseHCLFiles(cfgMap, varconfigMap)
	if diags.HasErrors() {
		return nil, diags
	}

	components, diags := config.LoadConfiguredComponents(cfg)
	if diags.HasErrors() {
		return nil, diags
	}

	return components[name], hcl.Diagnostics{}
}
