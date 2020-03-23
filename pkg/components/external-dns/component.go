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

package externaldns

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/pkg/errors"
)

const name = "external-dns"

// TODO Currently supporting only AWS Route53. Provide better conditional templates
// when multiple provider support is added.
const chartValuesTmpl = `
provider: aws
{{- if .Sources }}
sources:
  {{ range .Sources -}}
  - {{.}}
  {{ end }}
{{ end }}
{{- if .AwsConfig -}}
aws:
  credentials:
    secretKey: "{{ .AwsConfig.SecretAccessKey }}"
    accessKey: "{{ .AwsConfig.AccessKeyID }}"
  zoneType: {{ .AwsConfig.ZoneType }}
txtOwnerId: {{ .OwnerID }}
{{- end }}
policy: {{ .Policy }}
replicas: 3
`

func init() {
	components.Register(name, newComponent())
}

// AwsConfig provides configuration for AWS Route53 DNS.
type AwsConfig struct {
	ZoneID          string `hcl:"zone_id"`
	ZoneType        string `hcl:"zone_type,optional"`
	AccessKeyID     string `hcl:"aws_access_key_id,optional"`
	SecretAccessKey string `hcl:"aws_secret_access_key,optional"`
}

type component struct {
	// Once we support more providers, we should add additional field called Provider.
	Sources   []string  `hcl:"sources,optional"`
	Namespace string    `hcl:"namespace,optional"`
	Metrics   bool      `hcl:"metrics,optional"`
	Policy    string    `hcl:"policy,optional"`
	AwsConfig AwsConfig `hcl:"aws,block"`
	OwnerID   string    `hcl:"owner_id"`
}

func newComponent() *component {
	return &component{
		Namespace: "external-dns",
		Sources:   []string{"service"},
		AwsConfig: AwsConfig{
			ZoneType: "public",
		},
		Policy:  "upsert-only",
		Metrics: false,
	}
}

// LoadConfig loads the component config.
func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	return gohcl.DecodeBody(*configBody, evalContext, c)
}

// RenderManifests renders the helm chart templates with values provided.
func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := util.LoadChartFromAssets(fmt.Sprintf("/components/%s/manifests", name))
	if err != nil {
		return nil, errors.Wrap(err, "load chart from assets")
	}

	// Get the aws credentials from environment variable if not provided in the config.
	if c.AwsConfig.AccessKeyID == "" {
		accessKeyID, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
		if !ok || accessKeyID == "" {
			return nil, errors.New("AWS Credentials not found.")
		}
		c.AwsConfig.AccessKeyID = accessKeyID
	}

	if c.AwsConfig.SecretAccessKey == "" {
		secretAccessKey, ok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
		if !ok || secretAccessKey == "" {
			return nil, errors.New("AWS Credentials not found.")
		}
		c.AwsConfig.SecretAccessKey = secretAccessKey
	}

	values, err := util.RenderTemplate(chartValuesTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render chart values template")
	}

	renderedFiles, err := util.RenderChart(helmChart, name, c.Namespace, values)
	if err != nil {
		return nil, errors.Wrap(err, "render chart")
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: c.Namespace,
	}
}
