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

package aks

var terraformConfigTmpl = `{{- define "resource_group_name" -}}
{{- if .ManageResourceGroup -}}
azurerm_resource_group.aks.name
{{- else -}}
"{{ .ResourceGroupName }}"
{{- end -}}
{{- end -}}

{{- define "client_id" -}}
{{- if .ApplicationName -}}
azuread_application.aks.application_id
{{- else -}}
"{{ .ClientID }}"
{{- end -}}
{{- end -}}

{{- define "client_secret" -}}
{{- if .ApplicationName -}}
azuread_application_password.aks.value
{{- else -}}
"{{ .ClientSecret }}"
{{- end -}}
{{- end -}}

locals {
  subscription_id           = "{{ .SubscriptionID }}"
  tenant_id                 = "{{ .TenantID }}"
  application_name          = "{{ .ApplicationName }}"
  location                  = "{{ .Location }}"
  resource_group_name       = {{ template "resource_group_name" . }}
  kubernetes_version        = "{{ .KubernetesVersion }}"
  cluster_name              = "{{ .ClusterName }}"
  default_node_pool_name    = "{{ (index .WorkerPools 0).Name }}"
  default_node_pool_vm_size = "{{ (index .WorkerPools 0).VMSize }}"
  default_node_pool_count   = {{ (index .WorkerPools 0).Count  }}
  client_id                 = {{ template "client_id" . }}
  client_secret             = {{ template "client_secret" . }}
}

provider "azurerm" {
  version = "2.24.0"

  # https://github.com/terraform-providers/terraform-provider-azurerm/issues/5893
  features {}
}

provider "local" {
  version = "1.4.0"
}

{{- if .ApplicationName }}
provider "azuread" {
  version = "0.11.0"
}

provider "random" {
  version = "2.3.0"
}

resource "azuread_application" "aks" {
  name = local.application_name
}

resource "azuread_service_principal" "aks" {
  application_id = azuread_application.aks.application_id

  {{- if .Tags }}
  tags = [
    {{- range $k, $v := .Tags }}
    "{{ $k }}={{ $v }}",
    {{- end }}
  ]
  {{- end }}
}

resource "random_string" "password" {
  length  = 16
  special = true

  override_special = "/@\" "
}

resource "azuread_application_password" "aks" {
  application_object_id = azuread_application.aks.object_id
  value                 = random_string.password.result
  end_date_relative     = "86000h"
}

resource "azurerm_role_assignment" "aks" {
  scope                = "/subscriptions/${local.subscription_id}"
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.aks.id
}
{{- end }}

{{- if .ManageResourceGroup }}
resource "azurerm_resource_group" "aks" {
  name     = "{{ .ResourceGroupName }}"
  location = local.location

  {{- if .Tags }}
  tags = {
    {{- range $k, $v := .Tags }}
    "{{ $k }}" = "{{ $v }}"
    {{- end }}
  }
  {{- end }}
}
{{- end }}

resource "azurerm_kubernetes_cluster" "aks" {
  name                = local.cluster_name
  location            = local.location
  resource_group_name = local.resource_group_name
  kubernetes_version  = local.kubernetes_version
  dns_prefix          = local.cluster_name

  default_node_pool {
    name       = local.default_node_pool_name
    vm_size    = local.default_node_pool_vm_size
    node_count = local.default_node_pool_count

    {{- if (index .WorkerPools 0).Labels }}
    node_labels = {
      {{- range $k, $v := (index .WorkerPools 0).Labels }}
      "{{ $k }}" = "{{ $v }}"
      {{- end }}
    }
    {{- end }}

    {{- if (index .WorkerPools 0).Taints }}
    node_taints = [
      {{- range (index .WorkerPools 0).Taints }}
      "{{ . }}",
      {{- end }}
    ]
    {{- end }}
  }

  role_based_access_control {
    enabled = true
  }

  service_principal {
    client_id     = local.client_id
    client_secret = local.client_secret
  }

  network_profile {
    network_plugin = "kubenet"
    network_policy = "calico"
  }

  {{- if .Tags }}
  tags = {
    {{- range $k, $v := .Tags }}
    "{{ $k }}" = "{{ $v }}"
    {{- end }}
  }
  {{- end }}
}

{{ range $index, $pool := (slice .WorkerPools 1) }}
resource "azurerm_kubernetes_cluster_node_pool" "worker-{{ $pool.Name }}" {
  name                  = "{{ $pool.Name }}"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.aks.id
  vm_size               = "{{ $pool.VMSize }}"
  node_count            = "{{ $pool.Count }}"

  {{- if $pool.Labels }}
  node_labels = {
    {{- range $k, $v := $pool.Labels }}
    "{{ $k }}" = "{{ $v }}"
    {{- end }}
  }
  {{- end }}


  {{- if $pool.Taints }}
  node_taints = [
    {{- range $pool.Taints }}
    "{{ . }}",
    {{- end }}
  ]
  {{- end }}


  {{- if $.Tags }}
  tags = {
    {{- range $k, $v := $.Tags }}
    "{{ $k }}" = "{{ $v }}"
    {{- end }}
  }
  {{- end }}
}
{{- end }}

resource "local_file" "kubeconfig" {
  sensitive_content = azurerm_kubernetes_cluster.aks.kube_config_raw
  filename          = "../cluster-assets/auth/kubeconfig"
}

# Stub output which indicates, that Terraform ran at least once.
# Used when checking, if we should ask user for confirmation when
# applying changes to the cluster.
output "initialized" {
  value     = true
  sensitive = true
}
`
