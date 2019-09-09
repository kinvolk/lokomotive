package aws

var terraformConfigTmpl = `
module "aws-{{.Config.ClusterName}}" {
  source = "../lokomotive-kubernetes/aws/flatcar-linux/kubernetes"

  providers = {
    aws      = "aws.default"
    local    = "local.default"
    null     = "null.default"
    template = "template.default"
    tls      = "tls.default"
  }

  cluster_name = "{{.Config.ClusterName}}"
  dns_zone     = "{{.Config.DNSZone}}"
  dns_zone_id  = "{{.Config.DNSZoneID}}"

  ssh_authorized_key = "{{.SSHAuthorizedKey}}"
  asset_dir          = "../cluster-assets"

  controller_count = "{{.Config.ControllerCount}}"
  controller_type  = "{{.Config.ControllerType}}"

  worker_count = "{{.Config.WorkerCount}}"
  worker_type  = "{{.Config.WorkerType}}"

  os_image = "{{.Config.OSImage}}"

  controller_clc_snippets = {{.ControllerCLCSnippets}}
  worker_clc_snippets     = {{.WorkerCLCSnippets}}

  enable_aggregation = "{{.Config.EnableAggregation}}"
}

provider "aws" {
  version = "~> 2.25.0"
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
