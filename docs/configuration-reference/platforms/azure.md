---
title: Lokomotive Azure configuration reference
weight: 10
---

## Introduction

This configuration reference provides information on configuring a Lokomotive cluster on Azure with all the configuration options available to the user.

## Prerequisites

* `lokoctl` [installed locally](../../installer/lokoctl.md).
* `kubectl` installed locally to access the Kubernetes cluster.

## Configuration

To create a Lokomotive cluster, we need to define a configuration.

Example configuration file:

```tf
#myazurecluster.lokocfg
variable "route53_zone_id" {}
variable "asset_dir" {}
variable "cluster_name" {}
variable "controller_type" {}
variable "controller_clc_snippets" {}
variable "worker_type" {}
variable "ssh_pubkeys" {}
variable "tags" {}
variable "controller_count" {}
variable "region" {}
variable "cluster_domain_suffix" {}
variable "oidc_issuer_url" {}
variable "oidc_client_id" {}
variable "oidc_username_claim" {}

backend "s3" {
  bucket         = var.state_s3_bucket
  key            = var.state_s3_key
  region         = var.state_s3_region
  dynamodb_table = var.lock_dynamodb_table
}

# backend "local" {
#   path = "path/to/local/file"
#}


cluster "azure" {
  asset_dir = var.asset_dir

  cluster_name = var.cluster_name

  controller_count = var.controllers_count

  controller_type = var.controller_type

  worker_type = var.worker_type

  tags {
    key1 = "value1"
    key2 = "value2"
  }

  dns {
    zone     = var.route53_zone_id
    provider = "route53"
  }

  region = var.region

  os_image = "flatcar-stable"

  ssh_pubkeys = var.ssh_public_keys

  certs_validity_period_hours = 8760

  controller_clc_snippets = var.controller_clc_snippets

  region = var.region

  enable_aggregation = true

  enable_tls_bootstrap = true

  encrypt_pod_traffic = true

  pod_cidr = "10.2.0.0/16"

  service_cidr = "10.3.0.0/16"

  cluster_domain_suffix = "cluster.local"

  enable_reporting = false

  conntrack_max_per_core = 32768

  oidc {
    issuer_url     = var.oidc_issuer_url
    client_id      = var.oidc_client_id
    username_claim = var.oidc_username_claim
    groups_claim   = var.oidc_groups_claim
  }

  worker_pool "my-worker-pool" {
    count = 2

    ssh_pubkeys = var.ssh_public_keys

    cpu_manager_policy = "none"

    vm_type = "Standard_DS1_v2"

    labels = {
      "testlabel" = ""
    }

    taints = {
      "nodeType" = "storage:NoSchedule"
    }

    os_image = "flatcar-stable"

    priority = "Regular"

    target_groups = var.target_groups

    clc_snippets = var.worker_clc_snippets

    tags = {
      "key" = "value"
    }
  }
}
```

**NOTE**: Should you feel differently about the default values, you can set default values using the `variable`
block in the cluster configuration.

## Attribute reference

