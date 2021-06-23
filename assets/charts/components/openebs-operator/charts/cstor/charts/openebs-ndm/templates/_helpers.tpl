{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
This name is used for ndm daemonset
*/}}
{{- define "openebs-ndm.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "openebs-ndm.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified ndm daemonset app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "openebs-ndm.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains .Release.Name $name }}
{{- $name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "openebs-ndm.operator.name" -}}
{{- $ndmName := default .Chart.Name .Values.ndmOperator.nameOverride | trunc 63 | trimSuffix "-" }}
{{- $componentName := .Values.ndmOperator.name | trunc 63 | trimSuffix "-" }}
{{- printf "%s-%s" $ndmName $componentName | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified ndm operator app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "openebs-ndm.operator.fullname" -}}
{{- if .Values.ndmOperator.fullnameOverride }}
{{- .Values.ndmOperator.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $ndmOperatorName := include "openebs-ndm.operator.name" .}}

{{- $name := default $ndmOperatorName .Values.ndmOperator.nameOverride }}
{{- if contains .Release.Name $name }}
{{- $name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "openebs-ndm.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "openebs-ndm.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Define meta labels for ndm components
*/}}
{{- define "openebs-ndm.common.metaLabels" -}}
chart: {{ template "openebs-ndm.chart" . }}
heritage: {{ .Release.Service }}
openebs.io/version: {{ .Values.release.version | quote }}
{{- end -}}


{{/*
Create match labels for ndm daemonset component
*/}}
{{- define "openebs-ndm.matchLabels" -}}
app: {{ template "openebs-ndm.name" . }}
release: {{ .Release.Name }}
component: {{ .Values.ndm.componentName | quote }}
{{- end -}}

{{/*
Create component labels for ndm daemonset component
*/}}
{{- define "openebs-ndm.componentLabels" -}}
openebs.io/component-name: {{ .Values.ndm.componentName | quote }}
{{- end -}}


{{/*
Create labels for ndm daemonset component
*/}}
{{- define "openebs-ndm.labels" -}}
{{ include "openebs-ndm.common.metaLabels" . }}
{{ include "openebs-ndm.matchLabels" . }}
{{ include "openebs-ndm.componentLabels" . }}
{{- end -}}

{{/*
Create match labels for ndm operator deployment
*/}}
{{- define "openebs-ndm.operator.matchLabels" -}}
app: {{ template "openebs-ndm.operator.name" . }}
release: {{ .Release.Name }}
component: {{ default (include "openebs-ndm.operator.name" .) .Values.ndmOperator.componentName }}
{{- end -}}

{{/*
Create component labels for ndm operator component
*/}}
{{- define "openebs-ndm.operator.componentLabels" -}}
openebs.io/component-name: {{ default (include "openebs-ndm.operator.name" .) .Values.ndmOperator.componentName }}
{{- end -}}


{{/*
Create labels for ndm operator component
*/}}
{{- define "openebs-ndm.operator.labels" -}}
{{ include "openebs-ndm.common.metaLabels" . }}
{{ include "openebs-ndm.operator.matchLabels" . }}
{{ include "openebs-ndm.operator.componentLabels" . }}
{{- end -}}
