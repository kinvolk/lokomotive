# Lokomotive AWS configuration reference

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Destroying](#destroying)

## Introduction

This configuration reference provides information on configuring a Lokomotive cluster on AWS with all
the configuration options available to the user.

## Prerequisites

* `lokoctl` [installed locally.](../../installer/lokoctl.md)
* `kubectl` installed locally to access the Kubernetes cluster.

## Configuration

To create a Lokomotive cluster, we need to define a configuration.

Example configuration file:

```tf
#myawscluster.lokocfg
variable "dns_zone" {}
variable "route53_zone_id" {}
variable "ssh_public_keys" {}
variable "state_s3_bucket" {}
variable "lock_dynamodb_table" {}
variable "asset_dir" {}
variable "cluster_name" {}
variable "controllers_count" {}
variable "workers_count" {}
variable "state_s3_key" {}
variable "state_s3_region" {}
variable "workers_type" {}
variable "controller_clc_snippets" {}
variable "worker_clc_snippets" {}
variable "region" {}
variable "disk_size" {}
variable "disk_type" {}
variable "disk_iops" {}
variable "worker_disk_size" {}
variable "worker_disk_type" {}
variable "worker_disk_iops" {}
variable "worker_spot_price" {}
variable "lb_http_port" {}
variable "lb_https_port" {}
variable "target_groups" {}
variable "oidc_issuer_url" {}
variable "oidc_client_id" {}
variable "oidc_username_claim" {}
variable "oidc_groups_claim" {}

backend "s3" {
  bucket         = var.state_s3_bucket
  key            = var.state_s3_key
  region         = var.state_s3_region
  dynamodb_table = var.lock_dynamodb_table
}

# backend "local" {
#   path = "path/to/local/file"
#}


cluster "aws" {
  asset_dir = var.asset_dir

  cluster_name = var.cluster_name

  controller_count = var.controllers_count

  controller_type = var.controller_type

  os_channel = "stable"

  os_version = "current"

  tags {
    key1 = "value1"
    key2 = "value2"
  }

  dns_zone = var.dns_zone

  dns_zone_id = route53_zone_id

  enable_csi = true

  expose_nodeports = false

  ssh_pubkeys = var.ssh_public_keys

  certs_validity_period_hours = 8760

  controller_clc_snippets = var.controller_clc_snippets

  region = var.region

  enable_aggregation = true

  enable_tls_bootstrap = true

  disk_size = var.disk_size

  disk_type = var.disk_type

  disk_iops = var.disk_iops

  network_mtu = 1500

  host_cidr = ""

  pod_cidr = "10.2.0.0/16"

  service_cidr = "10.3.0.0/16"

  cluster_domain_suffix = "cluster.local"

  enable_reporting = false

  oidc {
    issuer_url     = var.oidc_issuer_url
    client_id      = var.oidc_client_id
    username_claim = var.oidc_username_claim
    groups_claim   = var.oidc_groups_claim
  }

  worker_pool "my-worker-pool" {
    count = 2

    instance_type = var.workers_type

    ssh_pubkeys = var.ssh_public_keys

    os_channel = "stable"

    os_version = "current"

    labels = "foo=bar,baz=zab"

    taints = "nodeType=storage:NoSchedule"

    disk_size = var.worker_disk_size

    disk_type = var.worker_disk_type

    disk_iops = var.worker_disk_iops

    spot_price = var.worker_spot_price

    lb_http_port = var.lb_http_port

    lb_https_port = var.lb_https_port

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

Example:

The default for instance_type in worker pool block is `t3.small`. If you wish to change the default, then you
define the variable and use it to refer in the cluster configuration.

```tf
variable "custom_worker_type" {
  default = "i3.large"
}

.
.
worker_pool "my-worker-pool" {
  worker_type = var.custom_worker_type
  .
  .
}
.

```

## Attribute reference

| Argument                      | Description                                                                                                                                                                                |     Default     |     Type     | Required |
|-------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------:|:------------:|:--------:|
| `asset_dir`                   | Location where Lokomotive stores cluster assets.                                                                                                                                           |        -        |    string    |   true   |
| `cluster_name`                | Name of the cluster. **NOTE**: It must be unique per DNS Zone and region.                                                                                                                  |        -        |    string    |   true   |
| `tags`                        | Optional details to tag on AWS resources.                                                                                                                                                  |        -        | map(string)  |  false   |
| `os_channel`                  | Flatcar Container Linux AMI channel to install from (stable, beta, alpha, edge).                                                                                                           |    "stable"     |    string    |  false   |
| `os_version`                  | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                       |    "current"    |    string    |  false   |
| `dns_zone`                    | Route 53 DNS Zone.                                                                                                                                                                         |        -        |    string    |   true   |
| `dns_zone_id`                 | Route 53 DNS Zone ID.                                                                                                                                                                      |        -        |    string    |   true   |
| `oidc`                        | OIDC configuration block.                                                                                                                                                                  |        -        |    object    |  false   |
| `oidc.issuer_url`             | URL of the provider which allows the API server to discover public signing keys. Only URLs which use the https:// scheme are accepted.                                                     |        -        |    string    |  false   |
| `oidc.client_id`              | A client id that all tokens must be issued for.                                                                                                                                            |    "gangway"    |    string    |  false   |
| `oidc.username_claim`         | JWT claim to use as the user name.                                                                                                                                                         |     "email"     |    string    |  false   |
| `oidc.groups_claim`           | JWT claim to use as the user’s group.                                                                                                                                                      |    "groups"     |    string    |  false   |
| `enable_csi`                  | Set up IAM role needed for dynamic volumes provisioning to work on AWS                                                                                                                     |      false      |     bool     |  false   |
| `expose_nodeports`            | Expose node ports `30000-32767` in the security group, if set to `true`.                                                                                                                   |      false      |     bool     |  false   |
| `ssh_pubkeys`                 | List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".               |        -        | list(string) |   true   |
| `controller_count`            | Number of controller nodes.                                                                                                                                                                |        1        |    number    |  false   |
| `controller_type`             | AWS instance type for controllers.                                                                                                                                                         |   "t3.small"    |    string    |  false   |
| `controller_clc_snippets`     | Controller Flatcar Container Linux Config snippets.                                                                                                                                        |       []        | list(string) |  false   |
| `region`                      | AWS region to use for deploying the cluster.                                                                                                                                               | "eu-central-1"  |    string    |  false   |
| `enable_aggregation`          | Enable the Kubernetes Aggregation Layer.                                                                                                                                                   |      true       |     bool     |  false   |
| `enable_tls_bootstrap`        | Enable TLS bootstraping for Kubelet.                                                                                                                                                       |      true       |     bool     |  false   |
| `disk_size`                   | Size of the EBS volume in GB.                                                                                                                                                              |       40        |    number    |  false   |
| `disk_type`                   | Type of the EBS volume (e.g. standard, gp2, io1).                                                                                                                                          |      "gp2"      |    string    |  false   |
| `disk_iops`                   | IOPS of the EBS volume (e.g 100).                                                                                                                                                          |        0        |    number    |  false   |
| `network_mtu`                 | Physical Network MTU. When using instance types with Jumbo frames, use 9001.                                                                                                               |      1500       |    number    |  false   |
| `host_cidr`                   | CIDR IPv4 range to assign to EC2 nodes.                                                                                                                                                    |  "10.0.0.0/16"  |    string    |  false   |
| `pod_cidr`                    | CIDR IPv4 range to assign Kubernetes pods.                                                                                                                                                 |  "10.2.0.0/16"  |    string    |  false   |
| `service_cidr`                | CIDR IPv4 range to assign Kubernetes services.                                                                                                                                             |  "10.3.0.0/16"  |    string    |  false   |
| `cluster_domain_suffix`       | Cluster's DNS domain.                                                                                                                                                                      | "cluster.local" |    string    |  false   |
| `enable_reporting`            | Enables usage or analytics reporting to upstream.                                                                                                                                          |      false      |     bool     |  false   |
| `certs_validity_period_hours` | Validity of all the certificates in hours.                                                                                                                                                 |      8760       |    number    |  false   |
| `worker_pool`                 | Configuration block for worker pools. There can be more than one. **NOTE**: worker pool name must be unique per DNS zone and region.                                                       |        -        | list(object) |   true   |
| `worker_pool.count`           | Number of workers in the worker pool. Can be changed afterwards to add or delete workers.                                                                                                  |        -        |    number    |   true   |
| `worker_pool.instance_type`   | AWS instance type for worker nodes.                                                                                                                                                        |   "t3.small"    |    string    |  false   |
| `worker_pool.ssh_pubkeys`     | List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".               |        -        | list(string) |   true   |
| `worker_pool.os_channel`      | Flatcar Container Linux channel to install from (stable, beta, alpha, edge).                                                                                                               |    "stable"     |    string    |  false   |
| `worker_pool.os_version`      | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                       |    "current"    |    string    |  false   |
| `worker_pool.labels`          | Custom labels to assign to worker nodes such as `foo=bar,baz=zab`.                                                                                                                         |        -        |    string    |  false   |
| `worker_pool.taints`          | Taints to assign to worker nodes such as `nodeType=storage:NoSchedule`.                                                                                                                    |        -        |    string    |  false   |
| `worker_pool.disk_size`       | Size of the EBS volume in GB.                                                                                                                                                              |       40        |    number    |  false   |
| `worker_pool.disk_type`       | Type of the EBS volume (e.g. standard, gp2, io1).                                                                                                                                          |      "gp2"      |    string    |  false   |
| `worker_pool.disk_iops`       | IOPS of the EBS volume (e.g 100).                                                                                                                                                          |        0        |    number    |  false   |
| `worker_pool.spot_price`      | Spot price in USD for autoscaling group spot instances. Leave as empty string for autoscaling group to use on-demand instances. Switching in-place from spot to on-demand is not possible. |       ""        |    string    |  false   |
| `worker_pool.target_groups`   | Additional target group ARNs to which worker instances should be added.                                                                                                                    |       []        | list(string) |  false   |
| `worker_pool.lb_http_port`    | Port the load balancer should listen on for HTTP connections.                                                                                                                              |       80        |    number    |  false   |
| `worker_pool.lb_https_port`   | Port the load balancer should listen on for HTTPS connections.                                                                                                                             |       443       |    number    |  false   |
| `worker_pool.clc_snippets`    | CWorker Flatcar Container Linux Config snippets.                                                                                                                                           |       []        | list(string) |  false   |
| `worker_pool.tags`            | Optional details to tag on AWS resources.                                                                                                                                                  |        -        | map(string)  |  false   |


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
