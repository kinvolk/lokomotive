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

package baremetal

var terraformConfigTmpl = `
module "bare-metal-{{.ClusterName}}" {
  source = "../lokomotive-kubernetes/bare-metal/flatcar-linux/kubernetes"

  providers = {
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
  }

  # bare-metal
  cluster_name           = "{{.ClusterName}}"
  matchbox_http_endpoint = "{{.MatchboxHTTPEndpoint}}"
  os_channel             = "{{.OSChannel}}"
  os_version             = "{{.OSVersion}}"

  # configuration
  cached_install     = "{{.CachedInstall}}"
  k8s_domain_name    = "{{.K8sDomainName}}"
  ssh_keys           = {{.SSHPublicKeys}}
  asset_dir          = "../cluster-assets"

  # machines
  controller_names   = {{.ControllerNames}}
  controller_macs    = {{.ControllerMacs}}
  controller_domains = {{.ControllerDomains}}
  worker_names       = {{.WorkerNames}}
  worker_macs        = {{.WorkerMacs}}
  worker_domains     = {{.WorkerDomains}}
}

provider "matchbox" {
  version     = "~> 0.3"
  endpoint    = "{{.MatchboxEndpoint}}"
  client_cert = file("{{.MatchboxClientCert}}")
  client_key  = file("{{.MatchboxClientKey}}")
  ca          = file("{{.MatchboxCA}}")
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
