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
module "aws-{{.AWSConfig.Metadata.ClusterName}}" {
  source = "../lokomotive-kubernetes/aws/flatcar-linux/kubernetes"

  providers = {
    aws      = aws.default
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
  }

  cluster_name = "{{.AWSConfig.Metadata.ClusterName}}"
  tags         = {{.ControllerTags}}
  dns_zone     = "{{.AWSConfig.DNSZone}}"
  dns_zone_id  = "{{.AWSConfig.DNSZoneID}}"
  {{- if .LokomotiveConfig.Cluster.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.LokomotiveConfig.Cluster.ClusterDomainSuffix}}"
  {{- end }}

  {{- if .AWSConfig.ExposeNodePorts }}
  expose_nodeports = {{.AWSConfig.ExposeNodePorts}}
  {{- end }}

  ssh_keys  = {{.SSHPubKeys}}
  asset_dir = "../cluster-assets"

 {{- if .LokomotiveConfig.Controller.Count}}
  controller_count = {{.LokomotiveConfig.Controller.Count}}
 {{- end }}

 {{- if .AWSConfig.Controller.Type}}
  controller_type  = "{{.AWSConfig.Controller.Type}}"
 {{- end }}

	# Do not allow creation of workers apart from using worker pools.
  worker_count = 0

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
  {{- if .AWSConfig.Network.HostCIDR }}
  host_cidr = "{{.AWSConfig.Network.HostCIDR}}"
  {{- end }}

 {{- if .AWSConfig.Flatcar.OSName }}
  os_name = "{{.AWSConfig.Flatcar.OSName}}"
 {{- end }}
 {{- if .LokomotiveConfig.Flatcar.Channel }}
  os_channel = "{{.LokomotiveConfig.Flatcar.Channel}}"
 {{- end }}
 {{- if .LokomotiveConfig.Flatcar.Version }}
  os_version = "{{.LokomotiveConfig.Flatcar.Version}}"
 {{- end }}

 {{- if ne .ControllerCLCSnippets "null" }}
  controller_clc_snippets = {{.ControllerCLCSnippets}}
 {{- end }}

  enable_aggregation = {{.LokomotiveConfig.Cluster.EnableAggregation}}

  {{- if .AWSConfig.Disk.Size }}
  disk_size = {{.AWSConfig.Disk.Size}}
  {{- end }}
  {{- if .AWSConfig.Disk.Type }}
  disk_type = "{{.AWSConfig.Disk.Type}}"
  {{- end }}
  {{- if .AWSConfig.Disk.IOPS }}
  disk_iops = {{.AWSConfig.Disk.IOPS}}
  {{- end }}

  {{- if .LokomotiveConfig.Cluster.CertsValidityPeriodHours }}
  certs_validity_period_hours = {{.LokomotiveConfig.Cluster.CertsValidityPeriodHours}}
  {{- end }}
}

{{ range $index, $pool := .AWSConfig.WorkerPools }}
module "worker-pool-{{ $index }}" {
  source = "../lokomotive-kubernetes/aws/flatcar-linux/kubernetes/workers"

  providers = {
    aws      = aws.default
  }

  vpc_id                = module.aws-{{ $.AWSConfig.Metadata.ClusterName }}.vpc_id
  subnet_ids            = flatten([module.aws-{{ $.AWSConfig.Metadata.ClusterName }}.subnet_ids])
  security_groups       = module.aws-{{ $.AWSConfig.Metadata.ClusterName }}.worker_security_groups
  kubeconfig            = module.aws-{{ $.AWSConfig.Metadata.ClusterName }}.kubeconfig

  {{- if $.LokomotiveConfig.Network.ServiceCIDR }}
  service_cidr          = "{{ $.LokomotiveConfig.Network.ServiceCIDR }}"
  {{- end }}

  {{- if $.LokomotiveConfig.Cluster.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{ $.LokomotiveConfig.Cluster.ClusterDomainSuffix }}"
  {{- end }}

  ssh_keys              = {{ index (index $.WorkerPoolsList $index) "ssh_pub_keys" }}
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
  target_groups         = "{{ index (index $.WorkerPoolsList $index) "target_groups" }}"
  {{- end }}

  {{- if ne (index (index $.WorkerPoolsList $index) "clc_snippets") "null" }}
  clc_snippets          = {{ index (index $.WorkerPoolsList $index) "clc_snippets" }}
  {{- end }}

  {{- if $pool.Tags }}
  tags                  = {{ index (index $.WorkerPoolsList $index) "tags" }}
  {{- end }}
}
{{- end }}

provider "aws" {
  version = "2.48.0"
  alias   = "default"

  region                  = "{{.AWSConfig.Region}}"
  {{- if .AWSConfig.CredsPath }}
  shared_credentials_file = "{{.AWSConfig.CredsPath}}"
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
  value = module.aws-{{.AWSConfig.Metadata.ClusterName}}.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value     = module.aws-{{.AWSConfig.Metadata.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.aws-{{.AWSConfig.Metadata.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.aws-{{.AWSConfig.Metadata.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value = module.aws-{{.AWSConfig.Metadata.ClusterName}}.calico_values
}
`
