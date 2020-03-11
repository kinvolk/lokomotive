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

package metallb

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
- apiGroups:
  - ''
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - ''
  resources:
  - services/status
  verbs:
  - update
- apiGroups:
  - ''
  resources:
  - events
  verbs:
  - create
  - patch
`

// Note: Diversion from upstream.
// This ClusterRole has added rule to use the Pod Security Policy.
const clusterRoleMetallbSystemSpeaker = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: metallb-system:speaker
  labels:
    app: metallb
rules:
- apiGroups:
  - ''
  resources:
  - services
  - endpoints
  - nodes
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ''
  resources:
  - events
  verbs:
  - create
  - patch
- apiGroups:
  - extensions
  resourceNames:
  - speaker
  resources:
  - podsecuritypolicies
  verbs:
  - use
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
- apiGroups:
  - ''
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
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
apiVersion: apps/v1
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
        image: quay.io/kinvolk/metallb-controller:v0.8.3-2-gf653773b
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
      {{- if .ControllerTolerationsJSON }}
      tolerations: {{ .ControllerTolerationsJSON }}
      {{- end }}
`

const daemonsetSpeaker = `
apiVersion: apps/v1
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
        image: quay.io/kinvolk/metallb-speaker:v0.8.3-2-gf653773b
        imagePullPolicy: IfNotPresent
        args:
        - --port=7472
        - --config=config
        env:
        - name: METALLB_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: METALLB_HOST
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
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
            - ALL
            add:
            - NET_ADMIN
            - NET_RAW
            - SYS_ADMIN
      {{- if .SpeakerTolerationsJSON }}
      tolerations: {{ .SpeakerTolerationsJSON }}
      {{- end }}
`

// Note: Diversion from upstream.
// This config was created specifically for clusters that has Pod Security Policy enabled on them.
const pspMetallbSpeaker = `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: docker/default
    seccomp.security.alpha.kubernetes.io/defaultProfileName: docker/default
  name: speaker
  labels:
    app: metallb
spec:
  hostNetwork: true
  hostPorts:
  - min: 7472
    max: 7472
  allowPrivilegeEscalation: false
  allowedCapabilities:
  - NET_RAW
  - NET_ADMIN
  - SYS_ADMIN
  seLinux:
    rule: RunAsAny
  fsGroup:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  privileged: true
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

// Needed by ServiceMonitor
const service = `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: metallb
  name: metallb-metrics
  namespace: metallb-system
spec:
  ports:
  - port: 7472
    name: metallb-metrics
  selector:
    app: metallb
`

// For autodiscovery by Prometheus operator
const serviceMonitor = `
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app: metallb
    release: prometheus-operator
  name: metallb
  namespace: metallb-system
spec:
  endpoints:
  - port: metallb-metrics
  namespaceSelector:
    matchNames:
    - metallb-system
  selector:
    matchLabels:
      app: metallb
`

const configMap = `
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    {{- range $k, $v := .AddressPools }}
    - name: {{ $k }}
      protocol: bgp
      addresses:
      {{- range $a := $v }}
      - {{ $a }}
      {{- end }}
    {{- end }}
`

const grafanaDashboard = `
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: grafana
    grafana_dashboard: "true"
  name: grafana-dashs
  namespace: metallb-system
