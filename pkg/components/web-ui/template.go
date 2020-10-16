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

package webui

const chartValuesTmpl = `
nameOverride: web-ui
fullnameOverride: web-ui
nodeSelector:
  beta.kubernetes.io/os: linux
image:
  repository: quay.io/kinvolk/lokomotive-web-ui
  tag: v0.1.1
  pullPolicy: Always
securityContext:
  capabilities:
    drop:
    - ALL
  runAsNonRoot: true
  runAsUser: 1000
{{- if .Ingress }}
ingress:
  enabled: true
  hosts:
  - host: {{ .Ingress.Host }}
    paths:
    - /
  tls:
  - secretName: {{ .Ingress.Host }}-tls
    hosts:
    - {{ .Ingress.Host }}
  annotations:
    kubernetes.io/ingress.class: {{ .Ingress.Class }}
    cert-manager.io/cluster-issuer: {{ .Ingress.CertManagerClusterIssuer }}
    contour.heptio.com/websocket-routes: "/"
{{- end }}
{{- if .OIDC }}
oidc:
  clientID: {{ .OIDC.ClientID }}
  clientSecret: {{ .OIDC.ClientSecret }}
  issuerURL: {{ .OIDC.IssuerURL }}
{{- end }}
`
