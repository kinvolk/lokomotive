package metallb

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

const name = "metallb"

const namespace = `
apiVersion: v1
kind: Namespace
metadata:
  name: metallb-system
  labels:
    app: metallb
`

const serviceAccountController = `
apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: metallb-system
  name: controller
  labels:
    app: metallb
`

const serviceAccountSpeaker = `
apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: metallb-system
  name: speaker
  labels:
    app: metallb
`
const clusterRoleMetallbSystemController = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: metallb-system:controller
  labels:
    app: metallb
rules:
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list", "watch", "update"]
- apiGroups: [""]
  resources: ["services/status"]
  verbs: ["update"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]
`

// Note: Diversion from upstream
// This ClusterRole has added rule to use the Pod Security Policy
const clusterRoleMetallbSystemSpeaker = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: metallb-system:speaker
  labels:
    app: metallb
rules:
- apiGroups: [""]
  resources: ["services", "endpoints", "nodes"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions"]
  resources: ["podsecuritypolicies"]
  resourceNames: ["metallb-speaker"]
  verbs: ["use"]
`

const roleConfigWatcher = `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: metallb-system
  name: config-watcher
  labels:
    app: metallb
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create"]
`

const clusterRoleBindingMetallbSystemController = `
## Role bindings
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: metallb-system:controller
  labels:
    app: metallb
subjects:
- kind: ServiceAccount
  name: controller
  namespace: metallb-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metallb-system:controller
`

const clusterRoleBindingMetallbSystemSpeaker = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: metallb-system:speaker
  labels:
    app: metallb
subjects:
- kind: ServiceAccount
  name: speaker
  namespace: metallb-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metallb-system:speaker
`

const roleBindingConfigWatcher = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: metallb-system
  name: config-watcher
  labels:
    app: metallb
subjects:
- kind: ServiceAccount
  name: controller
- kind: ServiceAccount
  name: speaker
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: config-watcher
`

const deploymentController = `
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  namespace: metallb-system
  name: controller
  labels:
    app: metallb
    component: controller
spec:
  revisionHistoryLimit: 3
  selector:
    matchLabels:
      app: metallb
      component: controller
  template:
    metadata:
      labels:
        app: metallb
        component: controller
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "7472"
    spec:
      {{- if .ControllerNodeSelectors }}
      nodeSelector:
        {{- range $key, $value := .ControllerNodeSelectors }}
        {{ $key }}: "{{ $value }}"
        {{- end }}
      {{- end }}
      serviceAccountName: controller
      terminationGracePeriodSeconds: 0
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534 # nobody
      containers:
      - name: controller
        image: metallb/controller:v0.7.3
        imagePullPolicy: IfNotPresent
        args:
        - --port=7472
        - --config=config
        ports:
        - name: monitoring
          containerPort: 7472
        resources:
          limits:
            cpu: 100m
            memory: 100Mi

        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - all
          readOnlyRootFilesystem: true
      {{- if .ControllerTolerations }}
      tolerations: {{ .ControllerTolerations }}
      {{- end }}
`

const daemonsetSpeaker = `
apiVersion: apps/v1beta2
kind: DaemonSet
metadata:
  namespace: metallb-system
  name: speaker
  labels:
    app: metallb
    component: speaker
spec:
  selector:
    matchLabels:
      app: metallb
      component: speaker
  template:
    metadata:
      labels:
        app: metallb
        component: speaker
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "7472"
    spec:
      {{- if .SpeakerNodeSelectors }}
      nodeSelector:
        {{- range $key, $value := .SpeakerNodeSelectors }}
        {{ $key }}: "{{ $value }}"
        {{- end }}
      {{- end }}
      serviceAccountName: speaker
      terminationGracePeriodSeconds: 0
      hostNetwork: true
      containers:
      - name: speaker
        image: metallb/speaker:v0.7.3
        imagePullPolicy: IfNotPresent
        args:
        - --port=7472
        - --config=config
        env:
        - name: METALLB_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        ports:
        - name: monitoring
          containerPort: 7472
        resources:
          limits:
            cpu: 100m
            memory: 100Mi

        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - all
            add:
            - net_raw
      {{- if .SpeakerTolerations }}
      tolerations: {{ .SpeakerTolerations }}
      {{- end }}
`

