// Copyright 2021 The Lokomotive Authors
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
  labels:
    app: metallb
  name: metallb-system:controller
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
- apiGroups:
  - policy
  resourceNames:
  - controller
  resources:
  - podsecuritypolicies
  verbs:
  - use
`

const clusterRoleMetallbSystemSpeaker = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: metallb
  name: metallb-system:speaker
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
  - policy
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
  labels:
    app: metallb
  name: config-watcher
  namespace: metallb-system
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

const rolePodLister = `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app: metallb
  name: pod-lister
  namespace: metallb-system
rules:
- apiGroups:
  - ''
  resources:
  - pods
  verbs:
  - list
`

const clusterRoleBindingMetallbSystemController = `
## Role bindings
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: metallb
  name: metallb-system:controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metallb-system:controller
subjects:
- kind: ServiceAccount
  name: controller
  namespace: metallb-system
`

const clusterRoleBindingMetallbSystemSpeaker = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: metallb
  name: metallb-system:speaker
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metallb-system:speaker
subjects:
- kind: ServiceAccount
  name: speaker
  namespace: metallb-system
`

const roleBindingConfigWatcher = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: metallb
  name: config-watcher
  namespace: metallb-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: config-watcher
subjects:
- kind: ServiceAccount
  name: controller
- kind: ServiceAccount
  name: speaker
`

const roleBindingPodLister = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app: metallb
  name: pod-lister
  namespace: metallb-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pod-lister
subjects:
- kind: ServiceAccount
  name: speaker
`

const deploymentController = `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: metallb
    component: controller
  name: controller
  namespace: metallb-system
spec:
  revisionHistoryLimit: 3
  selector:
    matchLabels:
      app: metallb
      component: controller
  template:
    metadata:
      annotations:
        prometheus.io/port: '7472'
        prometheus.io/scrape: 'true'
      labels:
        app: metallb
        component: controller
    spec:
      containers:
      - args:
        - --port=7472
        - --config=config
        image: quay.io/kinvolk/metallb-controller:v0.1.0-789-g85b7a46a
        imagePullPolicy: Always
        name: controller
        ports:
        - containerPort: 7472
          name: monitoring
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
      # XXX: Lokomotive specific change.
      {{- if .ControllerNodeSelectors }}
      nodeSelector:
        {{- range $key, $value := .ControllerNodeSelectors }}
        {{ $key }}: "{{ $value }}"
        {{- end }}
      {{- end }}
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccountName: controller
      terminationGracePeriodSeconds: 0
      # XXX: Lokomotive specific change.
      {{- if .ControllerTolerationsJSON }}
      tolerations: {{ .ControllerTolerationsJSON }}
      {{- end }}
`

// Divergence from upstream: disable fast dead node detection for layer 2. We use BGP and therefore
// have no use for this functionality. This way we also don't have to deal with generating
// memberlist secrets.
const daemonsetSpeaker = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: metallb
    component: speaker
  name: speaker
  namespace: metallb-system
spec:
  selector:
    matchLabels:
      app: metallb
      component: speaker
  template:
    metadata:
      annotations:
        prometheus.io/port: '7472'
        prometheus.io/scrape: 'true'
      labels:
        app: metallb
        component: speaker
    spec:
      containers:
      - args:
        - --metrics-port=7472
        - --status-port=7473
        - --config=config
        env:
        - name: METALLB_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: METALLB_METRICS_HOST
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        #- name: METALLB_ML_BIND_ADDR
        #  valueFrom:
        #    fieldRef:
        #      fieldPath: status.podIP
        # needed when another software is also using memberlist / port 7946
        #- name: METALLB_ML_BIND_PORT
        #  value: "7946"
        #- name: METALLB_ML_LABELS
        #  value: "app=metallb,component=speaker"
        #- name: METALLB_ML_NAMESPACE
        #  valueFrom:
        #    fieldRef:
        #      fieldPath: metadata.namespace
        #- name: METALLB_ML_SECRET_KEY
        #  valueFrom:
        #    secretKeyRef:
        #      name: memberlist
        #      key: secretkey
        image: quay.io/kinvolk/metallb-speaker:v0.1.0-789-g85b7a46a
        imagePullPolicy: Always
        name: speaker
        ports:
        - containerPort: 7472
          name: monitoring
        resources:
          limits:
            cpu: 100m
            memory: 100Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_ADMIN
            - NET_RAW
            - SYS_ADMIN
            drop:
            - ALL
          readOnlyRootFilesystem: true
      hostNetwork: true
      # XXX: Lokomotive specific change.
      {{- if .SpeakerNodeSelectors }}
      nodeSelector:
        {{- range $key, $value := .SpeakerNodeSelectors }}
        {{ $key }}: "{{ $value }}"
        {{- end }}
      {{- end }}
      serviceAccountName: speaker
      terminationGracePeriodSeconds: 2
      # XXX: Lokomotive specific change.
      {{- if .SpeakerTolerationsJSON }}
      tolerations: {{ .SpeakerTolerationsJSON }}
      {{- end }}
`

const pspMetallbController = `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  labels:
    app: metallb
  name: controller
  namespace: metallb-system
