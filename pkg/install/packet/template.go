package packet

var terraformConfigTmpl = `
module "packet-{{.Config.ClusterName}}" {
  source = "../lokomotive-kubernetes/packet/flatcar-linux/kubernetes"

  providers = {
    aws      = "aws.default"
    local    = "local.default"
    null     = "null.default"
    template = "template.default"
    tls      = "tls.default"
    packet   = "packet.default"
  }

  dns_zone    = "{{.Config.DNSZone}}"
  dns_zone_id = "{{.Config.DNSZoneID}}"

  ssh_keys  = {{.SSHPublicKeys}}
  asset_dir = "{{.Config.AssetDir}}"

  cluster_name = "{{.Config.ClusterName}}"
  project_id   = "{{.Config.ProjectID}}"
  facility     = "{{.Config.Facility}}"

  controller_count = "{{.Config.ControllerCount}}"
  {{- if .Config.ControllerType }}
  controller_type  = "{{ .Config.ControllerType }}"
  {{- end }}

  worker_count              = "{{ .WorkerCount }}"
  worker_nodes_hostnames    = "${concat({{- range $index, $pool := .Config.WorkerPools }}{{- if $index }}, {{- end }}"${module.worker-pool-{{- $index }}.worker_nodes_hostname}"{{- end }})}"

  {{- if .Config.OSChannel }}
  os_channel = "{{ .Config.OSChannel }}"
  {{- end }}

  {{- if .Config.IPXEScriptURL }}
  ipxe_script_url = "{{ .Config.IPXEScriptURL }}"
  {{ end }}
  management_cidrs = {{.ManagementCIDRs}}
  node_private_cidr = "{{.Config.NodePrivateCIDR}}"

  enable_aggregation = "{{.Config.EnableAggregation}}"
}

{{ range $index, $pool := .Config.WorkerPools }}
module "worker-pool-{{ $index }}" {
  source = "../lokomotive-kubernetes/packet/flatcar-linux/kubernetes/workers"

  providers = {
    local    = "local.default"
    template = "template.default"
    tls      = "tls.default"
    packet   = "packet.default"
  }

  ssh_keys  = {{$.SSHPublicKeys}}

  cluster_name = "{{$.Config.ClusterName}}"
  project_id   = "{{$.Config.ProjectID}}"
  facility     = "{{$.Config.Facility}}"

  pool_name = "{{ $pool.Name }}"
  count     = "{{ $pool.Count }}"
  {{- if $pool.NodeType }}
  type      = "{{ $pool.NodeType }}"
  {{- end }}

  {{- if $.Config.IPXEScriptURL }}
  ipxe_script_url = "{{ $.Config.IPXEScriptURL }}"
  {{- end }}

  {{- if $pool.OSChannel }}
  os_channel = "{{ $pool.OSChannel }}"
  {{- end }}
  {{- if $pool.OSVersion }}
  os_version = "{{ $pool.OSVersion }}"
  {{- end }}

  kubeconfig = "${module.packet-{{ $.Config.ClusterName }}.kubeconfig}"
}
{{ end }}

provider "aws" {
  version = "~> 1.57.0"
  alias   = "default"

  region                  = "{{.Config.AWSRegion}}"
  {{- if .Config.AWSCredsPath }}
  shared_credentials_file = "{{.Config.AWSCredsPath}}"
  {{- end }}
}

provider "ct" {
  version = "~> 0.3"
}

provider "local" {
  version = "~> 1.0"
  alias   = "default"
}

provider "null" {
  version = "~> 1.0"
  alias   = "default"
}

provider "template" {
  version = "~> 1.0"
  alias   = "default"
}

provider "tls" {
  version = "~> 1.0"
  alias   = "default"
}

provider "packet" {
  version = "~> 1.2"
  alias = "default"

  {{- if .Config.AuthToken }}
  auth_token = "{{.Config.AuthToken}}"
  {{- end }}
}
`
