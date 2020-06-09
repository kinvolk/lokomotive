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

package packet

var terraformConfigTmpl = `
module "packet-{{.Config.ClusterName}}" {
  source = "../lokomotive-kubernetes/packet/flatcar-linux/kubernetes"

  providers = {
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
    packet   = packet.default
  }

  dns_zone    = "{{.Config.DNS.Zone}}"

  ssh_keys  = {{.SSHPublicKeys}}
  asset_dir = "../cluster-assets"

  cluster_name = "{{.Config.ClusterName}}"
  tags         = {{.Tags}}
  project_id   = "{{.Config.ProjectID}}"
  facility     = "{{.Config.Facility}}"

  {{- if .Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.Config.ClusterDomainSuffix}}"
  {{- end }}

  controller_count = {{.Config.ControllerCount}}
  {{- if .Config.ControllerType }}
  controller_type  = "{{ .Config.ControllerType }}"
  {{- end }}

  {{- if .Config.OSArch }}
  os_arch = "{{ .Config.OSArch }}"
  {{- end }}
  {{- if .Config.OSChannel }}
  os_channel = "{{ .Config.OSChannel }}"
  {{- end }}
  {{- if .Config.OSVersion }}
  os_version = "{{ .Config.OSVersion }}"
  {{- end }}

  {{- if .Config.IPXEScriptURL }}
  ipxe_script_url = "{{ .Config.IPXEScriptURL }}"
  {{ end }}
  management_cidrs = {{.ManagementCIDRs}}
  node_private_cidr = "{{.Config.NodePrivateCIDR}}"

  enable_aggregation = {{.Config.EnableAggregation}}

  {{- if .Config.NetworkMTU }}
  network_mtu = {{.Config.NetworkMTU}}
  {{- end }}
  enable_reporting = {{.Config.EnableReporting}}

  {{- if .Config.PodCIDR }}
  pod_cidr = "{{.Config.PodCIDR}}"
  {{- end }}

  {{- if .Config.ServiceCIDR }}
  service_cidr = "{{.Config.ServiceCIDR}}"
  {{- end }}

  {{- if .Config.ReservationIDs }}
  reservation_ids = {
      {{- range $key, $value := .Config.ReservationIDs }}
      {{ $key }} = "{{ $value }}"
      {{- end }}
  }
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
  {{- if .Config.ReservationIDsDefault }}
  reservation_ids_default = "{{.Config.ReservationIDsDefault}}"
  {{- end }}
  {{- if .Config.CertsValidityPeriodHours }}
  certs_validity_period_hours = {{.Config.CertsValidityPeriodHours}}
  {{- end }}

  {{- if .Config.NodesDependOn }}
  nodes_depend_on = [
    {{- range $dep := .Config.NodesDependOn }}
    {{ $dep }},
    {{- end }}
  ]
  {{- end }}

  disable_self_hosted_kubelet = {{ .Config.DisableSelfHostedKubelet }}

  {{- if .Config.KubeAPIServerExtraFlags }}
  kube_apiserver_extra_flags = [
    {{- range .Config.KubeAPIServerExtraFlags }}
    "{{ . }}",
    {{- end }}
  ]
  {{- end }}
}

{{ range $index, $pool := .Config.WorkerPools }}
module "worker-{{ $pool.Name }}" {
  source = "../lokomotive-kubernetes/packet/flatcar-linux/kubernetes/workers"

  providers = {
    local    = local.default
    template = template.default
    tls      = tls.default
    packet   = packet.default
  }

  dns_zone = "{{$.Config.DNS.Zone}}"

  ssh_keys  = {{$.SSHPublicKeys}}

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

  cluster_name = "{{$.Config.ClusterName}}"

  {{- if $pool.Tags }}
  tags = [
      {{- range $key, $value := $pool.Tags }}
      "{{ $key }}:{{ $value }}",
      {{- end }}
	]
  {{- end }}

  project_id   = "{{$.Config.ProjectID}}"
  facility     = "{{$.Config.Facility}}"
  {{- if $.Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{$.Config.ClusterDomainSuffix}}"
  {{- end }}

  pool_name    = "{{ $pool.Name }}"
  worker_count = {{ $pool.Count }}
  {{- if $pool.NodeType }}
  type      = "{{ $pool.NodeType }}"
  {{- end }}

  {{- if $.Config.IPXEScriptURL }}
  ipxe_script_url = "{{ $pool.IPXEScriptURL }}"
  {{- end }}

  {{- if $pool.OSArch }}
  os_arch = "{{ $pool.OSArch }}"
  {{- end }}
  {{- if $pool.OSChannel }}
  os_channel = "{{ $pool.OSChannel }}"
  {{- end }}
  {{- if $pool.OSVersion }}
  os_version = "{{ $pool.OSVersion }}"
  {{- end }}

  kubeconfig = module.packet-{{ $.Config.ClusterName }}.kubeconfig

  {{- if $pool.Labels }}
  labels = "{{ $pool.Labels }}"
  {{- end }}
  {{- if $pool.Taints }}
  taints = "{{ $pool.Taints }}"
  {{- end }}
  {{- if $.Config.ServiceCIDR }}
  service_cidr = "{{$.Config.ServiceCIDR}}"
  {{- end }}

  {{- if $pool.SetupRaid }}
  setup_raid = {{ $pool.SetupRaid }}
  {{- end }}
  {{- if $pool.SetupRaidHDD }}
  setup_raid_hdd = {{ $pool.SetupRaidHDD }}
  {{- end }}
  {{- if $pool.SetupRaidSSD }}
  setup_raid_ssd = {{ $pool.SetupRaidSSD }}
  {{- end }}
  {{- if $pool.SetupRaidSSD }}
  setup_raid_ssd_fs = {{ $pool.SetupRaidSSDFS }}
  {{- end }}

  {{- if $pool.DisableBGP }}
  disable_bgp = true
  {{- end}}

  {{- if $pool.ReservationIDs }}
  reservation_ids = {
    {{- range $key, $value := $pool.ReservationIDs }}
    {{ $key }} = "{{ $value }}"
    {{- end }}
  }
  {{- end }}
  {{- if $pool.ReservationIDsDefault }}
  reservation_ids_default = "{{$pool.ReservationIDsDefault}}"
  {{- end }}

  {{- if $pool.NodesDependOn }}
  nodes_depend_on = [
    {{- range $dep := $pool.NodesDependOn }}
    {{ $dep }},
    {{- end }}
  ]
  {{- end }}

}
{{- end }}

module "dns" {
  source = "../lokomotive-kubernetes/dns/{{.Config.DNS.Provider}}"

  {{ if eq .Config.DNS.Provider "route53" -}}
  providers = {
    aws = aws.default
  }

  {{ end -}}
  cluster_name             = "{{ .Config.ClusterName }}"
  controllers_public_ipv4  = module.packet-{{.Config.ClusterName}}.controllers_public_ipv4
  controllers_private_ipv4 = module.packet-{{.Config.ClusterName}}.controllers_private_ipv4
  dns_zone                 = "{{ .Config.DNS.Zone }}"
}

{{- if eq .Config.DNS.Provider "manual" }}

output "dns_entries" {
  value = module.dns.entries
}
{{- end }}

{{- if eq .Config.DNS.Provider "route53" }}
provider "aws" {
  version = "2.48.0"
  alias   = "default"
  # The Route 53 service doesn't need a specific region to operate, however
  # the AWS Terraform provider needs it and the documentation suggests to use
  # "us-east-1": https://docs.aws.amazon.com/general/latest/gr/r53.html.
  region = "us-east-1"
}
{{- end }}

provider "ct" {
  version = "~> 0.3"
}

provider "local" {
  version = "1.4.0"
  alias   = "default"
}

provider "null" {
  version = "~> 2.1"
  alias   = "default"
}

provider "template" {
  version = "~> 2.1"
  alias   = "default"
}

provider "tls" {
  version = "~> 2.0"
  alias   = "default"
}

provider "packet" {
  version = "~> 2.7.3"
  alias = "default"

  {{- if .Config.AuthToken }}
  auth_token = "{{.Config.AuthToken}}"
  {{- end }}
}

# Stub output, which indicates, that Terraform run at least once.
# Used when checking, if we should ask user for confirmation, when
# applying changes to the cluster.
output "initialized" {
  value = true
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value = module.packet-{{.Config.ClusterName}}.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value     = module.packet-{{.Config.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.packet-{{.Config.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.packet-{{.Config.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value = module.packet-{{.Config.ClusterName}}.calico_values
}
`
