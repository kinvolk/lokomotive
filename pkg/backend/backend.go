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

package backend

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	"github.com/kinvolk/lokomotive/pkg/backend/s3"
	"github.com/kinvolk/lokomotive/pkg/config"
)

const (
	// Local represents a local backend.
	Local = "local"
	// S3 represents an S3 backend.
	S3 = "s3"
)

// Backend describes a Terraform state storage location.
type Backend struct {
	Type   string
	Config interface{}
}

// New creates a new Backend from the provided config and returns a pointer to it.
func New(c *config.Config) (*Backend, hcl.Diagnostics) {
	if c == nil && c.RootConfig.Backend == nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "nil backend config",
			},
		}
	}

	backendType := c.RootConfig.Backend.Type

	var bc interface{}

	var d hcl.Diagnostics

	switch backendType {
	case Local:
		bc, d = local.NewConfig(&c.RootConfig.Backend.Config, c.EvalContext)
	case S3:
		bc, d = s3.NewConfig(&c.RootConfig.Backend.Config, c.EvalContext)
	default:
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("unknown backend type %q", backendType),
			},
		}
	}

	if d.HasErrors() {
		return nil, d
	}

	return &Backend{
		Type:   backendType,
		Config: bc,
	}, nil
}
