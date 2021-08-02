// Copyright 2021 The Lokomotive Authors
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

// Package types contains common types used by the components. This helps in ensuring that, all the
// components expose same set of variables to the user to do similar tasks.
package types

const (
	defaultIngressClass             = "contour"
	defaultCertManagerClusterIssuer = "letsencrypt-production"
)

// Ingress is a generic object for specifying Kubernetes ingress manifest.
type Ingress struct {
	Host                     string `hcl:"host"`
	Class                    string `hcl:"class,optional"`
	CertManagerClusterIssuer string `hcl:"certmanager_cluster_issuer,optional"`
}

// SetDefaults sets default values for Ingress object only when user has not provided any
// information.
func (ing *Ingress) SetDefaults() {
	if ing.Class == "" {
		ing.Class = defaultIngressClass
	}

	if ing.CertManagerClusterIssuer == "" {
		ing.CertManagerClusterIssuer = defaultCertManagerClusterIssuer
	}
}
