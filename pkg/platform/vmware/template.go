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

package vmware

var terraformConfigTmpl = `
locals {
  cluster_name = "{{.Name}}"
  dns_zone     = "{{.DNS.Zone}}"

  # VMware configuration.
  datacenter      = "{{.Datacenter}}"
  datastore       = "{{.Datastore}}"
  compute_cluster = "{{.ComputeCluster}}"
  network         = "{{.Network}}"
  template        = "{{.Template}}"
  {{- if .Folder }}
  folder = "{{.Folder}}"
  {{- end }}

  controller_ips = [
  {{- range .ControllerIPAddresses}}
    "{{.}}",
  {{- end}}
  ]

  dns_servers = [
  {{- range .DNSServers}}
    "{{.}}",
  {{- end}}
  ]
  
  {{- range $index, $pool := .WorkerPools }}
  worker_{{ $pool.Name }}_ips = [
  {{- range $pool.IPAddresses}}
    "{{.}}",
  {{- end}}
  ]
  {{- end}}

  # Convert Go slices into Terraform lists.
  ssh_keys = [
  {{- range .SSHPublicKeys}}
    "{{.}}",
  {{- end}}
  ]
}

terraform {
  required_providers {
    vsphere = {
      source  = "hashicorp/vsphere"
      version = "1.24.3"
    }
  }
}

provider "vsphere" {}

module "controllers" {
  source = "../terraform-modules/platforms/vmware/controllers"

  # Generic configuration.
  asset_dir     = "../cluster-assets"
  cluster_name  = local.cluster_name
  ssh_keys      = local.ssh_keys
  dns_zone      = local.dns_zone

  datacenter      = local.datacenter
  datastore       = local.datastore
  compute_cluster = local.compute_cluster
  network         = local.network
  template        = local.template

  {{- if .Folder }}
  folder = local.folder
  {{- end }}

  nodes_ips   = local.controller_ips
  hosts_cidr  = "{{.HostsCIDR}}"
  dns_servers = local.dns_servers

  node_count = length(local.controller_ips)

  {{- if .CPUs }}
  cpus_count = {{.CPUs}}
  {{- end }}

  {{- if .Memory }}
  memory = {{.Memory}}
  {{- end }}

  {{- if .DiskSize }}
  disk_size = {{.DiskSize}}
  {{- end }}

  {{- if .ControllerCLCSnippets}}
  clc_snippets = [
  {{- range .ControllerCLCSnippets }}
    <<EOF
{{.}}
EOF
    ,
  {{- end}}
  ]
  {{- end}}

  enable_aggregation = {{.EnableAggregation}}

  {{- if .NetworkMTU }}
  network_mtu = {{.NetworkMTU}}
  {{- end }}

  {{- if .PodCIDR }}
  pod_cidr = "{{.PodCIDR}}"
  {{- end }}

  {{- if .ServiceCIDR }}
  service_cidr = "{{.ServiceCIDR}}"
  {{- end }}

  {{- if .ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.ClusterDomainSuffix}}"
  {{- end }}

  enable_reporting = {{.EnableReporting}}

  {{- if .CertsValidityPeriodHours }}
  certs_validity_period_hours = {{.CertsValidityPeriodHours}}
  {{- end }}

	conntrack_max_per_core = {{.ConntrackMaxPerCore}}

  worker_bootstrap_tokens = concat(
    {{- range $index, $pool := .WorkerPools }}
    module.worker_{{ $pool.Name }}.bootstrap_tokens,
    {{- end }}
  )
}

{{- range $index, $pool := .WorkerPools }}

module "worker_{{ $pool.Name }}" {
  source = "../terraform-modules/platforms/vmware/workerpool"

  kubeconfig             = module.controllers.kubeconfig
  cluster_dns_service_ip = module.controllers.cluster_dns_service_ip
  ca_cert                = module.controllers.ca_cert
  apiserver              = module.controllers.apiserver

  cluster_name = local.cluster_name
  name         = "{{ $pool.Name }}"

  # VMware configuration.
  datacenter      = local.datacenter
  datastore       = local.datastore
  compute_cluster = local.compute_cluster
  network         = local.network

  {{- if $.Folder }}
  folder = local.folder
  {{- end }}

  nodes_ips   = local.worker_{{ $pool.Name }}_ips
  hosts_cidr  = "{{ $.HostsCIDR }}"
  dns_servers = local.dns_servers

  {{- if $pool.Template }}
  template = "{{$pool.Template}}"
  {{- else }}
  template = "{{$.Template}}"
  {{- end }}

  node_count = length(local.worker_{{ $pool.Name }}_ips)

  {{- if $pool.CPUs }}
  cpus_count = {{$pool.CPUs}}
  {{- end }}

  {{- if $pool.Memory }}
  memory = {{$pool.Memory}}
  {{- end }}

  {{- if $pool.DiskSize }}
  disk_size = {{$pool.DiskSize}}
  {{- end }}

  {{- if $pool.SSHPublicKeys}}
  ssh_keys = [
  {{- range $pool.SSHPublicKeys}}
    "{{.}}",
  {{- end}}
  ]
  {{- end}}

  {{- if $pool.CLCSnippets}}
  clc_snippets = [
  {{- range $pool.CLCSnippets}}
    <<EOF
{{.}}
EOF
    ,
  {{- end}}
  ]
  {{- end}}

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
  {{- end }}

}
{{- end}}

module "dns" {
  source = "../terraform-modules/dns/{{.DNS.Provider}}"

  cluster_name             = local.cluster_name
  controllers_public_ipv4  = module.controllers.controllers_private_ipv4
  controllers_private_ipv4 = module.controllers.controllers_private_ipv4
  dns_zone                 = "{{ .DNS.Zone }}"
}

{{- if eq .DNS.Provider "manual" }}

output "dns_entries" {
  value = module.dns.entries
}
{{- end }}

# Stub output, which indicates, that Terraform run at least once.
# Used when checking, if we should ask user for confirmation, when
# applying changes to the cluster.
output "initialized" {
  value			= true
	sensitive = true
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value 		= module.controllers.pod-checkpointer_values
	sensitive = true
}

output "kube-apiserver_values" {
  value     = module.controllers.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.controllers.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.controllers.kubelet_values
  sensitive = true
}

output "calico_values" {
  value 		= module.controllers.calico_values
	sensitive = true
}

output "lokomotive_values" {
  value     = module.controllers.lokomotive_values
  sensitive = true
}

output "bootstrap-secrets_values" {
  value     = module.controllers.bootstrap-secrets_values
  sensitive = true
}
`
