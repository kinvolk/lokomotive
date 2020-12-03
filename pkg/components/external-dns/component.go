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

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents ExternalDNS component name as it should be referenced in function calls
	// and in configuration.
	Name = "external-dns"

	// TODO Currently supporting only AWS Route53. Provide better conditional templates
	// when multiple provider support is added.
	chartValuesTmpl = `
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

{{ if .ServiceMonitor }}
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    namespace: {{ .Namespace }}
    selector:
      release: prometheus-operator
{{ end }}
`
)

// AwsConfig provides configuration for AWS Route53 DNS.
type AwsConfig struct {
	ZoneID          string `hcl:"zone_id"`
	ZoneType        string `hcl:"zone_type,optional"`
	AccessKeyID     string `hcl:"aws_access_key_id,optional"`
	SecretAccessKey string `hcl:"aws_secret_access_key,optional"`
}

type component struct {
	// Once we support more providers, we should add additional field called Provider.
	Sources        []string  `hcl:"sources,optional"`
	Namespace      string    `hcl:"namespace,optional"`
	Metrics        bool      `hcl:"metrics,optional"`
	Policy         string    `hcl:"policy,optional"`
	ServiceMonitor bool      `hcl:"service_monitor,optional"`
	AwsConfig      AwsConfig `hcl:"aws,block"`
	OwnerID        string    `hcl:"owner_id"`
}

// NewConfig returns new ExternalDNS component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{
		Namespace: "external-dns",
		Sources:   []string{"ingress"},
		AwsConfig: AwsConfig{
			ZoneType: "public",
		},
		Policy:         "upsert-only",
		Metrics:        false,
		ServiceMonitor: false,
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
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	// Get the aws credentials from environment variable if not provided in the config.
	if c.AwsConfig.AccessKeyID == "" {
		accessKeyID, ok := os.LookupEnv("AWS_ACCESS_KEY_ID")
		if !ok || accessKeyID == "" {
			return nil, fmt.Errorf("AWS access key ID not found")
		}
		c.AwsConfig.AccessKeyID = accessKeyID
	}

	if c.AwsConfig.SecretAccessKey == "" {
		secretAccessKey, ok := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
		if !ok || secretAccessKey == "" {
			return nil, fmt.Errorf("AWS secret access key not found")
		}
		c.AwsConfig.SecretAccessKey = secretAccessKey
	}

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering chart values template: %w", err)
	}

	renderedFiles, err := util.RenderChart(helmChart, Name, c.Namespace, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: c.Namespace,
		},
	}
}
