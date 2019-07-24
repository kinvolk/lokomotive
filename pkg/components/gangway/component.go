package dex

import (
	"bytes"
	"text/template"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
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
    # The cluster name
    clusterName: {{ .ClusterName }}

    # This is the API server URL that you want users to use and configure in
    # their Kubectl. For the Heptio AWS quickstart this'll be an ELB name.
    apiServerURL: {{ .APIServerURL }}

    # The URL to send authorize requests to
    authorizeURL: {{ .AuthorizeURL }}

    # URL to get a token from
    tokenURL: {{ .TokenURL }}

    # API client ID as indicated by the identity provider
    clientID: {{ .ClientID }}

    # API client secret as indicated by the identity provider
    clientSecret: {{ .ClientSecret }}

    # Where to redirect back to. This should be a URL
    # Where gangway is reachable
    redirectURL: {{ .RedirectURL }}

    # Used to specify the scope of the requested Oauth authorization.
    scopes: ["openid", "profile", "email", "offline_access", "groups"]

    # The JWT claim to use as the Kubnernetes username
    usernameClaim: "email"

    # The JWT claim to use as the email claim
    emailClaim: "email"

    # Where to load the custom Lokomotive HTML templates from.
    # Requires the initContainer below to download the theme.
    customHTMLTemplatesDir: "/theme"
`

const deploymentManifest = `apiVersion: apps/v1beta1
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
      initContainers:
      - name: download-theme
        image: schu/alpine-git
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
          image: gcr.io/heptio-images/gangway:v3.0.0
          imagePullPolicy: Always
          command: ["gangway", "-config", "/gangway/gangway.yaml"]
          env:
            - name: GANGWAY_SESSION_SECURITY_KEY
              valueFrom:
                secretKeyRef:
                  name: gangway-session-key
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
  name: gangway-svc
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

const ingressTmpl = `apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: gangway
  namespace: gangway
  annotations:
    kubernetes.io/tls-acme: "true"
    certmanager.k8s.io/cluster-issuer: "letsencrypt-production"
    kubernetes.io/ingress.class: contour
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
          serviceName: gangway-svc
          servicePort: http
`

const secretTmpl = `
apiVersion: v1
kind: Secret
metadata:
  name: gangway-session-key
  namespace: gangway
type: Opaque
data:
  sessionkey: {{ .SessionKey }}
`

func init() {
	components.Register(name, newComponent())
}

type component struct {
	ClusterName  string `hcl:"cluster_name,attr"`
	IngressHost  string `hcl:"ingress_host,attr"`
	SessionKey   string `hcl:"session_key,attr"`
	APIServerURL string `hcl:"api_server_url,attr"`
	AuthorizeURL string `hcl:"authorize_url,attr"`
	TokenURL     string `hcl:"token_url,attr"`
	ClientID     string `hcl:"client_id,attr"`
	ClientSecret string `hcl:"client_secret,attr"`
	RedirectURL  string `hcl:"redirect_url,attr"`
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
	// TODO(schu): validate that there's at least one connector
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	tmpl, err := template.New("config-map").Parse(configMapTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}
	var configMapBuf bytes.Buffer
	if err := tmpl.Execute(&configMapBuf, c); err != nil {
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

	tmpl, err = template.New("secret").Parse(secretTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}
	var secretBuf bytes.Buffer
	if err := tmpl.Execute(&secretBuf, c); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
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

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
