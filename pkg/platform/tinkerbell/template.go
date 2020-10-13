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

package tinkerbell

var terraformConfigTmpl = `
locals {
  cluster_name = "{{.Name}}"
  dns_zone     = "{{.DNSZone}}"

  # Convert Go slices into Terraform lists.
  ssh_keys = [
  {{- range .SSHPublicKeys}}
    "{{.}}",
  {{- end}}
  ]

  controller_ips = [
  {{- range .ControllerIPAddresses}}
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

  # Calculate DNS names for etcd endpoints.
  controller_etcd_names = [
    for i in range(length(local.controller_ips)) : "${local.cluster_name}-etcd${i}.${local.dns_zone}"
  ]
}

{{- if .Sandbox }}
terraform {
  required_providers {
    tinkerbell = {
      source  = "tinkerbell/tinkerbell"
      version = "0.1.0"
    }
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.2"
    }
  }
}

provider "libvirt" {
  uri = "qemu:///system"
}

module "tinkerbell_sandbox" {
  source = "../terraform-modules/tinkerbell-sandbox"

  name               = local.cluster_name
  hosts_cidr         = "{{ .Sandbox.HostsCIDR }}"
  flatcar_image_path = "{{ .Sandbox.FlatcarImagePath }}"
  pool_path          = "{{ .Sandbox.PoolPath }}"

  ssh_keys = local.ssh_keys

  dns_hosts = concat(
    [for ip in local.controller_ips : {
      hostname = "${local.cluster_name}.${local.dns_zone}"
      ip       = ip
    }],
    [for i, name in local.controller_etcd_names : {
      hostname = name
      ip       = local.controller_ips[i]
    }],
  )
}

provider "tinkerbell" {
  grpc_authority = "${module.tinkerbell_sandbox.provisioner_ip}:42113"
  cert_url       = "http://${module.tinkerbell_sandbox.provisioner_ip}:42114/cert"
}

module "tink_controllers" {
  source = "../terraform-modules/tinkerbell-sandbox/worker"

  count = length(local.controller_ips)

  ip       = local.controller_ips[count.index]
  name     = "${local.cluster_name}-controller-${count.index}"

  sandbox = module.tinkerbell_sandbox

  depends_on = [
    module.tinkerbell_sandbox,
  ]
}

{{ range $index, $pool := .WorkerPools }}
module "tink_worker_{{ $pool.Name }}" {
  source = "../terraform-modules/tinkerbell-sandbox/worker"

  count = length(local.worker_{{ $pool.Name }}_ips)

  ip   = local.worker_{{ $pool.Name }}_ips[count.index]
  name = "${local.cluster_name}-worker-{{ $pool.Name }}-${count.index}"

  sandbox = module.tinkerbell_sandbox

  depends_on = [
    module.tinkerbell_sandbox,
  ]
}

{{- end }}
{{- end }}

module "controllers" {
  source = "../terraform-modules/platforms/tinkerbell/controllers"

  ip_addresses = local.controller_ips

  # Generic configuration.
  asset_dir    = "../cluster-assets"
  cluster_name = local.cluster_name
  ssh_keys     = local.ssh_keys
  dns_zone     = local.dns_zone

  {{- if .ControllerFlatcarInstallBaseURL}}
  flatcar_install_base_url = "{{.ControllerFlatcarInstallBaseURL}}"
  {{- end}}

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

  {{- if .OSChannel}}
  os_channel = "{{.OSChannel}}"
  {{- end }}

  {{- if .OSVersion}}
  os_version = "{{.OSVersion}}"
  {{- end }}

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

  {{- if $.Sandbox }}

  # If sandbox is enabled, set libvirt gateway as DNS server
  # for all nodes, so Kubernetes API FQDN and etcd DNS names
  # can be resolved.
  host_dns_ip = module.tinkerbell_sandbox.gateway

  depends_on = [
    # Create sandbox first, as otherwise Tinkerbell templates
    # and workflows can't be created.
    module.tinkerbell_sandbox,

    # Wait until controllers hardware is registered into Tinkerbell
    # before creating templates and workflows.
    module.tink_controllers,
  ]
  {{- end }}
}

{{- range $index, $pool := .WorkerPools }}

module "worker_{{ $pool.Name }}" {
  source = "../terraform-modules/platforms/tinkerbell/workerpool"

  kubeconfig             = module.controllers.kubeconfig
  cluster_dns_service_ip = module.controllers.cluster_dns_service_ip
  ca_cert                = module.controllers.ca_cert
  apiserver              = module.controllers.apiserver

  cluster_name = local.cluster_name
  name         = "{{ $pool.Name }}"

  {{- if $pool.SSHPublicKeys}}
  ssh_keys = [
  {{- range $pool.SSHPublicKeys}}
    "{{.}}",
  {{- end}}
  ]
  {{- end}}

  {{- if $pool.FlatcarInstallBaseURL}}
  flatcar_install_base_url = "{{$pool.FlatcarInstallBaseURL}}"
  {{- end}}

  ip_addresses = local.worker_{{ $pool.Name }}_ips

  {{- if $.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{$.ClusterDomainSuffix}}"
  {{- end }}

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

  {{- if $.Sandbox }}
  host_dns_ip = module.tinkerbell_sandbox.gateway

  depends_on = [
    # Create sandbox first, as otherwise Tinkerbell templates
    # and workflows can't be created.
    module.tinkerbell_sandbox,

    # Wait until workers hardware is registered into Tinkerbell
    # before creating templates and workflows.
    module.tink_worker_{{ $pool.Name }},
  ]
  {{- end }}
}
{{- end}}

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