data:
  metallb.json: |
    {
      "annotations": {
        "list": [
          {
            "builtIn": 1,
            "datasource": "-- Grafana --",
            "enable": true,
            "hide": true,
            "iconColor": "rgba(0, 211, 255, 1)",
            "name": "Annotations & Alerts",
            "type": "dashboard"
          }
        ]
      },
      "editable": true,
      "gnetId": null,
      "graphTooltip": 0,
      "links": [],
      "panels": [
        {
          "aliasColors": {},
          "bars": false,
          "dashLength": 10,
          "dashes": false,
          "datasource": "Prometheus",
          "description": "Rate of Kubernetes object updates that failed for some reason.",
          "fill": 1,
          "fillGradient": 0,
          "gridPos": {
            "h": 8,
            "w": 10,
            "x": 0,
            "y": 0
          },
          "hiddenSeries": false,
          "id": 6,
          "legend": {
            "avg": false,
            "current": false,
            "max": false,
            "min": false,
            "show": true,
            "total": false,
            "values": false
          },
          "lines": true,
          "linewidth": 1,
          "nullPointMode": "null",
          "options": {
            "dataLinks": []
          },
          "percentage": false,
          "pointradius": 2,
          "points": false,
          "renderer": "flot",
          "seriesOverrides": [],
          "spaceLength": 10,
          "stack": false,
          "steppedLine": false,
          "targets": [
            {
              "expr": "rate(metallb_k8s_client_update_errors_total[1m])",
              "legendFormat": "pod={{pod}}",
              "refId": "A"
            }
          ],
          "thresholds": [],
          "timeFrom": null,
          "timeRegions": [],
          "timeShift": null,
          "title": "Client Update Errors",
          "tooltip": {
            "shared": true,
            "sort": 0,
            "value_type": "individual"
          },
          "type": "graph",
          "xaxis": {
            "buckets": null,
            "mode": "time",
            "name": null,
            "show": true,
            "values": []
          },
          "yaxes": [
            {
              "format": "short",
              "label": null,
              "logBase": 1,
              "max": null,
              "min": null,
              "show": true
            },
            {
              "format": "short",
              "label": null,
              "logBase": 1,
              "max": null,
              "min": null,
              "show": true
            }
          ],
          "yaxis": {
            "align": false,
            "alignLevel": null
          }
        },
        {
          "aliasColors": {},
          "bars": false,
          "dashLength": 10,
          "dashes": false,
          "datasource": "Prometheus",
          "description": "Rate of Kubernetes object updates that have been processed.",
          "fill": 1,
          "fillGradient": 0,
          "gridPos": {
            "h": 8,
            "w": 10,
            "x": 10,
            "y": 0
          },
          "hiddenSeries": false,
          "id": 8,
          "legend": {
            "avg": false,
            "current": false,
            "max": false,
            "min": false,
            "show": true,
            "total": false,
            "values": false
          },
          "lines": true,
          "linewidth": 1,
          "nullPointMode": "null",
          "options": {
            "dataLinks": []
          },
          "percentage": false,
          "pointradius": 2,
          "points": false,
          "renderer": "flot",
          "seriesOverrides": [],
          "spaceLength": 10,
          "stack": false,
          "steppedLine": false,
          "targets": [
            {
              "expr": "rate(metallb_k8s_client_updates_total[1m])",
              "legendFormat": "pod={{pod}}",
              "refId": "A"
            }
          ],
          "thresholds": [],
          "timeFrom": null,
          "timeRegions": [],
          "timeShift": null,
          "title": "Kubernetes Client Updates",
          "tooltip": {
            "shared": true,
            "sort": 0,
            "value_type": "individual"
          },
          "type": "graph",
          "xaxis": {
            "buckets": null,
            "mode": "time",
            "name": null,
            "show": true,
            "values": []
          },
          "yaxes": [
            {
              "format": "short",
              "label": null,
              "logBase": 1,
              "max": null,
              "min": null,
              "show": true
            },
            {
              "format": "short",
              "label": null,
              "logBase": 1,
              "max": null,
              "min": null,
              "show": true
            }
          ],
          "yaxis": {
            "align": false,
            "alignLevel": null
          }
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": "Prometheus",
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "minValue": 0,
            "show": false,
            "thresholdLabels": false,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 5,
            "w": 5,
            "x": 0,
            "y": 8
          },
          "id": 4,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "full": false,
            "lineColor": "rgb(31, 120, 193)",
            "show": false,
            "ymax": null,
            "ymin": null
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "sum(metallb_allocator_addresses_in_use_total)",
              "refId": "A"
            }
          ],
          "thresholds": "",
          "timeFrom": null,
          "timeShift": null,
          "title": "Number of Public Addresses in use",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": "Prometheus",
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "minValue": 0,
            "show": false,
            "thresholdLabels": false,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 5,
            "w": 5,
            "x": 5,
            "y": 8
          },
          "id": 2,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "pluginVersion": "6.5.0",
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "full": false,
            "lineColor": "rgb(31, 120, 193)",
            "show": false,
            "ymax": null,
            "ymin": null
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "sum(metallb_bgp_session_up)",
              "refId": "A"
            }
          ],
          "thresholds": "",
          "timeFrom": null,
          "timeShift": null,
          "title": "Number of BGP Sessions",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        }
      ],
      "schemaVersion": 21,
      "style": "dark",
      "tags": [],
      "templating": {
        "list": []
      },
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "refresh_intervals": [
          "5s",
          "10s",
          "30s",
          "1m",
          "5m",
          "15m",
          "30m",
          "1h",
          "2h",
          "1d"
        ]
      },
      "timezone": "",
      "title": "MetalLB Dashboard",
      "uid": "0qwBMU_Wk",
      "version": 1
    }
`

const metallbPrometheusRule = `
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: alertmanager-rules
  namespace: metallb-system
  labels:
    release: prometheus-operator
    app: prometheus-operator
spec:
  groups:
  - name: metallb-rules
    rules:
    - alert: MetalLBNoBGPSession
      expr: metallb_bgp_session_up != 1
      for: 2m
      annotations:
        description: '{{ $labels.instance }}: MetalLB has not established a BGP session for more than 2 minutes.'
        summary: '{{ $labels.instance }}: MetalLB has not established BGP session.'
    - alert: MetalLBConfigStale
      expr: metallb_k8s_client_config_stale_bool != 0
      for: 2m
      annotations:
        description: '{{ $labels.instance }}: MetalLB instance has stale configuration.'
        summary: '{{ $labels.instance }}: MetalLB stale configuration.'
`
