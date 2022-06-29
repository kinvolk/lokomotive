// Copyright 2022 The Lokomotive Authors
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

package azure

var terraformConfigTmpl = `
module "azure-{{.Config.ClusterName}}" {
  source = "../terraform-modules/azure/flatcar-linux/kubernetes"
  dns_zone    = "{{.Config.DNS.Zone}}"
  region = "{{.Config.Region}}"
  ssh_keys  = {{.SSHPublicKeys}}
  asset_dir = "../cluster-assets"
  cluster_name = "{{.Config.ClusterName}}"
  tags         = {{.Tags}}

  {{- if .Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.Config.ClusterDomainSuffix}}"
  {{- end }}

  controller_count = {{.Config.ControllerCount}}

  {{- if .Config.ControllerType }}
  controller_type  = "{{ .Config.ControllerType }}"
  {{- end }}

  {{- if .Config.WorkerType }}
  worker_type  = "{{ .Config.WorkerType }}"
  {{- end }}

  {{- if .Config.OSImage }}
  os_image = "{{ .Config.OSImage }}"
  {{- end }}

  enable_aggregation = {{.Config.EnableAggregation}}
  enable_reporting = {{.Config.EnableReporting}}

  {{- if .Config.PodCIDR }}
  pod_cidr = "{{.Config.PodCIDR}}"
  {{- end }}

  {{- if .Config.ServiceCIDR }}
  service_cidr = "{{.Config.ServiceCIDR}}"
  {{- end }}

  {{- if .Config.ControllerCLCSnippets}}
  controller_clc_snippets = [
  {{- range $clc_snippet := .Config.ControllerCLCSnippets}}
    <<EOF
{{ $clc_snippet }}
EOF
    ,
  {{- end }}
  ]
  {{- end }}

  {{- if .Config.CertsValidityPeriodHours }}
  certs_validity_period_hours = {{.Config.CertsValidityPeriodHours}}
  {{- end }}

  disable_self_hosted_kubelet = {{ .Config.DisableSelfHostedKubelet }}

  {{- if .Config.KubeAPIServerExtraFlags }}
  kube_apiserver_extra_flags = [
    {{- range .Config.KubeAPIServerExtraFlags }}
    "{{ . }}",
    {{- end }}
  ]
  {{- end }}

  conntrack_max_per_core = {{.Config.ConntrackMaxPerCore}}
  enable_tls_bootstrap    = {{ .Config.EnableTLSBootstrap }}

  {{- if .Config.EncryptPodTraffic }}
  encrypt_pod_traffic = {{.Config.EncryptPodTraffic}}
  {{- end }}

  worker_bootstrap_tokens = [
    {{- range $index, $pool := .Config.WorkerPools }}
    module.worker-{{$pool.Name}}.worker_bootstrap_token,
    {{- end }}
  ]
}

{{ range $index, $pool := .Config.WorkerPools }}
module "worker-{{ $pool.Name }}" {
  source = "../terraform-modules/azure/flatcar-linux/kubernetes/workers"
  dns_zone    = "{{$.Config.DNS.Zone}}"
  cluster_name = "{{$.Config.ClusterName}}"
  pool_name = "{{ $pool.Name }}"
  resource_group_name = module.azure-{{ $.Config.ClusterName }}.resource_group_name
  region = "{{$.Config.Region}}"
  subnet_id = module.azure-{{ $.Config.ClusterName }}.subnet_id
  backend_address_pool_id = module.azure-{{ $.Config.ClusterName }}.backend_address_pool_id
  security_group_id = module.azure-{{ $.Config.ClusterName }}.security_group_id
  ca_cert               = module.azure-{{ $.Config.ClusterName }}.ca_cert
  apiserver             = module.azure-{{ $.Config.ClusterName }}.apiserver
  kubeconfig = module.azure-{{ $.Config.ClusterName }}.kubeconfig
  ssh_keys  = {{$.SSHPublicKeys}}
  worker_count = {{$pool.Count}}
  {{- if $pool.CLCSnippets}}
  clc_snippets = [
  {{- range $clc_snippet := $pool.CLCSnippets}}
    <<EOF
{{ $clc_snippet }}
EOF
    ,
  {{- end}}
  ]
  {{- end }}
  {{- if $pool.Labels }}
  labels = {
  {{- range $k, $v := $pool.Labels }}
    "{{ $k }}" = "{{ $v }}",
  {{- end }}
  }
  {{- end }}
  {{- if $pool.Taints }}
  taints = {
  {{- range $k, $v := $pool.Taints }}
    "{{ $k }}" = "{{ $v }}",
  {{- end }}
  }
  {{- end}}
  {{- if $pool.Tags }}
  tags = [
      {{- range $key, $value := $pool.Tags }}
      "{{ $key }}:{{ $value }}",
      {{- end }}
  ]
  {{- end }}
  {{- if $.Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{$.Config.ClusterDomainSuffix}}"
  {{- end }}
  {{- if $pool.VMType }}
  vm_type = "{{ $pool.VMType }}"
  {{- end }}
  {{- if $pool.OSImage }}
  os_image = "{{ $pool.OSImage }}"
  {{- end }}
  enable_tls_bootstrap = {{ $.Config.EnableTLSBootstrap }}
  {{- if $.Config.ServiceCIDR }}
  service_cidr = "{{$.Config.ServiceCIDR}}"
  {{- end }}
  {{- if $pool.Priority }}
  priority = "{{$pool.Priority}}"
  {{- end }}
  {{- if $pool.CLCSnippets }}
  clc_snippets = {{ (index $.WorkerpoolCfg $index "clc_snippets") }}
  {{- end }}
  {{- if $pool.CPUManagerPolicy }}
  cpu_manager_policy = "{{$pool.CPUManagerPolicy}}"
  {{- end}}
}
{{- end }}

provider "azurerm" {
  # https://github.com/terraform-providers/terraform-provider-azurerm/issues/5893
  features {}
}

terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.97.0"
    }
  }
}

module "dns" {
  source = "../terraform-modules/dns/{{.Config.DNS.Provider}}"
  cluster_name             = "{{ .Config.ClusterName }}"
  controllers_public_ipv4  = module.azure-{{.Config.ClusterName}}.controllers_public_ipv4
  controllers_private_ipv4 = module.azure-{{.Config.ClusterName}}.controllers_private_ipv4
  dns_zone                 = "{{ .Config.DNS.Zone }}"
}

{{- if eq .Config.DNS.Provider "manual" }}
output "dns_entries" {
  value = module.dns.entries
}
{{- end }}

{{- if eq .Config.DNS.Provider "route53" }}
provider "aws" {
  # The Route 53 service doesn't need a specific region to operate, however
  # the AWS Terraform provider needs it and the documentation suggests to use
  # "us-east-1": https://docs.aws.amazon.com/general/latest/gr/r53.html.
  region = "us-east-1"
}
{{- end }}

# Stub output, which indicates, that Terraform run at least once.
# Used when checking, if we should ask user for confirmation, when
# applying changes to the cluster.
output "initialized" {
  value     = true
  sensitive = true
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value     = module.azure-{{.Config.ClusterName}}.pod-checkpointer_values
  sensitive = true
}

output "kube-apiserver_values" {
  value     = module.azure-{{.Config.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.azure-{{.Config.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.azure-{{.Config.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value     = module.azure-{{.Config.ClusterName}}.calico_values
  sensitive = true
}

output "lokomotive_values" {
  value     = module.azure-{{.Config.ClusterName}}.lokomotive_values
  sensitive = true
}

output "bootstrap-secrets_values" {
  value     = module.azure-{{.Config.ClusterName}}.bootstrap-secrets_values
  sensitive = true
}

output "node-local-dns_values" {
  value     = module.azure-{{.Config.ClusterName}}.node-local-dns_values
  sensitive = true
}

output "kubeconfig" {
  value     = module.azure-{{.Config.ClusterName}}.kubeconfig-admin
  sensitive = true
}
`
