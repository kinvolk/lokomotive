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
  asset_dir = "../cluster-assets"

  cluster_name = "{{.Config.ClusterName}}"
  project_id   = "{{.Config.ProjectID}}"
  facility     = "{{.Config.Facility}}"

  {{- if .Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{.Config.ClusterDomainSuffix}}"
  {{- end }}

  controller_count = "{{.Config.ControllerCount}}"
  {{- if .Config.ControllerType }}
  controller_type  = "{{ .Config.ControllerType }}"
  {{- end }}

  {{- if .Config.OSChannel }}
  os_channel = "{{ .Config.OSChannel }}"
  {{- end }}

  {{- if .Config.IPXEScriptURL }}
  ipxe_script_url = "{{ .Config.IPXEScriptURL }}"
  {{ end }}
  management_cidrs = {{.ManagementCIDRs}}
  node_private_cidr = "{{.Config.NodePrivateCIDR}}"

  enable_aggregation = "{{.Config.EnableAggregation}}"

  {{- if .Config.Networking }}
  networking = "{{.Config.Networking}}"
  {{- end }}

  {{- if eq .Config.Networking "calico" }}
  network_mtu = "{{.Config.NetworkMTU}}"
  enable_reporting = "{{.Config.EnableReporting}}"
  {{- end }}

  {{- if .Config.PodCIDR }}
  pod_cidr = "{{.Config.PodCIDR}}"
  {{- end }}

  {{- if .Config.ServiceCIDR }}
  service_cidr = "{{.Config.ServiceCIDR}}"
  {{- end }}

  {{- if .Config.ReservationIDs }}
    reservation_ids = {
      {{- range $key, $value := .Config.ReservationIDs }}
      {{ $key }} = "{{ $value }}"
      {{- end }}
    }
  {{- end }}

  {{- if .Config.ReservationIDsDefault }}
  reservation_ids_default = "{{.Config.ReservationIDsDefault}}"
  {{- end }}
  {{- if .Config.CertsValidityPeriodHours }}
  certs_validity_period_hours = "{{.Config.CertsValidityPeriodHours}}"
  {{- end }}
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
  {{- if $.Config.ClusterDomainSuffix }}
  cluster_domain_suffix = "{{$.Config.ClusterDomainSuffix}}"
  {{- end }}

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

  {{- if $pool.Labels }}
  labels = "{{ $pool.Labels }}"
  {{- end }}
  {{- if $pool.Taints }}
  taints = "{{ $pool.Taints }}"
  {{- end }}
  {{- if $.Config.ServiceCIDR }}
  service_cidr = "{{$.Config.ServiceCIDR}}"
  {{- end }}

  {{- if $pool.SetupRaid }}
  setup_raid = "{{ $pool.SetupRaid }}"
  {{- end }}
  {{- if $pool.SetupRaidHDD }}
  setup_raid_hdd = "{{ $pool.SetupRaidHDD }}"
  {{- end }}
  {{- if $pool.SetupRaidSSD }}
  setup_raid_ssd = "{{ $pool.SetupRaidSSD }}"
  {{- end }}
  {{- if $pool.SetupRaidSSD }}
  setup_raid_ssd_fs = "{{ $pool.SetupRaidSSDFS }}"
  {{- end }}

  {{- if $.Config.ReservationIDs }}
    reservation_ids = {
      {{- range $key, $value := $.Config.ReservationIDs }}
      {{ $key }} = "{{ $value }}"
      {{- end }}
    }
  {{- end }}
  {{- if $.Config.ReservationIDsDefault }}
  reservation_ids_default = "{{$.Config.ReservationIDsDefault}}"
  {{- end }}
}
{{ end }}

provider "aws" {
  version = "~> 2.31.0"
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

provider "packet" {
  version = "~> 1.4"
  alias = "default"

  {{- if .Config.AuthToken }}
  auth_token = "{{.Config.AuthToken}}"
  {{- end }}
}
`
