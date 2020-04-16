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
module "bare-metal-{{.Config.Metadata.ClusterName}}" {
  source = "../lokomotive-kubernetes/bare-metal/flatcar-linux/kubernetes"

  providers = {
    local    = local.default
    null     = null.default
    template = template.default
    tls      = tls.default
  }

  # bare-metal
  cluster_name           = "{{.Config.Metadata.ClusterName}}"
  matchbox_http_endpoint = "{{.Config.Matchbox.HTTPEndpoint}}"
  os_channel             = "{{.Config.Flatcar.OSChannel}}"
  os_version             = "{{.Config.Flatcar.OSVersion}}"

  # configuration
  cached_install     = "{{.Config.CachedInstall}}"
  k8s_domain_name    = "{{.Config.K8sDomainName}}"
  ssh_keys           = {{.SSHPubKeys}}
  asset_dir          = "../cluster-assets"

  # machines
  controller_names   = {{.ControllerNames}}
  controller_macs    = {{.ControllerMACs}}
  controller_domains = {{.ControllerDomains}}
  worker_names       = {{.WorkerNames}}
  worker_macs        = {{.WorkerMACs}}
  worker_domains     = {{.WorkerDomains}}
}

provider "matchbox" {
  version     = "~> 0.3"
  endpoint    = "{{.Config.Matchbox.Endpoint}}"
  client_cert = file("{{.Config.Matchbox.ClientCertPath}}")
  client_key  = file("{{.Config.Matchbox.ClientKeyPath}}")
  ca          = file("{{.Config.Matchbox.CAPath}}")
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
  value = module.bare-metal-{{.Config.Metadata.ClusterName}}.pod-checkpointer_values
}

output "kube-apiserver_values" {
  value     = module.bare-metal-{{.Config.Metadata.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.bare-metal-{{.Config.Metadata.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.bare-metal-{{.Config.Metadata.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value = module.bare-metal-{{.Config.Metadata.ClusterName}}.calico_values
}
`
