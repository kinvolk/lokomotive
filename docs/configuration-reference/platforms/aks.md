---
title: Lokomotive AKS configuration reference
weight: 10
---

## Introduction

This configuration reference provides information on configuring a Lokomotive cluster on Azure AKS with all the configuration options available to the user.

## Prerequisites

* `lokoctl` [installed locally](../../installer/lokoctl.md).
* `kubectl` installed locally to access the Kubernetes cluster.

## Configuration

To create a Lokomotive cluster, we need to define a configuration.

Example configuration file:

```tf
#myakscluster.lokocfg
variable "state_s3_bucket" {}
variable "lock_dynamodb_table" {}
variable "asset_dir" {}
variable "cluster_name" {}
variable "workers_count" {}
variable "state_s3_key" {}
variable "state_s3_region" {}
variable "workers_vm_size" {}
variable "location" {}
variable "tenant_id" {}
variable "subscription_id" {}
variable "client_id" {}
variable "client_secret" {}
variable "resource_group_name" {}
variable "application_name" {}
variable "manage_resource_group" {}

backend "s3" {
  bucket         = var.state_s3_bucket
  key            = var.state_s3_key
  region         = var.state_s3_region
  dynamodb_table = var.lock_dynamodb_table
}

# backend "local" {
#   path = "path/to/local/file"
#}


cluster "aks" {
  asset_dir    = pathexpand(var.asset_dir)
  cluster_name = var.cluster_name

  tenant_id       = var.tenant_id
  subscription_id = var.subscription_id
  client_id       = var.client_id
  client_secret   = var.client_secret

  location              = var.location
  resource_group_name   = var.resource_group_name
  application_name      = var.application_name
  manage_resource_group = var.manage_resource_group

  worker_pool "default" {
    count   = var.workers_count
    vm_size = var.workers_vm_size

    labels = {
      "key" = "value",
    }

    taints = [
      "node-role.kubernetes.io/master=NoSchedule",
    ]
  }

  tags = {
    "key" = "value",
  }
}
```

**NOTE**: Should you feel differently about the default values, you can set default values using the `variable`
block in the cluster configuration.

## Attribute reference

| Argument                | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |    Default    |     Type     | Required |
|-------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------------:|:------------:|:--------:|
| `asset_dir`             | Location where Lokomotive stores cluster assets.                                                                                                                                                                                                                                                                                                                                                                                                                                                        |       -       |    string    |   true   |
| `cluster_name`          | Name of the cluster. **NOTE**: It must be unique per resource group.                                                                                                                                                                                                                                                                                                                                                                                                                                    |       -       |    string    |   true   |
| `tenant_id`             | Azure Tenant ID. Can also be provided using the `LOKOMOTIVE_AKS_TENANT_ID` environment variable.                                                                                                                                                                                                                                                                                                                                                                                                        |       -       |    string    |   true   |
| `subscription_id`       | Azure Subscription ID. Can also be provided using the `LOKOMOTIVE_AKS_SUBSCRIPTION_ID` environment variable.                                                                                                                                                                                                                                                                                                                                                                                            |       -       |    string    |   true   |
| `resource_group_name`   | Name of the resource group, where AKS cluster object will be created. Please note, that AKS will also create a separate resource group for workers and other required objects, like load balancers, disks etc. If `manage_resource_group` parameter is set to `false`, this resource group must be manually created before cluster creation.                                                                                                                                                            |       -       |    string    |   true   |
| `client_id`             | Azure service principal ID used  for running the AKS cluster. Can also be provided using the `LOKOMOTIVE_AKS_CLIENT_ID`. This parameter is mutually exclusive with `application_name` parameter.                                                                                                                                                                                                                                                                                                        |       -       |    string    |  false   |
| `client_secret`         | Azure service principal secret used  for running the AKS cluster. Can also be provided using the `LOKOMOTIVE_AKS_CLIENT_SECRET`. This parameter is mutually exclusive with `application_name` parameter.                                                                                                                                                                                                                                                                                                |       -       |    string    |  false   |
| `tags`                  | Additional tags for Azure resources.                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |       -       | map(string)  |  false   |
| `location`              | Azure location where resources will be created. Valid values can be obtained using the following command from Azure CLI: `az account list-locations -o table`.                                                                                                                                                                                                                                                                                                                                          | "West Europe" |    string    |  false   |
| `application_name`      | Azure AD application name. If specified, a new Application will be created in Azure AD together with a service principal, which will be used to run the AKS cluster on behalf of the user to provide full cluster creation automation. Please note that this requires [permissions to create applications in Azure AD](https://docs.microsoft.com/en-us/azure/active-directory/users-groups-roles/roles-delegate-app-roles). This parameter is mutually exclusive with `client_id` and `client_secret`. |       -       |    string    |  false   |
| `manage_resource_group` | If `true`, a resource group for the AKS object will be created on behalf of the user.                                                                                                                                                                                                                                                                                                                                                                                                                   |     true      |     bool     |  false   |
| `worker_pool`           | Configuration block for worker pools. At least one worker pool must be defined.                                                                                                                                                                                                                                                                                                                                                                                                                         |       -       | list(object) |   true   |
| `worker_pool.count`     | Number of workers in the worker pool. Can be changed afterwards to add or delete workers.                                                                                                                                                                                                                                                                                                                                                                                                               |       -       |    number    |   true   |
| `worker_pool.vm_size`   | Azure VM size for worker nodes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |       -       |    string    |   true   |
| `worker_pool.labels`    | Map of Kubernetes Node object labels.                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |       -       | map(string)  |  false   |
| `worker_pool.taints`    | List of Kubernetes Node taints.                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |       -       | list(string) |  false   |


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