// Note: Diversion from upstream
// This config was created specifically for clusters that has Pod Security
// Policy enabled on them
const pspMetallbSpeaker = `
apiVersion: extensions/v1beta1
kind: PodSecurityPolicy
metadata:
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: docker/default
    seccomp.security.alpha.kubernetes.io/defaultProfileName: docker/default
  name: metallb-speaker
spec:
  hostNetwork: true
  hostPorts:
  - min: 7472
    max: 7472
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  allowedCapabilities:
  - net_raw
  requiredDropCapabilities:
  - all
  seLinux:
    rule: RunAsAny
  fsGroup:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  runAsUser:
    # Require the container to run without root privileges.
    rule: RunAsAny
  supplementalGroups:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  volumes:
  - secret
`

func init() {
	components.Register(name, newComponent())
}

type component struct {
	ControllerNodeSelectors map[string]string `hcl:"controller_node_selectors,optional"`
	SpeakerNodeSelectors    map[string]string `hcl:"speaker_node_selectors,optional"`
	ControllerTolerations   []toleration      `hcl:"controller_toleration,block"`
	SpeakerTolerations      []toleration      `hcl:"speaker_toleration,block"`
}

type toleration struct {
	Key               string `hcl:"key,optional" json:"key,omitempty"`
	Effect            string `hcl:"effect,optional" json:"effect,omitempty"`
	Operator          string `hcl:"operator,optional" json:"operator,omitempty"`
	Value             string `hcl:"value,optional" json:"value,omitempty"`
	TolerationSeconds string `hcl:"toleration_seconds,optional" json:"toleration_seconds,omitempty"`
}

func newComponent() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

// renderTolerations takes a list of tolerations.
// It returns a json string and an error if any.
func renderTolerations(t []toleration) (string, error) {
	if len(t) == 0 {
		return "", nil
	}

	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *component) RenderManifests() (map[string]string, error) {
	st, err := renderTolerations(c.SpeakerTolerations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal speaker tolerations")
	}

	ct, err := renderTolerations(c.ControllerTolerations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal controller tolerations")
	}

	cv := struct {
		ControllerNodeSelectors map[string]string
		SpeakerNodeSelectors    map[string]string
		SpeakerTolerations      string
		ControllerTolerations   string
	}{
		ControllerNodeSelectors: c.ControllerNodeSelectors,
		SpeakerNodeSelectors:    c.SpeakerNodeSelectors,
		SpeakerTolerations:      st,
		ControllerTolerations:   ct,
	}

	tmpl, err := template.New("controller").Parse(deploymentController)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}

	var controllerBuf bytes.Buffer
	if err := tmpl.Execute(&controllerBuf, cv); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}

	tmpl, err = template.New("speaker").Parse(daemonsetSpeaker)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}

	var speakerBuf bytes.Buffer
	if err := tmpl.Execute(&speakerBuf, cv); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}

	return map[string]string{
		"namespace.yaml":                                    namespace,
		"service-account-controller.yaml":                   serviceAccountController,
		"service-account-speaker.yaml":                      serviceAccountSpeaker,
		"clusterrole-metallb-system-controller.yaml":        clusterRoleMetallbSystemController,
		"clusterrole-metallb-System-speaker.yaml":           clusterRoleMetallbSystemSpeaker,
		"role-config-watcher.yaml":                          roleConfigWatcher,
		"clusterrolebinding-metallb-system-controller.yaml": clusterRoleBindingMetallbSystemController,
		"clusterrolebinding-metallb-system-speaker.yaml":    clusterRoleBindingMetallbSystemSpeaker,
		"rolebinding-config-watcher.yaml":                   roleBindingConfigWatcher,
		"deployment-controller.yaml":                        controllerBuf.String(),
		"daemonset-speaker.yaml":                            speakerBuf.String(),
		"psp-metallb-speaker.yaml":                          pspMetallbSpeaker,
	}, nil
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
