package packet

var terraformConfigTmpl = `
module "packet-{{.Config.ClusterName}}" {
  source = "{{.Source}}"

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
  controller_type  = "{{.Config.ControllerType}}"
  worker_count     = "{{.Config.WorkerCount}}"
  worker_type      = "{{.Config.WorkerType}}"

  ipxe_script_url = "{{.Config.IPXEScriptURL}}"
  management_cidrs = {{.ManagementCIDRs}}
  node_private_cidr = "{{.Config.NodePrivateCIDR}}"
}

provider "aws" {
  version = "~> 1.57.0"
  alias   = "default"

  region                  = "{{.Config.AWSRegion}}"
  shared_credentials_file = "{{.Config.AWSCredsPath}}"
}

provider "ct" {
  version = "0.3.0"
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

  auth_token = "{{.Config.AuthToken}}"
}
`
