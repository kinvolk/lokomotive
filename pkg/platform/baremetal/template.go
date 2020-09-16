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

var terraformConfigTmpl = `terraform {
  {{- if .Backend }}
  {{- if eq .Backend.Type "local" }}
  backend "local" {
    {{- if .Backend.Config.Path }}
    path = "{{ .Backend.Config.Path }}"
    {{- end }}
  }
  {{- end }}
  {{- if eq .Backend.Type "s3" }}
  backend "s3" {
    bucket = "{{ .Backend.Config.Bucket }}"
    key    = "{{ .Backend.Config.Key }}"
    region = "{{ .Backend.Config.Region }}"
    {{- if .Backend.Config.AWSCredsPath }}
    shared_credentials_file = "{{ .Backend.Config.AWSCredsPath }}"
    {{- end }}
    {{- if .Backend.Config.DynamoDBTable }}
    dynamodb_table = "{{ .Backend.Config.DynamoDBTable }}"
    {{- end }}
  }
  {{- end }}
  {{- end }}
}

module "bare-metal-{{.ClusterName}}" {
  source = "../terraform-modules/bare-metal/flatcar-linux/kubernetes"

  # bare-metal
  cluster_name           = "{{.ClusterName}}"
  matchbox_http_endpoint = "{{.MatchboxHTTPEndpoint}}"
  os_channel             = "{{.OSChannel}}"
  os_version             = "{{.OSVersion}}"

  # Disable self hosted kubelet
  disable_self_hosted_kubelet = {{ .DisableSelfHostedKubelet }}

  # Enable TLS Bootstrap
  enable_tls_bootstrap = {{ .EnableTLSBootstrap }}

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

  {{- if .KubeAPIServerExtraFlags }}
  kube_apiserver_extra_flags = [
    {{- range .KubeAPIServerExtraFlags }}
    "{{ . }}",
    {{- end }}
  ]
  {{- end }}

  {{- if .Labels}}
  labels = {
  {{- range $key, $value := .Labels}}
    "{{$key}}" = "{{$value}}",
  {{- end}}
  }
  {{- end}}
}

provider "matchbox" {
  version     = "0.4.1"
  endpoint    = "{{.MatchboxEndpoint}}"
  client_cert = file("{{.MatchboxClientCert}}")
  client_key  = file("{{.MatchboxClientKey}}")
  ca          = file("{{.MatchboxCA}}")
}

provider "ct" {
  version = "0.6.1"
}

provider "local" {
  version = "1.4.0"
}

provider "null" {
  version = "2.1.2"
}

provider "template" {
  version = "2.1.2"
}

provider "tls" {
  version = "2.2.0"
}

# Stub output, which indicates, that Terraform run at least once.
# Used when checking, if we should ask user for confirmation, when
# applying changes to the cluster.
output "initialized" {
  value     = true
  sensitive = true
}

# values.yaml content for all deployed charts.
output "pod-checkpointer_values" {
  value     = module.bare-metal-{{.ClusterName}}.pod-checkpointer_values
  sensitive = true
}

output "kube-apiserver_values" {
  value     = module.bare-metal-{{.ClusterName}}.kube-apiserver_values
  sensitive = true
}

output "kubernetes_values" {
  value     = module.bare-metal-{{.ClusterName}}.kubernetes_values
  sensitive = true
}

output "kubelet_values" {
  value     = module.bare-metal-{{.ClusterName}}.kubelet_values
  sensitive = true
}

output "calico_values" {
  value     = module.bare-metal-{{.ClusterName}}.calico_values
  sensitive = true
}

output "lokomotive_values" {
  value     = module.bare-metal-{{.ClusterName}}.lokomotive_values
  sensitive = true
}

output "bootstrap-secrets_values" {
  value     = module.bare-metal-{{.ClusterName}}.bootstrap-secrets_values
  sensitive = true
}

output "kubeconfig" {
  value     = module.bare-metal-{{.ClusterName}}.kubeconfig-admin
  sensitive = true
}
`
