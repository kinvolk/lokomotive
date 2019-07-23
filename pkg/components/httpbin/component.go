package httpbin

import (
	"bytes"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

const name = "httpbin"

const namespaceManifest = `apiVersion: v1
kind: Namespace
metadata:
  name: httpbin
  labels:
    name: httpbin
`

const deploymentManifest = `apiVersion: extensions/v1beta1
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
      securityContext: {}
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
    certmanager.k8s.io/cluster-issuer: "letsencrypt-production"
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

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
