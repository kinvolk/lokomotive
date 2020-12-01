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

package aws

var terraformConfigTmpl = `
module "aws-{{.Config.ClusterName}}" {
  source = "../terraform-modules/aws/flatcar-linux/kubernetes"

  cluster_name = "{{.Config.ClusterName}}"
  tags         = {{.Tags}}
  dns_zone     = "{{.Config.DNSZone}}"
  dns_zone_id  = "{{.Config.DNSZoneID}}"
  enable_csi   = {{.Config.EnableCSI}}
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

 {{- if .Config.ControllerType}}
  controller_type  = "{{.Config.ControllerType}}"
 {{- end }}

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

 {{- if .Config.OSName }}
  os_name = "{{.Config.OSName}}"
 {{- end }}
 {{- if .Config.OSChannel }}
  os_channel = "{{.Config.OSChannel}}"
 {{- end }}
 {{- if .Config.OSVersion }}
  os_version = "{{.Config.OSVersion}}"
 {{- end }}

 {{- if ne .ControllerCLCSnippets "null" }}
  controller_clc_snippets = {{.ControllerCLCSnippets}}
 {{- end }}

  enable_aggregation = {{.Config.EnableAggregation}}

  {{- if .Config.DiskSize }}
  disk_size = {{.Config.DiskSize}}
  {{- end }}
  {{- if .Config.DiskType }}
  disk_type = "{{.Config.DiskType}}"
  {{- end }}
  {{- if .Config.DiskIOPS }}
  disk_iops = {{.Config.DiskIOPS}}
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

  worker_bootstrap_tokens = [
    {{- range $index, $pool := .Config.WorkerPools }}
    module.worker-pool-{{ $index }}.worker_bootstrap_token,
    {{- end }}
  ]
}

{{ range $index, $pool := .Config.WorkerPools }}
module "worker-pool-{{ $index }}" {
  source = "../terraform-modules/aws/flatcar-linux/kubernetes/workers"

  vpc_id                = module.aws-{{ $.Config.ClusterName }}.vpc_id
  subnet_ids            = flatten([module.aws-{{ $.Config.ClusterName }}.subnet_ids])
  security_groups       = module.aws-{{ $.Config.ClusterName }}.worker_security_groups
  kubeconfig            = module.aws-{{ $.Config.ClusterName }}.kubeconfig
  ca_cert               = module.aws-{{ $.Config.ClusterName }}.ca_cert
  apiserver             = module.aws-{{ $.Config.ClusterName }}.apiserver
  enable_tls_bootstrap  = {{ $.Config.EnableTLSBootstrap }}

  {{- if $.Config.ServiceCIDR }}
  service_cidr          = "{{ $.Config.ServiceCIDR }}"
  {{- end }}

  {{- if $.Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{ $.Config.ClusterDomainSuffix }}"
  {{- end }}

  ssh_keys              = {{ (index $.WorkerpoolCfg $index "ssh_pub_keys") }}
  cluster_name          = "{{ $.Config.ClusterName }}"
  pool_name             = "{{ $pool.Name }}"
  worker_count          = "{{ $pool.Count}}"
  os_name               = "flatcar"
  {{- if $pool.InstanceType }}
  instance_type         = "{{ $pool.InstanceType }}"
  {{- end }}

  lb_arn = module.aws-{{ $.Config.ClusterName }}.nlb_arn
  {{- if $pool.LBHTTPPort }}
  lb_http_port = {{ $pool.LBHTTPPort }}
  {{- end }}
  {{- if $pool.LBHTTPSPort }}
  lb_https_port = {{ $pool.LBHTTPSPort }}
  {{- end }}

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

  {{- if $pool.DiskType }}
  disk_type             = "{{ $pool.DiskType }}"
  {{- end }}

  {{- if $pool.DiskIOPS }}
  disk_iops             = "{{ $pool.DiskIOPS }}"
  {{- end }}

  {{- if $pool.SpotPrice }}
  spot_price            = "{{ $pool.SpotPrice }}"
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

provider "aws" {
  region                  = "{{.Config.Region}}"
  {{- if .Config.CredsPath }}
  shared_credentials_file = "{{.Config.CredsPath}}"
  {{- end }}
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
  value     = module.aws-{{.Config.ClusterName}}.pod-checkpointer_values
  sensitive = true
}

output "kube-apiserver_values" {
  value     = module.aws-{{.Config.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.aws-{{.Config.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.aws-{{.Config.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value     = module.aws-{{.Config.ClusterName}}.calico_values
  sensitive = true
}

output "lokomotive_values" {
  value     = module.aws-{{.Config.ClusterName}}.lokomotive_values
  sensitive = true
}

output "bootstrap-secrets_values" {
  value     = module.aws-{{.Config.ClusterName}}.bootstrap-secrets_values
  sensitive = true
}

output "kubeconfig" {
  value     = module.aws-{{.Config.ClusterName}}.kubeconfig-admin
  sensitive = true
}
`
