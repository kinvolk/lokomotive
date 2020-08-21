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

package s3

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/internal/template"
)

// Config represents the configuration of an S3 backend.
type Config struct {
	Bucket        string `hcl:"bucket"`
	Key           string `hcl:"key"`
	Region        string `hcl:"region"`
	AWSCredsPath  string `hcl:"aws_creds_path,optional"`
	DynamoDBTable string `hcl:"dynamodb_table,optional"`
}

// validate returns an error if the Config is invalid.
func (c *Config) validate() error {
	if c.Bucket == "" {
		return fmt.Errorf("bucket cannot be empty")
	}

	if c.Key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if c.Region == "" {
		return fmt.Errorf("region cannot be empty")
	}

	return nil
}

// NewConfig creates a new Config and returns a pointer to it as well as any HCL diagnostics.
func NewConfig(b *hcl.Body, ctx *hcl.EvalContext) (*Config, hcl.Diagnostics) {
	diags := hcl.Diagnostics{}

	c := &Config{}

	if b == nil {
		return nil, diags
	}

	if d := gohcl.DecodeBody(*b, ctx, c); len(d) != 0 {
		diags = append(diags, d...)
		return nil, diags
	}

	return c, diags
}

// Backend implements the Backend interface for an S3 backend.
type Backend struct {
	config *Config
	// A string containing the rendered Terraform code of the backend.
	rendered string
}

func (b *Backend) String() string {
	return b.rendered
}

// NewBackend constructs a Backend based on the provided config and returns a pointer to it.
func NewBackend(c *Config) (*Backend, error) {
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("validating backend config: %w", err)
	}

	rendered, err := template.Render(backendConfigTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering backend: %v", err)
	}

	return &Backend{config: c, rendered: rendered}, nil
}
