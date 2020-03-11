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
  source = "../lokomotive-kubernetes/aws/flatcar-linux/kubernetes"

  providers = {
    aws      = aws.default
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
  }

  cluster_name = "{{.Config.ClusterName}}"
  tags         = {{.Tags}}
  dns_zone     = "{{.Config.DNSZone}}"
  dns_zone_id  = "{{.Config.DNSZoneID}}"
  {{- if .Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.Config.ClusterDomainSuffix}}"
  {{- end }}

  ssh_keys  = {{$.SSHPublicKeys}}
  asset_dir = "../cluster-assets"

	{{- if .Config.ControllerCount}}
  controller_count = {{.Config.ControllerCount}}
	{{- end }}

	{{- if .Config.ControllerType}}
  controller_type  = "{{.Config.ControllerType}}"
	{{- end }}

  worker_count = {{.Config.WorkerCount}}
  {{- if .Config.WorkerType }}
  worker_type  = "{{.Config.WorkerType}}"
  {{- end }}
  {{- if .Config.WorkerPrice }}
  worker_price = "{{.Config.WorkerPrice}}"
  {{- end }}
  {{- if .Config.WorkerTargetGroups }}
  worker_target_groups = {{.WorkerTargetGroups}}
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
	{{- if ne .WorkerCLCSnippets "null" }}
  worker_clc_snippets     = {{.WorkerCLCSnippets}}
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
}

provider "aws" {
  version = "2.48.0"
  alias   = "default"

  region                  = "{{.Config.Region}}"
  {{- if .Config.CredsPath }}
  shared_credentials_file = "{{.Config.CredsPath}}"
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
`
