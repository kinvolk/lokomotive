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
  source = "../terraform-modules/bare-metal/flatcar-linux/kubernetes"

  # bare-metal
  cluster_name           = "{{.ClusterName}}"
  matchbox_http_endpoint = "{{.MatchboxHTTPEndpoint}}"
  os_channel             = "{{.OSChannel}}"
  os_version             = "{{.OSVersion}}"

  # Disable self hosted kubelet
  disable_self_hosted_kubelet = {{ .DisableSelfHostedKubelet }}

  {{- if .EncryptPodTraffic }}
  encrypt_pod_traffic = {{ .EncryptPodTraffic }}
  {{- end }}

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

  {{- if .NetworkMTU }}
  network_mtu = {{ .NetworkMTU }}
  {{- end }}

  {{- if .PodCIDR }}
  pod_cidr = "{{.PodCIDR}}"
  {{- end }}

  {{- if .ServiceCIDR }}
  service_cidr = "{{.ServiceCIDR}}"
  {{- end }}

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

  ignore_x509_cn_check   = {{.IgnoreX509CNCheck}}
  conntrack_max_per_core = {{.ConntrackMaxPerCore}}

  {{- if .InstallDisk }}
  install_disk = "{{ .InstallDisk }}"
  {{- end }}

  install_to_smallest_disk = {{ .InstallToSmallestDisk }}

  {{- if .KernelArgs }}
  kernel_args = [
  {{- range $arg := .KernelArgs }}
    "{{ $arg }}",
  {{- end }}
  ]
  {{- end }}

  download_protocol = "{{ .DownloadProtocol }}"

  network_ip_autodetection_method = "{{ .NetworkIPAutodetectionMethod }}"

  {{- if .CLCSnippets}}
  clc_snippets = {
    {{- range $nodeName, $clcSnippetList := .CLCSnippets }}
    "{{ $nodeName }}" = [
    {{- range $clcSnippet := $clcSnippetList }}
      <<EOF
{{ $clcSnippet }}
EOF
      ,
    {{- end }}
    ]
    {{- end }}
  }
  {{- end }}
}

terraform {
  required_providers {
    matchbox = {
      source  = "poseidon/matchbox"
      version = "0.4.1"
    }
  }
}

provider "matchbox" {
  endpoint    = "{{.MatchboxEndpoint}}"
  client_cert = file("{{.MatchboxClientCert}}")
  client_key  = file("{{.MatchboxClientKey}}")
  ca          = file("{{.MatchboxCA}}")
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