spec:
  allowPrivilegeEscalation: false
  allowedCapabilities: []
  allowedHostPaths: []
  defaultAddCapabilities: []
  defaultAllowPrivilegeEscalation: false
  fsGroup:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  hostIPC: false
  hostNetwork: false
  hostPID: false
  privileged: false
  readOnlyRootFilesystem: true
  requiredDropCapabilities:
  - ALL
  runAsUser:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    ranges:
    - max: 65535
      min: 1
    rule: MustRunAs
  volumes:
  - configMap
  - secret
  - emptyDir
`

const pspMetallbSpeaker = `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  labels:
    app: metallb
  name: speaker
  namespace: metallb-system
spec:
  allowPrivilegeEscalation: false
  allowedCapabilities:
  - NET_ADMIN
  - NET_RAW
  - SYS_ADMIN
  allowedHostPaths: []
  defaultAddCapabilities: []
  defaultAllowPrivilegeEscalation: false
  fsGroup:
    rule: RunAsAny
  hostIPC: false
  hostNetwork: true
  hostPID: false
  hostPorts:
  - max: 7472
    min: 7472
  privileged: true
  readOnlyRootFilesystem: true
  requiredDropCapabilities:
  - ALL
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  volumes:
  - configMap
  - secret
  - emptyDir
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
    peer-autodiscovery:
      from-labels:
        my-asn: metallb.lokomotive.io/my-asn
        peer-asn: metallb.lokomotive.io/peer-asn
        peer-address: metallb.lokomotive.io/peer-address
        peer-port: metallb.lokomotive.io/peer-port
        src-address: metallb.lokomotive.io/src-address
        hold-time: metallb.lokomotive.io/hold-time
        router-id: metallb.lokomotive.io/router-id
      from-annotations:
        my-asn: metallb.lokomotive.io/my-asn
        peer-address: metallb.lokomotive.io/peer-address
        peer-asn: metallb.lokomotive.io/peer-asn
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
      labels:
        severity: critical
      annotations:
        description: '{{ $labels.instance }}: MetalLB has not established a BGP session for more than 2 minutes.'
        summary: '{{ $labels.instance }}: MetalLB has not established BGP session.'
    - alert: MetalLBConfigStale
      expr: metallb_k8s_client_config_stale_bool != 0
      for: 2m
      labels:
        severity: critical
      annotations:
        description: '{{ $labels.instance }}: MetalLB instance has stale configuration.'
        summary: '{{ $labels.instance }}: MetalLB stale configuration.'
    - alert: MetalLBControllerPodsAvailability
      expr: kube_deployment_status_replicas_unavailable{deployment="controller",namespace="metallb-system"} != 0
      for: 1m
      labels:
        severity: critical
      annotations:
        description: '{{ $labels.instance }}: MetalLB Controller pod was not available in the last minute.'
        summary: '{{ $labels.instance }}: MetalLB Controller deployment pods.'
    - alert: MetalLBSpeakerPodsAvailability
      expr: kube_daemonset_status_number_unavailable{daemonset="speaker",namespace="metallb-system"} != 0
      for: 1m
      labels:
        severity: critical
      annotations:
        description: '{{ $labels.instance }}: MetalLB Speaker pod(s) were not available in the last minute.'
        summary: '{{ $labels.instance }}: MetalLB Speaker daemonset pods.'
`

const metallbPrometheusRuleUpdated = `
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: alertmanager-rules-for-prometheus-operator-0-43-2
  namespace: metallb-system
  labels:
    release: prometheus-operator
    app: kube-prometheus-stack
spec:
  groups:
  - name: metallb-rules
    rules:
    - alert: MetalLBNoBGPSession
      expr: metallb_bgp_session_up != 1
      for: 2m
      labels:
        severity: critical
      annotations:
        description: '{{ $labels.instance }}: MetalLB has not established a BGP session for more than 2 minutes.'
        summary: '{{ $labels.instance }}: MetalLB has not established BGP session.'
    - alert: MetalLBConfigStale
      expr: metallb_k8s_client_config_stale_bool != 0
      for: 2m
      labels:
        severity: critical
      annotations:
        description: '{{ $labels.instance }}: MetalLB instance has stale configuration.'
        summary: '{{ $labels.instance }}: MetalLB stale configuration.'
    - alert: MetalLBControllerPodsAvailability
      expr: kube_deployment_status_replicas_unavailable{deployment="controller",namespace="metallb-system"} != 0
      for: 1m
      labels:
        severity: critical
      annotations:
        description: '{{ $labels.instance }}: MetalLB Controller pod was not available in the last minute.'
        summary: '{{ $labels.instance }}: MetalLB Controller deployment pods.'
    - alert: MetalLBSpeakerPodsAvailability
      expr: kube_daemonset_status_number_unavailable{daemonset="speaker",namespace="metallb-system"} != 0
      for: 1m
      labels:
        severity: critical
      annotations:
        description: '{{ $labels.instance }}: MetalLB Speaker pod(s) were not available in the last minute.'
        summary: '{{ $labels.instance }}: MetalLB Speaker daemonset pods.'
`
