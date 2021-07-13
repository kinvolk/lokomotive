{{- define "azure-arc-onboarding.envVars" -}}
- name: AZURE_APPLICATION_CLIENT_ID
  value: {{ .Values.applicationClientID | required "applicationClientID is required" }}
- name: AZURE_APPLICATION_PASSWORD
  valueFrom:
    secretKeyRef:
      name: azure-application-password
      key: azure-application-password
- name: AZURE_TENANT_ID
  value: {{ .Values.tenantID | required "tenantID is required" }}
- name: CONNECTED_CLUSTER_NAME
  value: {{ .Values.clusterName | required "clusterName is required" }}
- name: AZURE_RESOURCE_GROUP
  value: {{ .Values.resourceGroup | required " resourceGroup is required" }}
{{- end -}}
