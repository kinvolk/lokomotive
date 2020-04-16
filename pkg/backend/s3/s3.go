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
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/util"
)

type s3 struct {
	Bucket        string `hcl:"bucket"`
	Key           string `hcl:"key"`
	Region        string `hcl:"region,optional"`
	AWSCredsPath  string `hcl:"aws_creds_path,optional"`
	DynamoDBTable string `hcl:"dynamodb_table,optional"`
}

// init registers s3 as a backend.
func init() {
	backend.Register("s3", NewS3Backend())
}

// LoadConfig loads the configuration for the s3 backend.
func (s *s3) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, s)
}

func NewS3Backend() *s3 {
	return &s3{}
}

// Render renders the Go template with s3 backend configuration.
func (s *s3) Render() (string, error) {
	return util.RenderTemplate(backendConfigTmpl, s)
}

// Validate validates the s3 backend configuration.
func (s *s3) Validate() error {
	if s.Bucket == "" {
		return errors.Errorf("no bucket specified")
	}

	if s.Key == "" {
		return errors.Errorf("no key specified")
	}

	if s.AWSCredsPath == "" && os.Getenv("AWS_SHARED_CREDENTIALS_FILE") == "" {
		if s.Region == "" && os.Getenv("AWS_DEFAULT_REGION") == "" {
			return errors.Errorf("no region specified: use Region field in backend configuration or AWS_DEFAULT_REGION environment variable")
		}
	}

	return nil
}
