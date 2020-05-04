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

package contour

var template = map[string]string{
	"02-service-envoy.yaml": `
---
apiVersion: v1
kind: Service
metadata:
  name: envoy
  namespace: projectcontour
  annotations:
    # This annotation puts the AWS ELB into "TCP" mode so that it does not
    # do HTTP negotiation for HTTPS connections at the ELB edge.
    # The downside of this is the remote IP address of all connections will
    # appear to be the internal address of the ELB. See docs/proxy-proto.md
    # for information about enabling the PROXY protocol on the ELB to recover
    # the original remote IP address.
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: tcp
    {{- if .IngressHosts }}
    external-dns.alpha.kubernetes.io/hostname: "{{ .IngressHostsRaw }}"
    {{- end }}
spec:
  externalTrafficPolicy: Local
  ports:
  - port: 80
    name: http
    protocol: TCP
  - port: 443
    name: https
    protocol: TCP
  selector:
    app: envoy
  type: LoadBalancer
`,

	"03-contour.yaml": `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: contour
  name: contour
  namespace: projectcontour
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      # This value of maxSurge means that during a rolling update
      # the new ReplicaSet will be created first.
      maxSurge: 50%
  selector:
    matchLabels:
      app: contour
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8000"
      labels:
        app: contour
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app: contour
              topologyKey: kubernetes.io/hostname
            weight: 100
        {{- if .NodeAffinity }}
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                {{- range $item := .NodeAffinity }}
                - key: {{ $item.Key }}
                  operator: {{ $item.Operator }}
                  {{- if $item.Values }}
                  values:
                    {{- range $val := $item.Values }}
                    - {{ $val }}
                    {{- end }}
                  {{- end }}
                {{- end }}
        {{- end}}
      {{- if .TolerationsRaw }}
      tolerations: {{ .TolerationsRaw }}
      {{- end }}
      containers:
      - args:
        - serve
        - --incluster
        - --xds-address=0.0.0.0
        - --xds-port=8001
        - --envoy-service-http-port=80
        - --envoy-service-https-port=443
        - --contour-cafile=/ca/cacert.pem
        - --contour-cert-file=/certs/tls.crt
        - --contour-key-file=/certs/tls.key
        - --config-path=/config/contour.yaml
        command: ["contour"]
        image: docker.io/projectcontour/contour:v1.3.0
        imagePullPolicy: Always
        name: contour
        ports:
        - containerPort: 8001
          name: xds
          protocol: TCP
        - containerPort: 8000
          name: debug
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8000
        readinessProbe:
          tcpSocket:
            port: 8001
          initialDelaySeconds: 15
          periodSeconds: 10
        volumeMounts:
          - name: contourcert
            mountPath: /certs
            readOnly: true
          - name: cacert
            mountPath: /ca
            readOnly: true
          - name: contour-config
            mountPath: /config
            readOnly: true
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
      dnsPolicy: ClusterFirst
      serviceAccountName: contour
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        runAsGroup: 65534
      volumes:
        - name: contourcert
          secret:
            secretName: contourcert
        - name: cacert
          secret:
            secretName: cacert
        - name: contour-config
          configMap:
            name: contour
            defaultMode: 0644
            items:
            - key: contour.yaml
              path: contour.yaml
`,

	"03-envoy.yaml": `
---
# XXX: Lokomotive specific change
apiVersion: v1
kind: ServiceAccount
metadata:
  name: envoy
  namespace: projectcontour
---
# XXX: Lokomotive specific change
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: envoy-privileged-psp
  namespace: projectcontour
roleRef:
  kind: ClusterRole
  name: privileged-psp
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: envoy
  namespace: projectcontour
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: envoy
  name: envoy
  namespace: projectcontour
spec:
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 10%
  selector:
    matchLabels:
      app: envoy
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8002"
        prometheus.io/path: "/stats/prometheus"
      labels:
        app: envoy
    spec:
      {{- if .NodeAffinity }}
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                {{- range $item := .NodeAffinity }}
                - key: {{ $item.Key }}
                  operator: {{ $item.Operator }}
                  {{- if $item.Values }}
                  values:
                    {{- range $val := $item.Values }}
                    - {{ $val }}
                    {{- end }}
                  {{- end }}
                {{- end }}
      {{- end}}
      {{- if .TolerationsRaw }}
      tolerations: {{ .TolerationsRaw }}
      {{- end }}
      containers:
      - command:
        - /bin/contour
        args:
          - envoy
          - shutdown-manager
        image: docker.io/projectcontour/contour:v1.3.0
        imagePullPolicy: Always
        lifecycle:
          preStop:
            httpGet:
              path: /shutdown
              port: 8090
              scheme: HTTP
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8090
          initialDelaySeconds: 3
          periodSeconds: 10
        name: shutdown-manager
      - args:
        - -c
        - /config/envoy.json
        - --service-cluster $(CONTOUR_NAMESPACE)
        - --service-node $(ENVOY_POD_NAME)
        - --log-level info
        command:
        - envoy
        image: docker.io/envoyproxy/envoy:v1.13.1
        imagePullPolicy: IfNotPresent
        name: envoy
        env:
        - name: CONTOUR_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: ENVOY_POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        ports:
        - containerPort: 80
          hostPort: 80
          name: http
          protocol: TCP
        - containerPort: 443
          hostPort: 443
          name: https
          protocol: TCP
        readinessProbe:
          httpGet:
            path: /ready
            port: 8002
          initialDelaySeconds: 3
          periodSeconds: 4
        volumeMounts:
          - name: envoy-config
            mountPath: /config
          - name: envoycert
            mountPath: /certs
          - name: cacert
            mountPath: /ca
        lifecycle:
          preStop:
            httpGet:
              path: /shutdown
              port: 8090
              scheme: HTTP
      initContainers:
      - args:
        - bootstrap
        - /config/envoy.json
        - --xds-address=contour
        - --xds-port=8001
        - --envoy-cafile=/ca/cacert.pem
        - --envoy-cert-file=/certs/tls.crt
        - --envoy-key-file=/certs/tls.key
        command:
        - contour
        image: docker.io/projectcontour/contour:v1.3.0
        imagePullPolicy: Always
        name: envoy-initconfig
        volumeMounts:
        - name: envoy-config
          mountPath: /config
        - name: envoycert
          mountPath: /certs
          readOnly: true
        - name: cacert
          mountPath: /ca
          readOnly: true
        env:
        - name: CONTOUR_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      automountServiceAccountToken: false
      # XXX: Lokomotive specific change
      serviceAccountName: envoy
      terminationGracePeriodSeconds: 300
      volumes:
        - name: envoy-config
          emptyDir: {}
        - name: envoycert
          secret:
            secretName: envoycert
        - name: cacert
          secret:
            secretName: cacert
      restartPolicy: Always

`,
}
