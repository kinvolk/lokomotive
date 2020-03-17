# Lokomotive AWS configuration reference

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Uninstalling](#uninstalling)

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
variable "target_groups" {}

backend "s3" {
  bucket         = var.state_s3_bucket
  key            = var.state_s3_key
  region         = var.state_s3_region
  dynamodb_table = var.lock_dynamodb_table
}

# backed "local" {
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

  ssh_pubkeys = var.ssh_public_keys

  certs_validity_period_hours = 8760

  controller_clc_snippets = var.controller_clc_snippets

  region = var.region

  enable_aggreation = true

  disk_size = var.disk_size

  disk_type = var.disk_type

  disk_iops = var.disk_iops

  network_mtu = 1480

  host_cidr = ""

  pod_cidr = "10.2.0.0/16"

  service_cidr = "10.3.0.0/16"

  cluster_domain_suffix = "cluster.local"

  enable_reporting = false

  worker_pool "my-worker-pool {
    count = 2

    instance_type = var.workers_type

    ssh_pubkeys = var.ssh_public_keys

    os_channel = "stable"

    os_version = "current"

    disk_size = var.worker_disk_size

    disk_type = var.worker_disk_type

    disk_iops = var.worker_disk_iops

    spot_price = var.worker_spot_price

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

| Argument                      | Description                                                                                                                                                                                | Default         | Required |
|-------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------:|:--------:|
| `asset_dir`                   | Location where Lokomotive stores cluster assets.                                                                                                                                           | -               | true     |
| `cluster_name`                | Name of the cluster.                                                                                                                                                                       | -               | true     |
| `tags`                        | Optional details to tag on AWS resources.                                                                                                                                                  | -               | false    |
| `os_channel`                  | Flatcar Container Linux AMI channel to install from (stable, beta, alpha, edge).                                                                                                           | "stable"        | false    |
| `os_version`                  | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                       | "current"       | false    |
| `dns_zone`                    | Route 53 DNS Zone.                                                                                                                                                                         | -               | true     |
| `dns_zone_id`                 | Route 53 DNS Zone ID.                                                                                                                                                                      | -               | true     |
| `ssh_pubkeys`                 | SSH public keys for user `core`.                                                                                                                                                           | -               | true     |
| `controller_count`            | Number of controller nodes.                                                                                                                                                                | 1               | false    |
| `controller_type`             | AWS instance type for controllers.                                                                                                                                                         | "t3.small"      | false    |
| `controller_clc_snippets`     | Controller Flatcar Container Linux Config snippets.                                                                                                                                        | []              | false    |
| `region`                      | AWS region to use for deploying the cluster.                                                                                                                                               | "eu-central-1"  | false    |
| `enable_aggregation`          | Enable the Kubernetes Aggregation Layer.                                                                                                                                                   | true            | false    |
| `disk_size`                   | Size of the EBS volume in GB.                                                                                                                                                              | 40              | false    |
| `disk_type`                   | Type of the EBS volume (e.g. standard, gp2, io1).                                                                                                                                          | "gp2"           | false    |
| `disk_iops`                   | IOPS of the EBS volume (e.g 100).                                                                                                                                                          | 0               | false    |
| `network_mtu`                 | CNI interface MTU. Use 8981 if using instances types with Jumbo frames.                                                                                                                    | 1480            | false    |
| `host_cidr`                   | CIDR IPv4 range to assign to EC2 nodes.                                                                                                                                                    | "10.0.0.0/16"   | false    |
| `pod_cidr`                    | CIDR IPv4 range to assign Kubernetes pods.                                                                                                                                                 | "10.2.0.0/16"   | false    |
| `service_cidr`                | CIDR IPv4 range to assign Kubernetes services.                                                                                                                                             | "10.3.0.0/16"   | false    |
| `cluster_domain_suffix`       | Cluster's DNS domain.                                                                                                                                                                      | "cluster.local" | false    |
| `enable_reporting`            | Enables usage or analytics reporting to upstream.                                                                                                                                          | false           | false    |
| `certs_validity_period_hours` | Validity of all the certificates in hours.                                                                                                                                                 | 8760            | false    |
| `worker_pool`                 | Configuration block for worker pools. There can be more than one.                                                                                                                          | -               | true     |
| `worker_pool.count`           | Number of workers in the worker pool. Can be changed afterwards to add or delete workers.                                                                                                  | -               | true     |
| `worker_pool.instance_type`   | AWS instance type for worker nodes.                                                                                                                                                        | "t3.small"      | false    |
| `worker_pool.ssh_pubkeys`     | SSH public keys for user `core`.                                                                                                                                                           | -               | true     |
| `worker_pool.os_channel`      | Flatcar Container Linux channel to install from (stable, beta, alpha, edge).                                                                                                               | "stable"        | false    |
| `worker_pool.os_version`      | Flatcar Container Linux version to install. Version such as "2303.3.1" or "current".                                                                                                       | "current"       | false    |
| `worker_pool.disk_size`       | Size of the EBS volume in GB.                                                                                                                                                              | 40              | false    |
| `worker_pool.disk_type`       | Type of the EBS volume (e.g. standard, gp2, io1).                                                                                                                                          | "gp2"           | false    |
| `worker_pool.disk_iops`       | IOPS of the EBS volume (e.g 100).                                                                                                                                                          | 0               | false    |
| `worker_pool.spot_price`      | Spot price in USD for autoscaling group spot instances. Leave as empty string for autoscaling group to use on-demand instances. Switching in-place from spot to on-demand is not possible. | ""              | false    |
| `worker_pool.target_groups`   | Additional target group ARNs to which worker instances should be added.                                                                                                                    | []              | false    |
| `worker_pool.clc_snippets`    | CWorker Flatcar Container Linux Config snippets.                                                                                                                                           | []              | false    |
| `worker_pool.tags`            | Optional details to tag on AWS resources.                                                                                                                                                  | -               | false    |

## Applying

To create the cluster, execute the following command:

```console
lokoctl cluster apply
```

## Uninstalling

To destroy the Lokomotive cluster, execute the following command:

```console
lokoctl cluster destroy --confirm
```
