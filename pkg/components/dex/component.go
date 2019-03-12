package dex

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

const name = "dex"

const namespaceManifest = `apiVersion: v1
kind: Namespace
metadata:
  name: dex
  labels:
    name: dex
`

const serviceManifest = `apiVersion: v1
kind: Service
metadata:
  name: dex
  namespace: dex
spec:
  ports:
  - name: dex
    port: 5556
    protocol: TCP
    targetPort: 5556
  selector:
    app: dex
`

const serviceAccountManifest = `apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: dex
  name: dex
  namespace: dex
`

const clusterRoleManifest = `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: dex
rules:
- apiGroups: ["dex.coreos.com"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["create"]
`

const clusterRoleBindingManifest = `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: dex
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: dex
subjects:
- kind: ServiceAccount
  name: dex
  namespace: dex
`

const deploymentManifest = `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: dex
  name: dex
  namespace: dex
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: dex
    spec:
      serviceAccountName: dex
      initContainers:
      - name: download-theme
        image: schu/alpine-git
        command:
         - git
         - clone
         - "https://github.com/kinvolk/dex-theme.git"
         - /theme
        volumeMounts:
        - name: theme
          mountPath: /theme/
      containers:
      - image: quay.io/dexidp/dex:v2.15.0
        name: dex
        command: ["/usr/local/bin/dex", "serve", "/etc/dex/cfg/config.yaml"]
        ports:
        - name: https
          containerPort: 5556
        volumeMounts:
        - name: config
          mountPath: /etc/dex/cfg
        - mountPath: /web/themes/custom/
          name: theme
      volumes:
      - name: config
        configMap:
          name: dex
          items:
          - key: config.yaml
            path: config.yaml
      - name: theme
        emptyDir: {}
`

const configMapTmpl = `apiVersion: v1
kind: ConfigMap
metadata:
  name: dex
  namespace: dex
data:
  config.yaml: |
    issuer: {{ .IssuerHost }}
    storage:
      type: kubernetes
      config:
        inCluster: true
    web:
      http: 0.0.0.0:5556
    frontend:
      theme: custom
    connectors: {{ .Connectors }}
    oauth2:
      skipApprovalScreen: true
    staticClients: {{ .StaticClients }}
`

const ingressTmpl = `apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: nginx
  namespace: dex
  annotations:
    kubernetes.io/ingress.class: "nginx"
    kubernetes.io/tls-acme: "true"
    certmanager.k8s.io/cluster-issuer: "letsencrypt-production"
spec:
  tls:
    - hosts:
       - {{ .IngressHost }}
      secretName: {{ .IngressHost }}-tls
  rules:
  - host: {{ .IngressHost }}
    http:
      paths:
      - backend:
          serviceName: dex
          servicePort: 5556
`

func init() {
	components.Register(name, newComponent())
}

type org struct {
	Name  string   `hcl:"name,attr" json:"name"`
	Teams []string `hcl:"teams,attr" json:"teams"`
}

type oidcConfig struct {
	ClientID      string  `hcl:"client_id,attr" json:"clientID"`
	ClientSecret  string  `hcl:"client_secret,attr" json:"clientSecret"`
	Issuer        *string `hcl:"issuer,attr" json:"issuer"`
	RedirectURI   string  `hcl:"redirect_uri,attr" json:"redirectURI"`
	TeamNameField *string `hcl:"team_name_field,attr" json:"teamNameField"`
	Orgs          []org   `hcl:"org,block" json:"orgs"`
}

type connector struct {
	Type   string      `hcl:"type,label" json:"type"`
	ID     string      `hcl:"id,attr" json:"id"`
	Name   string      `hcl:"name,attr" json:"name"`
	Config *oidcConfig `hcl:"config,block" json:"config"`
}

type staticClient struct {
	ID           string   `hcl:"id,attr" json:"id"`
	RedirectURIs []string `hcl:"redirect_uris,attr" json:"redirectURIs"`
	Name         string   `hcl:"name,attr" json:"name"`
	Secret       string   `hcl:"secret,attr" json:"secret"`
}

type component struct {
	IngressHost   string         `hcl:"ingress_host,attr"`
	IssuerHost    string         `hcl:"issuer_host,attr"`
	Connectors    []connector    `hcl:"connector,block"`
	StaticClients []staticClient `hcl:"static_client,block"`
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
	tmpl, err := template.New("config-map").Parse(configMapTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}
	connectorsStr, err := marshalToStr(c.Connectors)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal connectors")
	}
	staticClientsStr, err := marshalToStr(c.StaticClients)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal staticClients")
	}
	var configMap bytes.Buffer
	configMapCreds := struct {
		IssuerHost    string
		Connectors    string
		StaticClients string
	}{
		IssuerHost:    c.IssuerHost,
		Connectors:    connectorsStr,
		StaticClients: staticClientsStr,
	}
	if err := tmpl.Execute(&configMap, configMapCreds); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}

	tmpl, err = template.New("ingress").Parse(ingressTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}
	var ingressBuf bytes.Buffer
	if err := tmpl.Execute(&ingressBuf, c); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}

	return map[string]string{
		"namespace.yml":            namespaceManifest,
		"service.yml":              serviceManifest,
		"service-account.yml":      serviceAccountManifest,
		"cluster-role.yml":         clusterRoleManifest,
		"cluster-role-binding.yml": clusterRoleBindingManifest,
		"deployment.yml":           deploymentManifest,
		"config-map.yml":           configMap.String(),
		"ingress.yml":              ingressBuf.String(),
	}, nil
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
