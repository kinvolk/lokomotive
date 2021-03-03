{{ define "containers" }}
      hostNetwork: true
      nodeSelector:
        node.kubernetes.io/master: ""
      priorityClassName: system-cluster-critical
      serviceAccountName: kube-apiserver
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      containers:
      - name: kube-apiserver
        image: {{ .Values.apiserver.image }}
        command:
        - kube-apiserver
        - --advertise-address=$(POD_IP)
        - --allow-privileged=true
        - --anonymous-auth=false
        {{- if .Values.apiserver.enableTLSBootstrap }}
        - --authorization-mode=Node,RBAC
        {{- else }}
        - --authorization-mode=RBAC
        {{- end }}
        - --bind-address=0.0.0.0
        - --client-ca-file=/etc/kubernetes/secrets/ca.crt
        - --cloud-provider={{ .Values.apiserver.cloudProvider }}
        {{- if .Values.apiserver.enableTLSBootstrap }}
        - --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultTolerationSeconds,DefaultStorageClass,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,Priority,PodSecurityPolicy,NodeRestriction
        - --enable-bootstrap-token-auth=true
        {{- else }}
        - --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultTolerationSeconds,DefaultStorageClass,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,Priority,PodSecurityPolicy
        {{- end }}
        - --etcd-cafile=/etc/kubernetes/secrets/etcd-client-ca.crt
        - --etcd-certfile=/etc/kubernetes/secrets/etcd-client.crt
        - --etcd-keyfile=/etc/kubernetes/secrets/etcd-client.key
        - --etcd-servers={{ .Values.apiserver.etcdServers}}
        - --insecure-port=0
        - --kubelet-client-certificate=/etc/kubernetes/secrets/apiserver.crt
        - --kubelet-client-key=/etc/kubernetes/secrets/apiserver.key
        - --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname
        - --secure-port=6443
        - --service-account-key-file=/etc/kubernetes/secrets/service-account.key
        - --service-account-signing-key-file=/etc/kubernetes/secrets/service-account.key
        - --service-account-issuer=https://kubernetes.default.svc
        - --service-cluster-ip-range={{ .Values.apiserver.serviceCIDR }}
        - --storage-backend=etcd3
        - --tls-cert-file=/etc/kubernetes/secrets/apiserver.crt
        - --tls-private-key-file=/etc/kubernetes/secrets/apiserver.key
        - --token-auth-file=/etc/kubernetes/secrets/token-auth-file
        {{- if .Values.apiserver.enableAggregation }}
        - --proxy-client-cert-file=/etc/kubernetes/secrets/aggregation-client.crt
        - --proxy-client-key-file=/etc/kubernetes/secrets/aggregation-client.key
        - --requestheader-client-ca-file=/etc/kubernetes/secrets/aggregation-ca.crt
        - --requestheader-extra-headers-prefix=X-Remote-Extra-
        - --requestheader-group-headers=X-Remote-Group
        - --requestheader-username-headers=X-Remote-User
        {{- end }}
        {{- range .Values.apiserver.extraFlags }}
        - {{ . }}
        {{- end }}
        - --permit-port-sharing=true
        env:
        {{- if .Values.apiserver.ignoreX509CNCheck }}
        - name: GODEBUG
          value: x509ignoreCN=0
        {{- end }}
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        readinessProbe:
          httpGet:
            httpHeaders:
            - name: Authorization
              value: Bearer {{ template "token" . }}
            path: /healthz
            port: 6443
            scheme: HTTPS
        volumeMounts:
        - name: secrets
          mountPath: /etc/kubernetes/secrets
          readOnly: true
        - name: ssl-certs-host
          mountPath: /etc/ssl/certs
          readOnly: true
      volumes:
      - name: secrets
        secret:
          secretName: kube-apiserver
      - name: ssl-certs-host
        hostPath:
          path: {{ .Values.apiserver.trustedCertsDir }}
{{- end }}
