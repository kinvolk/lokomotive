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

package oidc

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/hcl/v2"
)

const (
	defaultClientID      = "gangway"
	defaultUsernameClaim = "email"
	defaultGroupsClaim   = "groups"
)

// Config deals with providing OIDC related fields to the Lokomotive
// cluster for OIDC based Authentication via Dex and Gangway.
type Config struct {
	IssuerURL     string `hcl:"issuer_url,optional"`
	ClientID      string `hcl:"client_id,optional"`
	UsernameClaim string `hcl:"username_claim,optional"`
	GroupsClaim   string `hcl:"groups_claim,optional"`
}

// newConfig returns a new config with default values.
func newConfig() *Config {
	return &Config{
		ClientID:      defaultClientID,
		UsernameClaim: defaultUsernameClaim,
		GroupsClaim:   defaultGroupsClaim,
	}
}

// withDefaults returns an instance of Config which combines the user input and
// default values for fields which the user has not provided any input for.
func (c *Config) withDefaults() *Config {
	// Get new config with defaults.
	cfg := newConfig()
	// Use values provided by the user if not empty.
	if c.IssuerURL != "" {
		cfg.IssuerURL = c.IssuerURL
	}

	if c.ClientID != "" {
		cfg.ClientID = c.ClientID
	}

	if c.UsernameClaim != "" {
		cfg.UsernameClaim = c.UsernameClaim
	}

	if c.GroupsClaim != "" {
		cfg.GroupsClaim = c.GroupsClaim
	}

	return cfg
}

// ToKubeAPIServerFlags configures the Config fields after setting default values
// and returns a list of oidc flags and errors if any.
func (c *Config) ToKubeAPIServerFlags(clusterDomain string) ([]string, hcl.Diagnostics) {
	cfg := c.withDefaults()

	if cfg.IssuerURL == "" && clusterDomain == "" {
		return []string{}, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("found oidc.issuer_url and clusterdomain empty"),
			},
		}
	}

	if cfg.IssuerURL == "" {
		cfg.IssuerURL = fmt.Sprintf("https://dex.%s", clusterDomain)
	}
	// Validate the oidc configuration.
	diags := cfg.validate()
	if diags.HasErrors() {
		return []string{}, diags
	}

	oidcFlags := []string{
		fmt.Sprintf("--oidc-issuer-url=%s", cfg.IssuerURL),
		fmt.Sprintf("--oidc-client-id=%s", cfg.ClientID),
		fmt.Sprintf("--oidc-username-claim=%s", cfg.UsernameClaim),
		fmt.Sprintf("--oidc-groups-claim=%s", cfg.GroupsClaim),
	}

	return oidcFlags, hcl.Diagnostics{}
}

// validate validates the values of the oidc Config fields.
func (c *Config) validate() hcl.Diagnostics {
	var diags hcl.Diagnostics

	u, err := url.Parse(c.IssuerURL)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("invalid oidc.issuer_url: %q", err),
		})
	}

	if u.Scheme != "https" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("oidc.issuer_url scheme must be https, got: %s", u.Scheme),
		})
	}

	return diags
}
