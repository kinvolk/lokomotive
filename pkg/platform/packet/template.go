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
module "packet-{{.ClusterName}}" {
  source = "../lokomotive-kubernetes/packet/flatcar-linux/kubernetes"

  providers = {
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
    packet   = packet.default
  }

  dns_zone    = "{{.DNS.Zone}}"

  ssh_keys  = {{.SSHPubKeysRaw}}
  asset_dir = "../cluster-assets"

  cluster_name = "{{.ClusterName}}"
  tags         = {{.TagsRaw}}
  project_id   = "{{.ProjectID}}"
  facility     = "{{.Facility}}"

  {{- if .ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.ClusterDomainSuffix}}"
  {{- end }}

  controller_count = {{.ControllerCount}}
  {{- if .ControllerType }}
  controller_type  = "{{ .ControllerType }}"
  {{- end }}

  {{- if .OSArch }}
  os_arch = "{{ .OSArch }}"
  {{- end }}
  {{- if .OSChannel }}
  os_channel = "{{ .OSChannel }}"
  {{- end }}
  {{- if .OSVersion }}
  os_version = "{{ .OSVersion }}"
  {{- end }}

  {{- if .IPXEScriptURL }}
  ipxe_script_url = "{{ .IPXEScriptURL }}"
  {{ end }}
  management_cidrs = {{.ManagementCIDRsRaw}}
  node_private_cidr = "{{.NodePrivateCIDR}}"

  enable_aggregation = {{.EnableAggregation}}

  {{- if .NetworkMTU }}
  network_mtu = {{.NetworkMTU}}
  {{- end }}
  enable_reporting = {{.EnableReporting}}

  {{- if .PodCIDR }}
  pod_cidr = "{{.PodCIDR}}"
  {{- end }}

  {{- if .ServiceCIDR }}
  service_cidr = "{{.ServiceCIDR}}"
  {{- end }}

  {{- if .ReservationIDs }}
    reservation_ids = {
      {{- range $key, $value := .ReservationIDs }}
      {{ $key }} = "{{ $value }}"
      {{- end }}
  }
  {{- end }}

  {{- if .ReservationIDsDefault }}
  reservation_ids_default = "{{.ReservationIDsDefault}}"
  {{- end }}
  {{- if .CertsValidityPeriodHours }}
  certs_validity_period_hours = {{.CertsValidityPeriodHours}}
  {{- end }}

  {{- if .Config.NodesDependOn }}
  nodes_depend_on = [
    {{- range $dep := .Config.NodesDependOn }}
    {{ $dep }},
    {{- end }}
  ]
  {{- end }}
}

{{ range $index, $pool := .WorkerPools }}
module "worker-{{ $pool.Name }}" {
  source = "../lokomotive-kubernetes/packet/flatcar-linux/kubernetes/workers"

  providers = {
    local    = local.default
    template = template.default
    tls      = tls.default
    packet   = packet.default
  }

  ssh_keys  = {{$.SSHPubKeysRaw}}

  cluster_name = "{{$.ClusterName}}"
  tags         = {{$.TagsRaw}}
  project_id   = "{{$.ProjectID}}"
  facility     = "{{$.Facility}}"
  {{- if $.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{$.ClusterDomainSuffix}}"
  {{- end }}

  pool_name    = "{{ $pool.Name }}"
  worker_count = {{ $pool.Count }}
  {{- if $pool.NodeType }}
  type      = "{{ $pool.NodeType }}"
  {{- end }}

  {{- if $.IPXEScriptURL }}
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

  kubeconfig = module.packet-{{ $.ClusterName }}.kubeconfig

  {{- if $pool.Labels }}
  labels = "{{ $pool.Labels }}"
  {{- end }}
  {{- if $pool.Taints }}
  taints = "{{ $pool.Taints }}"
  {{- end }}
  {{- if $.ServiceCIDR }}
  service_cidr = "{{$.ServiceCIDR}}"
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
{{ end }}

{{- if .DNS.Provider.Manual }}
output "dns_entries" {
  value = module.packet-{{.ClusterName}}.dns_entries
}
{{- end }}

{{- if .DNS.Provider.Route53 }}
module "dns" {
  source = "../lokomotive-kubernetes/dns/route53"

  providers = {
    aws = aws.default
  }

  entries = module.packet-{{.ClusterName}}.dns_entries
  aws_zone_id = "{{.DNS.Provider.Route53.ZoneID}}"
}

provider "aws" {
  version = "2.48.0"
  alias   = "default"
  # The Route 53 service doesn't need a specific region to operate, however
  # the AWS Terraform provider needs it and the documentation suggests to use
  # "us-east-1": https://docs.aws.amazon.com/general/latest/gr/r53.html.
  region = "us-east-1"
  {{- if .DNS.Provider.Route53.AWSCredsPath }}
  shared_credentials_file = "{{.DNS.Provider.Route53.AWSCredsPath}}"
  {{- end }}
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

  {{- if .AuthToken }}
  auth_token = "{{.AuthToken}}"
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
  value = module.packet-{{.ClusterName}}.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value     = module.packet-{{.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.packet-{{.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.packet-{{.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value = module.packet-{{.ClusterName}}.calico_values
}
`
