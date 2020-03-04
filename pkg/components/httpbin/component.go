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

package httpbin

import (
	"bytes"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/components"
)

const name = "httpbin"

const namespaceManifest = `apiVersion: v1
kind: Namespace
metadata:
  name: httpbin
  labels:
    name: httpbin
`

const deploymentManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: httpbin
  name: httpbin
  namespace: httpbin
spec:
  replicas: 1
  selector:
    matchLabels:
      app: httpbin
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: httpbin
    spec:
      containers:
      - image: docker.io/kennethreitz/httpbin
        name: httpbin
        ports:
        - containerPort: 8080
          name: http
        command: ["gunicorn"]
        args: ["-b", "0.0.0.0:8080", "httpbin:app"]
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        runAsGroup: 65534
      terminationGracePeriodSeconds: 30
`

const serviceManifest = `apiVersion: v1
kind: Service
metadata:
  name: httpbin
  namespace: httpbin
spec:
  ports:
  - port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    app: httpbin
`

const ingressTmpl = `apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: httpbin
  namespace: httpbin
  annotations:
    kubernetes.io/tls-acme: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-production"
    kubernetes.io/ingress.class: contour
spec:
  tls:
  - secretName: {{ .IngressHost }}-tls
    hosts:
    - {{ .IngressHost }}
  rules:
  - host: {{ .IngressHost }}
    http:
      paths:
      - backend:
          serviceName: httpbin
          servicePort: 8080
`

func init() {
	components.Register(name, newComponent())
}

type component struct {
	IngressHost string `hcl:"ingress_host,attr"`
}

func newComponent() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	tmpl, err := template.New("ingress").Parse(ingressTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}
	return map[string]string{
		"namespace.yml":  namespaceManifest,
		"deployment.yml": deploymentManifest,
		"service.yml":    serviceManifest,
		"ingress.yml":    buf.String(),
	}, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: name,
	}
}
