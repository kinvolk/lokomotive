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

package local

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/backend"
)

type local struct {
	Path string `hcl:"path,optional"`
}

// init registers local as a backend.
func init() {
	backend.Register("local", NewConfig())
}

// LoadConfig loads the configuration for the local backend.
func (l *local) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, l)
}

// NewConfig returns new Local backend configuration with default values set.
//
//nolint:golint
func NewConfig() *local {
	return &local{}
}

// Render renders the Go template with local backend configuration.
func (l *local) Render() (string, error) {
	return template.Render(backendConfigTmpl, l)
}

// Validate validates the local backend configuration.
func (l *local) Validate() error {
	return nil
}
