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
module "aws-{{.ClusterName}}" {
  source = "../lokomotive-kubernetes/aws/flatcar-linux/kubernetes"

  providers = {
    aws      = aws.default
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
  }

  cluster_name = "{{.ClusterName}}"
  tags         = {{.TagsRaw}}
  dns_zone     = "{{.DNSZone}}"
  dns_zone_id  = "{{.DNSZoneID}}"
  {{- if .ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.ClusterDomainSuffix}}"
  {{- end }}

  {{- if .ExposeNodePorts }}
  expose_nodeports = {{.ExposeNodePorts}}
  {{- end }}

  ssh_keys  = {{$.SSHPubKeysRaw}}
  asset_dir = "../cluster-assets"

 {{- if .ControllerCount}}
  controller_count = {{.ControllerCount}}
 {{- end }}

 {{- if .ControllerType}}
  controller_type  = "{{.ControllerType}}"
 {{- end }}

	# Do not allow creation of workers apart from using worker pools.
  worker_count = 0

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
  {{- if .HostCIDR }}
  host_cidr = "{{.HostCIDR}}"
  {{- end }}

 {{- if .OSName }}
  os_name = "{{.OSName}}"
 {{- end }}
 {{- if .OSChannel }}
  os_channel = "{{.OSChannel}}"
 {{- end }}
 {{- if .OSVersion }}
  os_version = "{{.OSVersion}}"
 {{- end }}

 {{- if ne .ControllerCLCSnippetsRaw "null" }}
  controller_clc_snippets = {{.ControllerCLCSnippetsRaw}}
 {{- end }}

  enable_aggregation = {{.EnableAggregation}}

  {{- if .DiskSize }}
  disk_size = {{.DiskSize}}
  {{- end }}
  {{- if .DiskType }}
  disk_type = "{{.DiskType}}"
  {{- end }}
  {{- if .DiskIOPS }}
  disk_iops = {{.DiskIOPS}}
  {{- end }}

  {{- if .CertsValidityPeriodHours }}
  certs_validity_period_hours = {{.CertsValidityPeriodHours}}
  {{- end }}
}

{{ range $index, $pool := .WorkerPools }}
module "worker-pool-{{ $index }}" {
  source = "../lokomotive-kubernetes/aws/flatcar-linux/kubernetes/workers"

  providers = {
    aws      = aws.default
  }

  vpc_id                = module.aws-{{ $.ClusterName }}.vpc_id
  subnet_ids            = flatten([module.aws-{{ $.ClusterName }}.subnet_ids])
  security_groups       = module.aws-{{ $.ClusterName }}.worker_security_groups
  kubeconfig            = module.aws-{{ $.ClusterName }}.kubeconfig

  {{- if $.ServiceCIDR }}
  service_cidr          = "{{ $.ServiceCIDR }}"
  {{- end }}

  {{- if $.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{ $.ClusterDomainSuffix }}"
  {{- end }}

  ssh_keys              = {{ index (index $.WorkerPoolsListRaw $index) "ssh_pub_keys" }}
  name                  = "{{ $pool.Name }}"
  worker_count          = "{{ $pool.Count}}"
  os_name               = "flatcar"
  {{- if $pool.InstanceType }}
  instance_type         = "{{ $pool.InstanceType }}"
  {{- end }}

  {{- if $pool.OSChannel }}
  os_channel            = "{{ $pool.OSChannel }}"
  {{- end }}

  {{- if $pool.OSVersion }}
  os_version            = "{{ $pool.OSVersion }}"
  {{- end }}

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
  target_groups         = "{{ index (index $.WorkerPoolsListRaw $index) "target_groups" }}"
  {{- end }}

  {{- if ne (index (index $.WorkerPoolsListRaw $index) "clc_snippets") "null" }}
  clc_snippets          = {{ index (index $.WorkerPoolsListRaw $index) "clc snippets" }}
  {{- end }}

  {{- if $pool.Tags }}
  tags                  = {{ index (index $.WorkerPoolsListRaw $index) "tags" }}
  {{- end }}
}
{{- end }}

provider "aws" {
  version = "2.48.0"
  alias   = "default"

  region                  = "{{.Region}}"
  {{- if .CredsPath }}
  shared_credentials_file = "{{.CredsPath}}"
  {{- end }}
}

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

# Stub output, which indicates, that Terraform run at least once.
# Used when checking, if we should ask user for confirmation, when
# applying changes to the cluster.
output "initialized" {
  value = true
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value = module.aws-{{.ClusterName}}.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value     = module.aws-{{.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.aws-{{.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.aws-{{.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value = module.aws-{{.ClusterName}}.calico_values
}
`
