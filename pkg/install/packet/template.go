package packet

var terraformConfigTmpl = `
module "packet-{{.ClusterName}}" {
  source = "{{.Source}}"

  providers = {
    aws      = "aws.default"
    local    = "local.default"
    null     = "null.default"
    template = "template.default"
    tls      = "tls.default"
    packet   = "packet.default"
  }

  dns_zone    = "{{.DNSZone}}"
  dns_zone_id = "{{.DNSZoneID}}"

  ssh_keys  = {{.SSHKeys}}
  asset_dir = "{{.AssetDir}}"

  cluster_name = "{{.ClusterName}}"
  project_id   = "{{.ProjectID}}"
  facility     = "{{.Facility}}"

  controller_count = "{{.ControllerCount}}"
  controller_type  = "{{.ControllerType}}"
  worker_count     = "{{.WorkerCount}}"
  worker_type      = "{{.WorkerType}}"
}

provider "aws" {
  version = "~> 1.57.0"
  alias   = "default"

  region                  = "{{.AWSRegion}}"
  shared_credentials_file = "{{.AWSCredsPath}}"
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

  auth_token = "{{.AuthToken}}"
}
`
