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

package dex

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	internaltemplate "github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const name = "dex"

func init() { //nolint:gochecknoinits
	components.Register(name, newComponent())
}

type org struct {
	Name  string   `hcl:"name,attr" json:"name"`
	Teams []string `hcl:"teams,attr" json:"teams"`
}

type config struct {
	ClientID               string   `hcl:"client_id,attr" json:"clientID"`
	ClientSecret           string   `hcl:"client_secret,attr" json:"clientSecret"`
	Issuer                 string   `hcl:"issuer,optional" json:"issuer"`
	RedirectURI            string   `hcl:"redirect_uri,attr" json:"redirectURI"`
	TeamNameField          string   `hcl:"team_name_field,optional" json:"teamNameField"`
	Orgs                   []org    `hcl:"org,block" json:"orgs"`
	AdminEmail             string   `hcl:"admin_email,optional" json:"adminEmail"`
	HostedDomains          []string `hcl:"hosted_domains,optional" json:"hostedDomains"`
	ServiceAccountFilePath string   `json:"serviceAccountFilePath"`
}

type connector struct {
	Type   string  `hcl:"type,label" json:"type"`
	ID     string  `hcl:"id,attr" json:"id"`
	Name   string  `hcl:"name,attr" json:"name"`
	Config *config `hcl:"config,block" json:"config"`
}

type staticClient struct {
	ID           string   `hcl:"id,attr" json:"id"`
	RedirectURIs []string `hcl:"redirect_uris,attr" json:"redirectURIs"`
	Name         string   `hcl:"name,attr" json:"name"`
	Secret       string   `hcl:"secret,attr" json:"secret"`
}

type component struct {
	IngressHost              string         `hcl:"ingress_host,attr"`
	IssuerHost               string         `hcl:"issuer_host,attr"`
	Connectors               []connector    `hcl:"connector,block"`
	StaticClients            []staticClient `hcl:"static_client,block"`
	GSuiteJSONConfigPath     string         `hcl:"gsuite_json_config_path,optional"`
	CertManagerClusterIssuer string         `hcl:"certmanager_cluster_issuer,optional"`

	// Those are fields not accessible by user
	ConnectorsRaw    string
	StaticClientsRaw string
}

func newComponent() *component {
	return &component{
		CertManagerClusterIssuer: "letsencrypt-production",
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}
	// TODO(schu):
	// * validate that there's at least one connector
	// * make sure config w/o a static client does lead to valid output
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func marshalToStr(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	// Add the default path to google's connector, this is the default path
	// where the user given google suite json file will be available via a
	// secret volume, this value is also hardcoded in the deployment yaml
	for _, connc := range c.Connectors {
		if connc.Type != "google" || connc.Config == nil {
			continue
		}

		connc.Config.ServiceAccountFilePath = "/config/googleAuth.json"
	}

	connectors, err := marshalToStr(c.Connectors)
	if err != nil {
		return nil, fmt.Errorf("marshaling connectors: %w", err)
	}

	c.ConnectorsRaw = connectors

	staticClients, err := marshalToStr(c.StaticClients)
	if err != nil {
		return nil, fmt.Errorf("marshaling static clients: %w", err)
	}

	c.StaticClientsRaw = staticClients

	values, err := internaltemplate.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

	// Generate YAML for the dex deployment.
	renderedFiles, err := util.RenderChart(helmChart, name, c.Metadata().Namespace.Name, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: name,
		Namespace: k8sutil.Namespace{
			Name: name,
		},
	}
}
