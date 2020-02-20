{{/* vim: set filetype=mustache: */}}
{{/*
Compile all warnings into a single message, and call fail.
*/}}
{{- define "calico.validateValues" -}}
{{- $messages := list -}}
{{- $messages := append $messages (include "calico.validateValues.managementCIDRs.set" .) -}}
{{- $messages := append $messages (include "calico.validateValues.managementCIDRs.noEmptyValues" .) -}}
{{- $messages := append $messages (include "calico.validateValues.clusterCIDRs.set" .) -}}
{{- $messages := append $messages (include "calico.validateValues.clusterCIDRs.noEmptyValues" .) -}}
{{- $messages := without $messages "" -}}
{{- $message := join "\n" $messages -}}

{{- if $message -}}
{{-   printf "\nVALUES VALIDATION:\n%s" $message | fail -}}
{{- end -}}
{{- end -}}

{{- define "calico.validateValues.managementCIDRs.set" -}}
{{- if eq (len .Values.managementCIDRs) 0 -}}
calico:managementCIDRs
    You must set at least on management CIDR.
    Please set the.managementCIDRs parameter (--set.managementCIDRs={})
{{- end -}}
{{- end -}}

{{- define "calico.validateValues.managementCIDRs.noEmptyValues" -}}
{{- range $i, $v := .Values.managementCIDRs }}
{{- if eq $v "" }}
calico:managementCIDRs
    Value at index {{ $i }} is empty. All values must be valid CIDR notation.
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "calico.validateValues.clusterCIDRs.set" -}}
{{- if eq (len .Values.clusterCIDRs) 0 -}}
calico:clusterCIDRs
    You must set at least on management CIDR.
    Please set the.clusterCIDRs parameter (--set.clusterCIDRs={})
{{- end -}}
{{- end -}}

{{- define "calico.validateValues.clusterCIDRs.noEmptyValues" -}}
{{- range $i, $v := .Values.clusterCIDRs }}
{{- if eq $v "" }}
calico:clusterCIDRs
    Value at index {{ $i }} is empty. All values must be valid CIDR notation.
{{- end -}}
{{- end -}}
{{- end -}}
