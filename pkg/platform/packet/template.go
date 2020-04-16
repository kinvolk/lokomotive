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
module "packet-{{.PacketConfig.Metadata.ClusterName}}" {
  source = "../lokomotive-kubernetes/packet/flatcar-linux/kubernetes"

  providers = {
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
    packet   = packet.default
  }

  dns_zone    = "{{.PacketConfig.DNS.Zone}}"

  ssh_keys  = {{.SSHPubKeys}}
  asset_dir = "../cluster-assets"

  cluster_name = "{{.PacketConfig.Metadata.ClusterName}}"
  tags         = {{.ControllerTags}}
  project_id   = "{{.PacketConfig.ProjectID}}"
  facility     = "{{.PacketConfig.Facility}}"

  {{- if .LokomotiveConfig.Cluster.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.LokomotiveConfig.Cluster.ClusterDomainSuffix}}"
  {{- end }}

  controller_count = {{.LokomotiveConfig.Controller.Count}}
  {{- if .PacketConfig.Controller.Type }}
  controller_type  = "{{ .PacketConfig.Controller.Type }}"
  {{- end }}

  {{- if .PacketConfig.Flatcar.Arch }}
  os_arch = "{{ .PacketConfig.Flatcar.Arch }}"
  {{- end }}
  {{- if .LokomotiveConfig.Flatcar.Channel }}
  os_channel = "{{ .LokomotiveConfig.Flatcar.Channel }}"
  {{- end }}
  {{- if .LokomotiveConfig.Flatcar.Version }}
  os_version = "{{ .LokomotiveConfig.Flatcar.Version }}"
  {{- end }}

  {{- if .PacketConfig.Flatcar.IPXEScriptURL }}
  ipxe_script_url = "{{ .PacketConfig.Flatcar.IPXEScriptURL }}"
  {{ end }}
  management_cidrs = {{.ManagementCIDRs}}
  node_private_cidr = "{{.PacketConfig.Network.NodePrivateCIDR}}"

  enable_aggregation = {{.LokomotiveConfig.Cluster.EnableAggregation}}

  {{- if .LokomotiveConfig.Network.NetworkMTU }}
  network_mtu = {{.LokomotiveConfig.Network.NetworkMTU}}
  {{- end }}
  enable_reporting = {{.LokomotiveConfig.Network.EnableReporting}}

  {{- if .LokomotiveConfig.Network.PodCIDR }}
  pod_cidr = "{{.LokomotiveConfig.Network.PodCIDR}}"
  {{- end }}

  {{- if .LokomotiveConfig.Network.ServiceCIDR }}
  service_cidr = "{{.LokomotiveConfig.Network.ServiceCIDR}}"
  {{- end }}

  {{- if .PacketConfig.ReservationIDs }}
    reservation_ids = {
      {{- range $key, $value := .PacketConfig.ReservationIDs }}
      {{ $key }} = "{{ $value }}"
      {{- end }}
    }
  {{- end }}

  {{- if .PacketConfig.ReservationIDsDefault }}
  reservation_ids_default = "{{.PacketConfig.ReservationIDsDefault}}"
  {{- end }}
  {{- if .LokomotiveConfig.Cluster.CertsValidityPeriodHours }}
  certs_validity_period_hours = {{.LokomotiveConfig.Cluster.CertsValidityPeriodHours}}
  {{- end }}
}

{{ range $index, $pool := .PacketConfig.WorkerPools }}
module "worker-{{ $pool.Name }}" {
  source = "../lokomotive-kubernetes/packet/flatcar-linux/kubernetes/workers"

  providers = {
    local    = local.default
    template = template.default
    tls      = tls.default
    packet   = packet.default
  }

  ssh_keys  = {{$.SSHPubKeys}}

  cluster_name = "{{$.PacketConfig.Metadata.ClusterName}}"
  tags         = {{$.ControllerTags}}
  project_id   = "{{$.PacketConfig.ProjectID}}"
  facility     = "{{$.PacketConfig.Facility}}"
  {{- if $.LokomotiveConfig.Cluster.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{$.LokomotiveConfig.Cluster.ClusterDomainSuffix}}"
  {{- end }}

  pool_name    = "{{ $pool.Name }}"
  worker_count = {{ $pool.Count }}
  {{- if $pool.NodeType }}
  type      = "{{ $pool.NodeType }}"
  {{- end }}

  {{- if $pool.IPXEScriptURL }}
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

  kubeconfig = module.packet-{{ $.PacketConfig.Metadata.ClusterName }}.kubeconfig

  {{- if $pool.Labels }}
  labels = "{{ $pool.Labels }}"
  {{- end }}
  {{- if $pool.Taints }}
  taints = "{{ $pool.Taints }}"
  {{- end }}
  {{- if $.LokomotiveConfig.Network.ServiceCIDR }}
  service_cidr = "{{$.LokomotiveConfig.Network.ServiceCIDR}}"
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

  {{- if $.PacketConfig.ReservationIDs }}
    reservation_ids = {
      {{- range $key, $value := $.PacketConfig.ReservationIDs }}
      {{ $key }} = "{{ $value }}"
      {{- end }}
    }
  {{- end }}
  {{- if $.PacketConfig.ReservationIDsDefault }}
  reservation_ids_default = "{{$.PacketConfig.ReservationIDsDefault}}"
  {{- end }}
}
{{ end }}

{{- if .PacketConfig.DNS.Provider.Manual }}
output "dns_entries" {
  value = module.packet-{{.PacketConfig.Metadata.ClusterName}}.dns_entries
}
{{- end }}

{{- if .PacketConfig.DNS.Provider.Route53 }}
module "dns" {
  source = "../lokomotive-kubernetes/dns/route53"

  providers = {
    aws = aws.default
  }

  entries = module.packet-{{.PacketConfig.Metadata.ClusterName}}.dns_entries
  aws_zone_id = "{{.PacketConfig.DNS.Provider.Route53.ZoneID}}"
}

provider "aws" {
  version = "2.48.0"
  alias   = "default"
  # The Route 53 service doesn't need a specific region to operate, however
  # the AWS Terraform provider needs it and the documentation suggests to use
  # "us-east-1": https://docs.aws.amazon.com/general/latest/gr/r53.html.
  region = "us-east-1"
  {{- if .PacketConfig.DNS.Provider.Route53.AWSCredsPath }}
  shared_credentials_file = "{{.PacketConfig.DNS.Provider.Route53.AWSCredsPath}}"
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

  {{- if .PacketConfig.AuthToken }}
  auth_token = "{{.PacketConfig.AuthToken}}"
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
  value = module.packet-{{.PacketConfig.Metadata.ClusterName}}.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value     = module.packet-{{.PacketConfig.Metadata.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.packet-{{.PacketConfig.Metadata.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.packet-{{.PacketConfig.Metadata.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value = module.packet-{{.PacketConfig.Metadata.ClusterName}}.calico_values
}
`
