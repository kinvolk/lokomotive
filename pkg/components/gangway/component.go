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
	"bytes"
	"fmt"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const name = "gangway"

const namespaceManifest = `apiVersion: v1
kind: Namespace
metadata:
  name: gangway
`

const configMapTmpl = `apiVersion: v1
kind: ConfigMap
metadata:
  name: gangway
  namespace: gangway
data:
  gangway.yaml: |
    # The address to listen on. Defaults to 0.0.0.0 to listen on all interfaces.
    # Env var: GANGWAY_HOST
    # host: 0.0.0.0

    # The port to listen on. Defaults to 8080.
    # Env var: GANGWAY_PORT
    # port: 8080

    # Should Gangway serve TLS vs. plain HTTP? Default: false
    # Env var: GANGWAY_SERVE_TLS
    # serveTLS: false

    # The public cert file (including root and intermediates) to use when serving
    # TLS.
    # Env var: GANGWAY_CERT_FILE
    # certFile: /etc/gangway/tls/tls.crt

    # The private key file when serving TLS.
    # Env var: GANGWAY_KEY_FILE
    # keyFile: /etc/gangway/tls/tls.key

    # The cluster name. Used in UI and kubectl config instructions.
    # Env var: GANGWAY_CLUSTER_NAME
    clusterName: {{ .ClusterName }}

    # OAuth2 URL to start authorization flow.
    # Env var: GANGWAY_AUTHORIZE_URL
    authorizeURL: {{ .AuthorizeURL }}

    # OAuth2 URL to obtain access tokens.
    # Env var: GANGWAY_TOKEN_URL
    tokenURL: {{ .TokenURL }}

    # Endpoint that provides user profile information [optional]. Not all providers
    # will require this.
    # Env var: GANGWAY_AUDIENCE
    audience: "https://${DNS_NAME}/userinfo"

    # Used to specify the scope of the requested Oauth authorization.
    scopes: ["openid", "profile", "email", "offline_access", "groups"]

    # Where to redirect back to. This should be a URL where gangway is reachable.
    # Typically this also needs to be registered as part of the oauth application
    # with the oAuth provider.
    # Env var: GANGWAY_REDIRECT_URL
    redirectURL: {{ .RedirectURL }}

    # API client ID as indicated by the identity provider
    # Env var: GANGWAY_CLIENT_ID
    clientID: {{ .ClientID }}

    # API client secret as indicated by the identity provider
    # Env var: GANGWAY_CLIENT_SECRET
    clientSecret: {{ .ClientSecret }}

    # Some identity providers accept an empty client secret, this
    # is not generally considered a good idea. If you have to use an
    # empty secret and accept the risks that come with that then you can
    # set this to true.
    #allowEmptyClientSecret: false

    # The JWT claim to use as the username. This is used in UI.
    # Default is "nickname". This is combined with the clusterName
    # for the "user" portion of the kubeconfig.
    # Env var: GANGWAY_USERNAME_CLAIM
    usernameClaim: "email"

    # The JWT claim to use as the email claim
    emailClaim: "email"

    # The API server endpoint used to configure kubectl
    # Env var: GANGWAY_APISERVER_URL
    apiServerURL: {{ .APIServerURL }}

    # The path to find the CA bundle for the API server. Used to configure kubectl.
    # This is typically mounted into the default location for workloads running on
    # a Kubernetes cluster and doesn't need to be set.
    # Env var: GANGWAY_CLUSTER_CA_PATH
    # cluster_ca_path: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

    # The path to a root CA to trust for self signed certificates at the Oauth2 URLs
    # Env var: GANGWAY_TRUSTED_CA_PATH
    #trustedCAPath: /cacerts/rootca.crt

    # The path gangway uses to create urls (defaults to "")
    # Env var: GANGWAY_HTTP_PATH
    #httpPath: "https://${GANGWAY_HTTP_PATH}"

    # The path to find custom HTML templates
    # Env var: GANGWAY_CUSTOM_HTTP_TEMPLATES_DIR
    customHTMLTemplatesDir: "/theme"
`

const deploymentManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: gangway
  namespace: gangway
  labels:
    app: gangway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gangway
  strategy:
  template:
    metadata:
      labels:
        app: gangway
        revision: "1"
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        runAsGroup: 65534
      initContainers:
      - name: download-theme
        image: alpine/git:1.0.7
        command:
         - git
         - clone
         - "https://github.com/kinvolk/gangway-theme.git"
         - /theme
        volumeMounts:
        - name: theme
          mountPath: /theme/
      containers:
        - name: gangway
          image: gcr.io/heptio-images/gangway:v3.2.0
          imagePullPolicy: Always
          command: ["gangway", "-config", "/gangway/gangway.yaml"]
          env:
            - name: GANGWAY_SESSION_SECURITY_KEY
              valueFrom:
                secretKeyRef:
                  name: gangway-key
                  key: sessionkey
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "200m"
              memory: "512Mi"
          volumeMounts:
            - name: gangway
              mountPath: /gangway/
            - name: theme
              mountPath: /theme/
          livenessProbe:
            httpGet:
              path: /
              port: 8080
            initialDelaySeconds: 20
            timeoutSeconds: 1
            periodSeconds: 60
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /
              port: 8080
            timeoutSeconds: 1
            periodSeconds: 10
            failureThreshold: 3
      volumes:
        - name: gangway
          configMap:
            name: gangway
        - name: theme
          emptyDir: {}
`

const serviceManifest = `
kind: Service
apiVersion: v1
metadata:
  name: gangwaysvc
  namespace: gangway
  labels:
    app: gangway
spec:
  type: ClusterIP
  ports:
    - name: "http"
      protocol: TCP
      port: 80
      targetPort: "http"
  selector:
    app: gangway
`

const ingressTmpl = `apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: gangway
  namespace: gangway
  labels:
    app.kubernetes.io/managed-by: Helm
  annotations:
    kubernetes.io/tls-acme: "true"
    cert-manager.io/cluster-issuer: {{ .CertManagerClusterIssuer }}
    kubernetes.io/ingress.class: contour
    meta.helm.sh/release-name: gangway
    meta.helm.sh/release-namespace: gangway
spec:
  tls:
  - secretName: gangway
    hosts:
    - {{ .IngressHost }}
  rules:
  - host: {{ .IngressHost }}
    http:
      paths:
      - backend:
          serviceName: gangwaysvc
          servicePort: http
`

const secretTmpl = `
apiVersion: v1
kind: Secret
metadata:
  name: gangway-key
  namespace: gangway
type: Opaque
data:
  sessionkey: {{ .SessionKey }}
`

func init() {
	components.Register(name, newComponent)
}

type component struct {
	ClusterName              string `hcl:"cluster_name,attr"`
	IngressHost              string `hcl:"ingress_host,attr"`
	SessionKey               string `hcl:"session_key,attr"`
	APIServerURL             string `hcl:"api_server_url,attr"`
	AuthorizeURL             string `hcl:"authorize_url,attr"`
	TokenURL                 string `hcl:"token_url,attr"`
	ClientID                 string `hcl:"client_id,attr"`
	ClientSecret             string `hcl:"client_secret,attr"`
	RedirectURL              string `hcl:"redirect_url,attr"`
	CertManagerClusterIssuer string `hcl:"certmanager_cluster_issuer,optional"`
}

func newComponent() components.Component {
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
	// TODO(schu): validate that there's at least one connector
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

// TODO: Convert to Helm chart.
func (c *component) RenderManifests() (map[string]string, error) {
	tmpl, err := template.New("config-map").Parse(configMapTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing ConfigMap template: %w", err)
	}
	var configMapBuf bytes.Buffer
	if err := tmpl.Execute(&configMapBuf, c); err != nil {
		return nil, fmt.Errorf("executing ConfigMap template: %w", err)
	}

	tmpl, err = template.New("ingress").Parse(ingressTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing Ingress template: %w", err)
	}
	var ingressBuf bytes.Buffer
	if err := tmpl.Execute(&ingressBuf, c); err != nil {
		return nil, fmt.Errorf("executing Ingress template: %w", err)
	}

	tmpl, err = template.New("secret").Parse(secretTmpl)
	if err != nil {
		return nil, fmt.Errorf("parsing Secret template: %w", err)
	}
	var secretBuf bytes.Buffer
	if err := tmpl.Execute(&secretBuf, c); err != nil {
		return nil, fmt.Errorf("executing Secret template: %w", err)
	}

	return map[string]string{
		"namespace.yml":  namespaceManifest,
		"config-map.yml": configMapBuf.String(),
		"deployment.yml": deploymentManifest,
		"service.yml":    serviceManifest,
		"ingress.yml":    ingressBuf.String(),
		"secret.yml":     secretBuf.String(),
	}, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: name,
		Namespace: k8sutil.Namespace{
			Name: name,
		},
	}
}
