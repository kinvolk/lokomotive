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

package kvmlibvirt

var terraformConfigTmpl = `
module "kvm-libvirt-{{.Config.ClusterName}}" {
  source = "../terraform-modules/kvm-libvirt/flatcar-linux/kubernetes"

  cluster_name = "{{.Config.ClusterName}}"

  ssh_keys = {{ StringsJoin .Config.SSHPubKeys "\", \"" | printf "[%q]" }}

  asset_dir = "../cluster-assets"

  os_image = "{{.Config.OSImage}}"

  machine_domain = "{{.Config.MachineDomain}}"

  {{- if .Config.NodeIpPool}}
  node_ip_pool   = "{{.Config.NodeIpPool}}"
  {{- end }}

  {{- if .Config.ControllerCount}}
  controller_count = {{.Config.ControllerCount}}
  {{- end }}

  {{- if .Config.ControllerVirtualCPUs}}
  virtual_cpus = {{ .Config.ControllerVirtualCPUs}}
  {{- end }}

  {{- if .Config.ControllerVirtualMemory}}
  virtual_memory {{ .Config.ControllerVirtualMemory}}
  {{- end }}

  {{- if .Config.NetworkMTU }}
  network_mtu = {{.Config.NetworkMTU}}
  {{- end }}

  {{- if .Config.NetworkIpAutodetectionMethod }}
  network_ip_autodetection_method = {{ .Config.NetworkIpAutodetectionMethod }}
  {{- end }}

  {{- if .Config.PodCidr }}
  pod_cidr = {{.Config.PodCidr }}
  {{- end }}
  {{- if .Config.ServiceCidr }}
  service_cidr = {{.Config.ServiceCidr }}
  {{- end }}

  {{- if .Config.ClusterDomainSuffix }}
  cluster_domain_suffix = {{.Config.ClusterDomainSuffix }}
  {{- end }}

  enable_reporting   = {{.Config.EnableReporting}}
  enable_aggregation = {{.Config.EnableAggregation}}

  {{- if .Config.ControllerCLCSnippets }}
  controller_clc_snippets = {{ StringsJoin .Config.ControllerCLCSnippets "\", \"" | printf "[%q]" }}
  {{- end }}

  disable_self_hosted_kubelet = {{ .Config.DisableSelfHostedKubelet }}

  {{- if .Config.KubeAPIServerExtraFlags }}
  kube_apiserver_extra_flags = {{ StringsJoin .Config.KubeAPIServerExtraFlags "\", \"" | printf "[%q]" }}
  {{- end }}

  {{- if .Config.CertsValidityPeriodHours }}
  certs_validity_period_hours = {{.Config.CertsValidityPeriodHours}}
  {{- end }}
}

{{ range $index, $pool := .Config.WorkerPools }}
module "worker-pool-{{ $index }}" {
  source = "../terraform-modules/kvm-libvirt/flatcar-linux/kubernetes/workers"

  ssh_keys = {{ StringsJoin $.Config.SSHPubKeys "\", \"" | printf "[%q]" }}
  machine_domain        = "{{$.Config.MachineDomain}}"
  cluster_name          = "{{$.Config.ClusterName}}"
  {{- if $.Config.ClusterDomainSuffix }}
  cluster_domain_suffix = {{$.Config.ClusterDomainSuffix }}
  {{- end }}
  {{- if $.Config.ServiceCidr }}
  service_cidr = {{$.Config.ServiceCidr }}
  {{- end }}

  libvirtpool           = module.kvm-libvirt-{{$.Config.ClusterName}}.libvirtpool
  libvirtbaseid         = module.kvm-libvirt-{{$.Config.ClusterName}}.libvirtbaseid
  kubeconfig            = module.kvm-libvirt-{{$.Config.ClusterName}}.kubeconfig

  pool_name             = "{{ $pool.Name }}"
  worker_count          = "{{ $pool.Count}}"

  {{- if $pool.VirtualCPUs }}
  virtual_cpus = {{ $pool.VirtualCPUs }}
  {{- end }}

  {{- if $pool.VirtualMemory }}
  virtual_memory {{ $pool.VirtualMemory }}
  {{- end }}

  {{- if $pool.Labels }}
  labels = {{ $pool.Labels }}
  {{- end }}

  {{- if $pool.CLCSnippets }}
  clc_snippets = {{ StringsJoin $pool.CLCSnippets "\", \"" | printf "[%q]" }}
  {{- end }}
}
{{- end }}

provider "libvirt" {
  uri     = "qemu:///system"
  version = "~> 0.6.0"
}

provider "ct" {
  version = "~> 0.5.0"
}

provider "local" {
  version = "~> 1.2"
}

provider "null" {
  version = "~> 2.1"
}

provider "template" {
  version = "~> 2.1"
}

provider "tls" {
  version = "~> 2.0"
}

provider "random" {
  version = "~> 2.2"
}


# Stub output, which indicates, that Terraform run at least once.
# Used when checking, if we should ask user for confirmation, when
# applying changes to the cluster.
output "initialized" {
  value     = true
  sensitive = true
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value     = module.kvm-libvirt-{{.Config.ClusterName}}.pod-checkpointer_values
  sensitive = true
}

output "kube-apiserver_values" {
  value     = module.kvm-libvirt-{{.Config.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.kvm-libvirt-{{.Config.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.kvm-libvirt-{{.Config.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value     = module.kvm-libvirt-{{.Config.ClusterName}}.calico_values
  sensitive = true
}
`
