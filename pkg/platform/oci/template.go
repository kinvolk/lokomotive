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

package oci

var terraformConfigTmpl = `
module "oci-{{.Config.ClusterName}}" {
  source = "../terraform-modules/oci/flatcar-linux/kubernetes"

  compartment_id = "{{.Config.CompartmentID}}"
  tenancy_id     = "{{.Config.TenancyID}}"

  region = "{{.Config.Region}}"

  cluster_name = "{{.Config.ClusterName}}"
  tags         = {{.Tags}}
  dns_zone     = "{{.Config.DNSZone}}"
  dns_zone_id  = "{{.Config.DNSZoneID}}"
  {{- if .Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.Config.ClusterDomainSuffix}}"
  {{- end }}

  {{- if .Config.ExposeNodePorts }}
  expose_nodeports = {{.Config.ExposeNodePorts}}
  {{- end }}

  ssh_keys  = {{$.SSHPublicKeys}}
  asset_dir = "../cluster-assets"

 {{- if .Config.ControllerCount}}
  controller_count = {{.Config.ControllerCount}}
 {{- end }}

  controller_instance_shape  = "{{.Config.ControllerType}}"
  controller_image_id        = "{{.Config.ControllerImageID}}"

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
  {{- if .Config.HostCIDR }}
  host_cidr = "{{.Config.HostCIDR}}"
  {{- end }}

 {{- if .Config.OSChannel }}
  os_channel = "{{.Config.OSChannel}}"
 {{- end }}
 {{- if .Config.OSVersion }}
  os_version = "{{.Config.OSVersion}}"
 {{- end }}
 {{- if .Config.OSArch }}
  os_arch = "{{.Config.OSArch}}"
 {{- end }}

 {{- if ne .ControllerCLCSnippets "null" }}
  controller_clc_snippets = {{.ControllerCLCSnippets}}
 {{- end }}

  enable_aggregation = {{.Config.EnableAggregation}}

  {{- if .Config.DiskSize }}
  disk_size = {{.Config.DiskSize}}
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

  enable_tls_bootstrap    = {{ .Config.EnableTLSBootstrap }}

  {{- if .Config.EncryptPodTraffic }}
  encrypt_pod_traffic = {{.Config.EncryptPodTraffic}}
  {{- end }}

  ignore_x509_cn_check = {{.Config.IgnoreX509CNCheck}}

  conntrack_max_per_core = {{.Config.ConntrackMaxPerCore}}

  {{- if ne .Config.ControllerCPUs 0 }}
  controller_cpus = {{ .Config.ControllerCPUs }}
  {{- end }}
  {{- if ne .Config.ControllerMemory 0 }}
  controller_memory = {{ .Config.ControllerMemory }}
  {{- end }}

  worker_bootstrap_tokens = [
    {{- range $index, $pool := .Config.WorkerPools }}
    module.worker-pool-{{ $index }}.worker_bootstrap_token,
    {{- end }}
  ]
}

{{ range $index, $pool := .Config.WorkerPools }}
module "worker-pool-{{ $index }}" {
  source = "../terraform-modules/oci/flatcar-linux/kubernetes/workers"

  compartment_id = "{{$.Config.CompartmentID}}"
  tenancy_id     = "{{$.Config.TenancyID}}"

  worker_image_id = "{{ $pool.ImageID }}"

  dns_zone = "{{ $.Config.DNSZone }}"

  subnet_id             = module.oci-{{ $.Config.ClusterName }}.subnet_id
  nsg_id                = module.oci-{{ $.Config.ClusterName }}.nsg_id
  kubeconfig            = module.oci-{{ $.Config.ClusterName }}.kubeconfig
  ca_cert               = module.oci-{{ $.Config.ClusterName }}.ca_cert
  apiserver             = module.oci-{{ $.Config.ClusterName }}.apiserver
  enable_tls_bootstrap  = {{ $.Config.EnableTLSBootstrap }}

  {{- if $.Config.ServiceCIDR }}
  service_cidr          = "{{ $.Config.ServiceCIDR }}"
  {{- end }}

  {{- if ne $pool.WorkerCPUs 0 }}
  worker_cpus = {{ $pool.WorkerCPUs }}
  {{- end }}
  {{- if ne $pool.WorkerMemory 0 }}
  worker_memory = {{ $pool.WorkerMemory }}
  {{- end }}


  {{- if $.Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{ $.Config.ClusterDomainSuffix }}"
  {{- end }}

  ssh_keys              = {{ (index $.WorkerpoolCfg $index "ssh_pub_keys") }}
  cluster_name          = "{{ $.Config.ClusterName }}"
  pool_name             = "{{ $pool.Name }}"
  worker_count          = "{{ $pool.Count}}"
  worker_instance_shape = "{{ $pool.InstanceType }}"

  {{- if $pool.OSChannel }}
  os_channel            = "{{ $pool.OSChannel }}"
  {{- end }}

  {{- if $pool.OSVersion }}
  os_version            = "{{ $pool.OSVersion }}"
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

  {{- if $pool.DiskSize }}
  disk_size             = "{{ $pool.DiskSize }}"
  {{- end }}

  {{- if $pool.TargetGroups }}
  target_groups         = {{ (index $.WorkerpoolCfg $index "target_groups") }}
  {{- end }}

  {{- if $pool.CLCSnippets }}
  clc_snippets          = {{ (index $.WorkerpoolCfg $index "clc_snippets") }}
  {{- end }}

  {{- if $pool.Tags }}
  tags                  = {{ index (index $.WorkerpoolCfg $index) "tags" }}
  {{- end }}
}
{{- end }}

provider "oci" {
  region       = "{{.Config.Region}}"
  tenancy_ocid = "{{.Config.TenancyID}}"
  user_ocid    = "{{.Config.User}}"
  fingerprint  = "{{.Config.Fingerprint}}"
  private_key_path = "{{.Config.PrivateKeyPath}}"
}

# Currently this module only supports Route 53 fro DNS
provider "aws" {
  # The Route 53 service doesn't need a specific region to operate, however
  # the AWS Terraform provider needs it and the documentation suggests to use
  # "us-east-1": https://docs.aws.amazon.com/general/latest/gr/r53.html.
  region = "us-east-1"
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
  value     = module.oci-{{.Config.ClusterName}}.pod-checkpointer_values
  sensitive = true
}

output "kube-apiserver_values" {
  value     = module.oci-{{.Config.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.oci-{{.Config.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.oci-{{.Config.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value     = module.oci-{{.Config.ClusterName}}.calico_values
  sensitive = true
}

output "lokomotive_values" {
  value     = module.oci-{{.Config.ClusterName}}.lokomotive_values
  sensitive = true
}

output "bootstrap-secrets_values" {
  value     = module.oci-{{.Config.ClusterName}}.bootstrap-secrets_values
  sensitive = true
}

output "kubeconfig" {
  value     = module.oci-{{.Config.ClusterName}}.kubeconfig-admin
  sensitive = true
}
`
