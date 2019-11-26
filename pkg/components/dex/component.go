package dex

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
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

const deploymentManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: dex
  name: dex
  namespace: dex
spec:
  selector:
    matchLabels:
      app: dex
  replicas: 3
  template:
    metadata:
      labels:
        app: dex
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        runAsGroup: 65534
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
      # This image is built from PR that adds support of google connector to
      # dex https://github.com/dexidp/dex/pull/1185. Once this PR is merged
      # please use the official relased image.
      - image: quay.io/kinvolk/dex:google_connector
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
        - name: gsuite-auth
          mountPath: /config/
      volumes:
      - name: config
        configMap:
          name: dex
          items:
          - key: config.yaml
            path: config.yaml
      - name: theme
        emptyDir: {}
      - name: gsuite-auth
        secret:
          secretName: gsuite-auth
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
    connectors: {{ .ConnectorsRaw }}
    oauth2:
      skipApprovalScreen: true
    staticClients: {{ .StaticClientsRaw }}
`

const ingressTmpl = `apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: dex
  namespace: dex
  annotations:
    kubernetes.io/ingress.class: contour
    kubernetes.io/tls-acme: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-production"
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

const secretTmpl = `kind: Secret
apiVersion: v1
metadata:
  name: gsuite-auth
  namespace: dex
data:
  googleAuth.json: {{ .SecretData }}
`

func init() {
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
	IngressHost          string         `hcl:"ingress_host,attr"`
	IssuerHost           string         `hcl:"issuer_host,attr"`
	Connectors           []connector    `hcl:"connector,block"`
	StaticClients        []staticClient `hcl:"static_client,block"`
	GSuiteJSONConfigPath string         `hcl:"gsuite_json_config_path,optional"`

	// Those are fields not accessible by user
	ConnectorsRaw    string
	StaticClientsRaw string
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
		return nil, errors.Wrap(err, "failed to marshal connectors")
	}
	c.ConnectorsRaw = connectors

	staticClients, err := marshalToStr(c.StaticClients)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal staticClients")
	}
	c.StaticClientsRaw = staticClients

	configMap, err := util.RenderTemplate(configMapTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}

	ingressBuf, err := util.RenderTemplate(ingressTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}

	// based on the user's input of gsuite file path the secret will be created.
	// If user is not using google connector then this will just be an empty
	// output, but we create it anyway so we can keep using same deployment
	// manifest for google connector and others.
	secretManifest, err := createSecretManifest(c.GSuiteJSONConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "can't create secret from gsuite json file")
	}

	return map[string]string{
		"namespace.yml":            namespaceManifest,
		"service.yml":              serviceManifest,
		"service-account.yml":      serviceAccountManifest,
		"cluster-role.yml":         clusterRoleManifest,
		"cluster-role-binding.yml": clusterRoleBindingManifest,
		"secret.yml":               secretManifest,
		"deployment.yml":           deploymentManifest,
		"config-map.yml":           configMap,
		"ingress.yml":              ingressBuf,
	}, nil
}

func createSecretManifest(path string) (string, error) {

	// Takes in the raw data and returns a Kubernetes Secret config
	generateSecret := func(data []byte) (string, error) {
		encodedData := b64.StdEncoding.EncodeToString(data)

		secretTmplData := struct {
			SecretData string
		}{
			SecretData: encodedData,
		}
		tmpl, err := template.New("secret").Parse(secretTmpl)
		if err != nil {
			return "", err
		}
		var secret bytes.Buffer
		if err := tmpl.Execute(&secret, secretTmplData); err != nil {
			return "", err
		}
		return secret.String(), nil
	}

	// if user is not using google connector then user won't provide the file
	// path hence create secret with empty value
	if path == "" {
		return generateSecret([]byte(""))
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return generateSecret(data)
}
