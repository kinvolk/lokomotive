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
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/pkg/errors"
)

type Config struct {
	Bucket        string `hcl:"bucket"`
	Key           string `hcl:"key"`
	Region        string `hcl:"region,optional"`
	AWSCredsPath  string `hcl:"aws_creds_path,optional"`
	DynamoDBTable string `hcl:"dynamodb_table,optional"`
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

type Backend struct {
	config *Config
	// A string containing the rendered Terraform code of the backend.
	rendered string
}

func (b *Backend) String() string {
	return b.rendered
}

func (b *Backend) Validate() error {
	if b.config.AWSCredsPath == "" && os.Getenv("AWS_SHARED_CREDENTIALS_FILE") == "" {
		if b.config.Region == "" && os.Getenv("AWS_DEFAULT_REGION") == "" {
			return errors.New("no region specified")
		}
	}

	return nil
}

// NewBackend constructs a Backend based on the provided config and returns a pointer to it.
func NewBackend(c *Config) (*Backend, error) {
	if c.Bucket == "" {
		return nil, errors.New("no bucket specified")
	}

	if c.Key == "" {
		return nil, errors.New("no key specified")
	}

	rendered, err := template.Render(backendConfigTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering backend: %v", err)
	}

	return &Backend{config: c, rendered: rendered}, nil
}
