package baremetal

var terraformConfigTmpl = `
module "bare-metal-{{.ClusterName}}" {
  source = "../lokomotive-kubernetes/bare-metal/flatcar-linux/kubernetes"

  providers = {
    local    = "local.default"
    null     = "null.default"
    template = "template.default"
    tls      = "tls.default"
  }

  # bare-metal
  cluster_name           = "{{.ClusterName}}"
  matchbox_http_endpoint = "{{.MatchboxHTTPEndpoint}}"
  os_channel             = "{{.OSChannel}}"
  os_version             = "{{.OSVersion}}"

  # configuration
  cached_install     = "{{.CachedInstall}}"
  k8s_domain_name    = "{{.K8sDomainName}}"
  ssh_authorized_key = "${file("{{.SSHAuthorizedKey}}")}"
  asset_dir          = "{{.AssetDir}}"

  # machines
  controller_names   = {{.ControllerNames}}
  controller_macs    = {{.ControllerMacs}}
  controller_domains = {{.ControllerDomains}}
  worker_names       = {{.WorkerNames}}
  worker_macs        = {{.WorkerMacs}}
  worker_domains     = {{.WorkerDomains}}
}

provider "matchbox" {
  version     = "~> 0.2"
  endpoint    = "{{.MatchboxEndpoint}}"
  client_cert = "${file("{{.MatchboxClientCert}}")}"
  client_key  = "${file("{{.MatchboxClientKey}}")}"
  ca          = "${file("{{.MatchboxCA}}")}"
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
`
