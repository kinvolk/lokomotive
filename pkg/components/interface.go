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
	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v3/pkg/release"
)

// Component represents functionality each Lokomotive component should implement.
type Component interface {
	// LoadConfig loads specific component configuration from HCL configuration.
	// If an error occurs, HCL diagnostics are returned.
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	// RenderManifests returns a map of Kubernetes manifests in YAML format, where
	// the key is the file from which the content comes.
	RenderManifests() (*release.Release, error)
	// Metadata returns component metadata.
	Metadata() Metadata
}
