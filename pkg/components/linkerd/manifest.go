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
global:
  identityTrustAnchorsPEM: |
{{ .CA }}
  proxy:
    resources:
      cpu:
        limit: "1"
        request: 100m
      memory:
        limit: 250Mi
        request: 20Mi

identity:
  issuer:
    crtExpiry: {{ .Expiry }}
    tls:
      crtPEM: |
{{ .Cert }}
      keyPEM: |
{{ .Key }}

# Following is from values-ha.yaml which contains the values needed to enable HA mode.
enablePodAntiAffinity: true

# controller configuration
controllerReplicas: {{.ControllerReplicas}}
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
prometheusResources:
  cpu:
    limit: "4"
    request: 300m
  memory:
    limit: 8192Mi
    request: 300Mi

# proxy injector configuration
proxyInjectorResources: *controller_resources

# NOTE: This has been set to 'Ignore' while upstream suggests to set it to 'Fail'.
# When set to 'Fail' the installation of this component is unreliable because the Webhook configs
# are installed first which blocks installation of any pod after that because it introduces
# deadlock.
#
# Since the webhook config is installed first the apiserver is fails any pod request that is not
# validated by webhook pod. But webhook pod is not up yet because webhook config is not validated.
# So the best way to break this cycle is to install the webhook pods first and then install the
# webhook config.
webhookFailurePolicy: Ignore

# service profile validator configuration
spValidatorResources: *controller_resources

# tap configuration
tapResources: *controller_resources

# web configuration
webResources: *controller_resources
`