| Argument                         | Description                                                                                                                                                                                                                                                                                                  | Default           | Type         | Required |
|----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------|--------------|----------|
| `asset_dir`                      | Location where Lokomotive stores cluster assets.                                                                                                                                                                                                                                                             | -                 | string       | true     |
| `cluster_name`                   | Name of the cluster.                                                                                                                                                                                                                                                                                         | -                 | string       | true     |
| `tags`                           | List of tags that will be propagated to master nodes.                                                                                                                                                                                                                                                        | -                 | map(string)  | false    |
| `controller_count`               | Number of controller nodes.                                                                                                                                                                                                                                                                                  | 1                 | number       | false    |
| `controller_type`                | Azure instance type for controllers. Changing this field on existing cluster will be ignored. To actually apply it, cluster needs to be destroyed and re-created.                                                                                                                                            | "Standard_B2s"    | string       | false    |
| `controller_clc_snippets`        | Controller Flatcar Container Linux Config snippets.                                                                                                                                                                                                                                                          | []                | list(string) | false    |
| `dns`                            | DNS configuration block.                                                                                                                                                                                                                                                                                     | -                 | object       | true     |
| `dns.zone`                       | A DNS zone to use for the cluster. The following format is used for cluster-related DNS records: `<record>.<cluster_name>.<dns_zone>`                                                                                                                                                                        | -                 | string       | true     |
| `dns.provider`                   | DNS provider to use for the cluster. Valid values: `cloudflare`, `route53`, `manual`.                                                                                                                                                                                                                        | -                 | string       | true     |
| `oidc`                           | OIDC configuration block.                                                                                                                                                                                                                                                                                    | -                 | object       | false    |
| `oidc.issuer_url`                | URL of the provider which allows the API server to discover public signing keys. Only URLs which use the https:// scheme are accepted.                                                                                                                                                                       | -                 | string       | false    |
| `oidc.client_id`                 | A client id that all tokens must be issued for.                                                                                                                                                                                                                                                              | "clusterauth"     | string       | false    |
| `oidc.username_claim`            | JWT claim to use as the user name.                                                                                                                                                                                                                                                                           | "email"           | string       | false    |
| `oidc.groups_claim`              | JWT claim to use as the user’s group.                                                                                                                                                                                                                                                                        | "groups"          | string       | false    |
| `region`                         | Azure region to use for deploying the cluster.                                                                                                                                                                                                                                                               | "West Europe"     | string       | false    |
| `ssh_pubkeys`                    | List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".                                                                                                                                 | -                 | list(string) | true     |
| `os_image`                       | Channel for a Container Linux derivative.                                                                                                                                                                                                                                                                    | "flatcar-stable"  | string       | false    |
| `enable_aggregation`             | Enable the Kubernetes Aggregation Layer.                                                                                                                                                                                                                                                                     | true              | bool         | false    |
| `enable_tls_bootstrap`           | Enable TLS bootstraping for Kubelet.                                                                                                                                                                                                                                                                         | true              | bool         | false    |
| `encrypt_pod_traffic`            | Enable in-cluster pod traffic encryption. If true `network_mtu` is reduced by 60 to make room for the encryption header.                                                                                                                                                                                     | false             | bool         | false    |
| `pod_cidr`                       | CIDR IPv4 range to assign Kubernetes pods.                                                                                                                                                                                                                                                                   | "10.2.0.0/16"     | string       | false    |
| `service_cidr`                   | CIDR IPv4 range to assign Kubernetes services.                                                                                                                                                                                                                                                               | "10.3.0.0/16"     | string       | false    |
| `cluster_domain_suffix`          | Cluster's DNS domain.                                                                                                                                                                                                                                                                                        | "cluster.local"   | string       | false    |
| `enable_reporting`               | Enables usage or analytics reporting to upstream.                                                                                                                                                                                                                                                            | false             | bool         | false    |
| `certs_validity_period_hours`    | Validity of all the certificates in hours.                                                                                                                                                                                                                                                                   | 8760              | number       | false    |
| `conntrack_max_per_core`         | Maximum number of entries in conntrack table per CPU on all nodes in the cluster. If you require more fain-grained control over this value, set it to 0 and add CLC snippet setting `net.netfilter.nf_conntrack_max sysctl setting per node pool. See [Flatcar documentation about sysctl] for more details. | 32768             | number       | false    |
| `worker_pool`                    | Configuration block for worker pools. There can be more than one.                                                                                                                                                                                                                                            | -                 | list(object) | true     |
| `worker_pool.count`              | Number of workers in the worker pool. Can be changed afterwards to add or delete workers.                                                                                                                                                                                                                    | 1                 | number       | true     |
| `worker_pool.ssh_pubkeys`        | List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".                                                                                                                                 | -                 | list(string) | true     |
| `worker_pool.clc_snippets`       | Flatcar Container Linux Config snippets for nodes in the worker pool.                                                                                                                                                                                                                                        | []                | list(string) | false    |
| `worker_pool.tags`               | List of tags that will be propagated to nodes in the worker pool.                                                                                                                                                                                                                                            | -                 | map(string)  | false    |
| `worker_pool.cpu_manager_policy` | CPU Manager policy to use. Possible values: `none`, `static`.                                                                                                                                                                                                                                                | "none"            | string       | false    |
| `worker_pool.os_image`           | Channel for a Container Linux derivative.                                                                                                                                                                                                                                                                    | "flatcar-stable"  | string       | false    |
| `worker_pool.labels`             | Map of extra Kubernetes Node labels for worker nodes.                                                                                                                                                                                                                                                        | -                 | map(string)  | false    |
| `worker_pool.taints`             | Map of Taints to assign to worker nodes.                                                                                                                                                                                                                                                                     | -                 | map(string)  | false    |
| `worker.priority`                | Set priority to Spot to use reduced cost surplus capacity, with the tradeoff that instances can be deallocated at any time.                                                                                                                                                                                  | "Regular"         | string       | true     |
| `worker_pool.vm_type`            | Machine type for workers.                                                                                                                                                                                                                                                                                    | "Standard_DS1_v2" | string       | false    |


## Applying

To create the cluster, execute the following command:

```console
lokoctl cluster apply
```

## Destroying

To destroy the Lokomotive cluster, execute the following command:

```console
lokoctl cluster destroy --confirm
```
