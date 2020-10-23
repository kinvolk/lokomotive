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

// Package linkerd has code related to deployment of istio operator component.
package linkerd

const chartValuesTmpl = `
enableMonitoring: {{.EnableMonitoring}}
global:
  identityTrustAnchorsPEM: |
{{ .Cert.CA }}

identity:
  issuer:
    crtExpiry: {{ .Cert.Expiry }}
    tls:
      crtPEM: |
{{ .Cert.Cert }}
      keyPEM: |
{{ .Cert.Key }}

# controller configuration
controllerReplicas: {{.ControllerReplicas}}
`

// Contents of the values-ha.yaml file are copied here verbatim. Necessary fields are overridden in
// `chartValuesTmpl` using user provided information.
const valuesHA = `
# This values.yaml file contains the values needed to enable HA mode.
# Usage:
#   helm install -f values.yaml -f values-ha.yaml

enablePodAntiAffinity: true

global:
  # proxy configuration
  proxy:
    resources:
      cpu:
        limit: "1"
        request: 100m
      memory:
        limit: 250Mi
        request: 20Mi

# controller configuration
controllerReplicas: 3
controllerResources: &controller_resources
  cpu: &controller_resources_cpu
    limit: "1"
    request: 100m
  memory:
    limit: 250Mi
    request: 50Mi
destinationResources: *controller_resources
publicAPIResources: *controller_resources

# identity configuration
identityResources:
  cpu: *controller_resources_cpu
  memory:
    limit: 250Mi
    request: 10Mi

# grafana configuration
grafana:
  resources:
    cpu: *controller_resources_cpu
    memory:
      limit: 1024Mi
      request: 50Mi

# heartbeat configuration
heartbeatResources: *controller_resources

# prometheus configuration
prometheus:
  resources:
    cpu:
      limit: "4"
      request: 300m
    memory:
      limit: 8192Mi
      request: 300Mi

# proxy injector configuration
proxyInjectorResources: *controller_resources

webhookFailurePolicy: Fail

# service profile validator configuration
spValidatorResources: *controller_resources

# tap configuration
tapResources: *controller_resources

# web configuration
webResources: *controller_resources
`
