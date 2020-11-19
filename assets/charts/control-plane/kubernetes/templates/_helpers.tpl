{{ define "scheduler-containers" }}
      nodeSelector:
        node.kubernetes.io/master: ""
      priorityClassName: system-cluster-critical
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccountName: kube-scheduler
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      containers:
      - name: kube-scheduler
        image: "{{ .Values.kubeScheduler.image }}"
        command:
        - kube-scheduler
        - --leader-elect=true
        livenessProbe:
          httpGet:
            scheme: HTTPS
            path: /healthz
            port: 10259
          initialDelaySeconds: 15
          timeoutSeconds: 15
{{- end }}

{{ define "controller-manager-containers" }}
      nodeSelector:
        node.kubernetes.io/master: ""
      priorityClassName: system-cluster-critical
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccountName: kube-controller-manager
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      containers:
      - name: kube-controller-manager
        image: {{ .Values.controllerManager.image }}
        command:
        - kube-controller-manager
        - --use-service-account-credentials
        - --allocate-node-cidrs=true
        - --cloud-provider={{ .Values.controllerManager.cloudProvider }}
        - --cluster-cidr={{ .Values.controllerManager.podCIDR }}
        - --service-cluster-ip-range={{ .Values.controllerManager.serviceCIDR }}
        - --cluster-signing-cert-file=/etc/kubernetes/secrets/ca.crt
        - --cluster-signing-key-file=/etc/kubernetes/secrets/ca.key
        - --configure-cloud-routes=false
        - --leader-elect=true
        - --flex-volume-plugin-dir=/var/lib/kubelet/volumeplugins
        - --pod-eviction-timeout=1m
        - --root-ca-file=/etc/kubernetes/secrets/ca.crt
        - --service-account-private-key-file=/etc/kubernetes/secrets/service-account.key
        - --cluster-signing-duration=45m
        livenessProbe:
          httpGet:
            scheme: HTTPS
            path: /healthz
            port: 10257
          initialDelaySeconds: 15
          timeoutSeconds: 15
        volumeMounts:
        - name: secrets
          mountPath: /etc/kubernetes/secrets
          readOnly: true
        - name: volumeplugins
          mountPath: /var/lib/kubelet/volumeplugins
          readOnly: true
        - name: ssl-host
          mountPath: /etc/ssl/certs
          readOnly: true
      volumes:
      - name: secrets
        secret:
          secretName: kube-controller-manager
      - name: ssl-host
        hostPath:
          path: {{ .Values.controllerManager.trustedCertsDir }}
      - name: volumeplugins
        hostPath:
          path: /var/lib/kubelet/volumeplugins
      dnsPolicy: Default # Don't use cluster DNS.
{{- end }}

{{ define "coredns-containers" }}
      nodeSelector:
        node.kubernetes.io/master: ""
      priorityClassName: system-cluster-critical
      serviceAccountName: coredns
      tolerations:
        - key: node-role.kubernetes.io/master
          effect: NoSchedule
      containers:
        - name: coredns
          image: {{ .Values.coredns.image }}
          resources:
            limits:
              memory: 170Mi
            requests:
              cpu: 100m
              memory: 70Mi
          args: [ "-conf", "/etc/coredns/Corefile" ]
          volumeMounts:
            - name: config
              mountPath: /etc/coredns
              readOnly: true
          ports:
            - name: dns
              protocol: UDP
              containerPort: 53
            - name: dns-tcp
              protocol: TCP
              containerPort: 53
            - name: metrics
              protocol: TCP
              containerPort: 9153
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
              scheme: HTTP
            initialDelaySeconds: 60
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 5
          readinessProbe:
            httpGet:
              path: /ready
              port: 8181
              scheme: HTTP
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              add:
              - NET_BIND_SERVICE
              drop:
              - all
            readOnlyRootFilesystem: true
      dnsPolicy: Default
      volumes:
        - name: config
          configMap:
            name: coredns
            items:
            - key: Corefile
              path: Corefile
{{- end }}
