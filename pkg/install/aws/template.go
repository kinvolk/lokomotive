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
  dns_zone     = "{{.Config.DNSZone}}"
  dns_zone_id  = "{{.Config.DNSZoneID}}"
  {{- if .Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.Config.ClusterDomainSuffix}}"
  {{- end }}

  ssh_keys  = {{$.SSHPublicKeys}}
  asset_dir = "../cluster-assets"

  controller_count = {{.Config.ControllerCount}}
  controller_type  = "{{.Config.ControllerType}}"

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

  {{- if .Config.Networking }}
  networking = "{{.Config.Networking}}"
  {{- end }}
  {{- if eq .Config.Networking "calico" }}
  {{- if .Config.NetworkMTU }}
  network_mtu = {{.Config.NetworkMTU}}
  {{- end }}
  enable_reporting = {{.Config.EnableReporting}}
  {{- end }}
  {{- if .Config.PodCIDR }}
  pod_cidr = "{{.Config.PodCIDR}}"
  {{- end }}
  {{- if .Config.ServiceCIDR }}
  service_cidr = "{{.Config.ServiceCIDR}}"
  {{- end }}
  {{- if .Config.HostCIDR }}
  host_cidr = "{{.Config.HostCIDR}}"
  {{- end }}

  os_image = "{{.Config.OSImage}}"

  controller_clc_snippets = {{.ControllerCLCSnippets}}
  worker_clc_snippets     = {{.WorkerCLCSnippets}}

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
  version = "~> 2.31.0"
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
  version = "~> 1.2"
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
`
