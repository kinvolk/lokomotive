# Velero configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Argument reference](#argument-reference)
* [Installation](#installation)
* [Uninstallation](#uninstallation)

## Introduction

[Velero](https://github.com/vmware-tanzu/velero) helps you back up and restore your Kubernetes
cluster resources and persistent volumes.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl` deployed on Azure AKS.

* Permissions to create Service Principals with Azure AKS.

## Configuration

The Velero component is currently supported only on Azure AKS. Support for another platforms will be added in the future.

In order to use Velero on Azure, you need to have Application (Service Principal) created
for it. This service account needs to have access to a storage account with blob storage,
where backups will be stored.

Follow [velero-plugin-for-microsoft-azure#setup](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure#setup) to set it up.

Velero component configuration example:

```tf
# velero.lokocfg
component "velero" {

  # Required.
  azure {
    # Required arguments.
    subscription_id = "9e5ac23c-6df8-44c4-9790-6f6decf96268"
    tenant_id       = "78bdc534-b34f-4bda-a6ca-6df52915b0b5"
    client_id       = "d44117a8-b69d-437b-9073-e4e3b25e164a"
    client_secret   = "c26f9698-a563-409e-87ee-4dcf96007b73"
    resource_group  = "my-resource-group"

    backup_storage_location {
      resource_group  = "my-resource-group"
      storage_account = "mybackupstorageaccount"
      bucket          = "backupscontainer"
    }

    # Optional parameters
    volume_snapshot_location {
      resource_group = "my-resource-group"
      api_timeout    = "10m"
    }
  }

  # Optional.
  metrics {
    enabled = false
    service_monitor = false
  }

  provider = "azure"
  namespace = "velero"
}
```

## Argument reference

Table of all the arguments accepted by the component.

Example:

| Argument                                        | Description                                                                                                                | Default                                           | Required |
|-------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------------:|:--------:|
| `provider`                                      | Supported provider name. Only `azure` is supported for now.                                                                | "azure"                                           | false    |
| `namespace`                                     | Namespace to install Velero.                                                                                               | "velero"                                          | false    |
| `metrics`                                       | Configure Prometheus to scrape Velero metrics. Needs the [Prometheus Operator component](prometheus-operator.md) installed.| -                                                 | false    |
| `metrics.enabled`                               | Adds Prometheus annotations to Velero deployment if enabled.                                                               | false                                             | false    |
| `metrics.service_monitor`                       | Adds ServiceMonitor resource for Prometheus. Requires `metrics.enabled` as true.                                           | false                                             | false    |
| `azure`                                         | Configure Azure provider for Velero.                                                                                       | -                                                 | true     |
| `azure.subscription_id`                         | Azure Subscription ID where client application is created. Can be obtained with `az account list`.                         | -                                                 | true     |
| `azure.tenant_id`                               | Azure Tenant ID where your subscription is created. Can be obtained with `az account list`.                                | -                                                 | true     |
| `azure.client_id`                               | Azure Application Client ID to perform Azure operations.                                                                   | -                                                 | true     |
| `azure.client_secret`                           | Azure Application Client secret.                                                                                           | -                                                 | true     |
| `azure.resource_group`                          | Azure resource group, where PVC Disks are created. If this argument is wrong, Velero will fail to create PVC snapshots.    | -                                                 | true     |
| `azure.backup_storage_location`                 | Configure backup storage location and metadata.                                                                            | -                                                 | true     |
| `azure.backup_storage_location.resource_group`  | Name of the resource group containing the storage account for this backup storage location.                                | -                                                 | true     |
| `azure.backup_storage_location.storage_account` | Name of the storage account for this backup storage location.                                                              | -                                                 | true     |
| `azure.backup_storage_location.bucket`          | Name of the storage container to store backups.                                                                            | -                                                 | true     |
| `azure.volume_snapshot_location`                | Configure PVC snapshot location.                                                                                           | -                                                 | false    |
| `azure.volume_snapshot_location.resource_group` | Azure Resource Group where snapshots will be stored.                                                                       | Stored in the same resource group as the cluster. | false    |
| `azure.volume_snapshot_location.api_timeout`    | Azure API timeout.                                                                                                         | "10m"                                             | false    |

## Installation

To install the Velero component:

```bash
lokoctl component install velero
```

### Post-insallation

For day-to-day tasks, the `velero` CLI tool is the recommended way to interact with Velero.

You can find how to install it in the [official documentation](https://velero.io/docs/master/basic-install/#install-the-cli).

To learn more on taking backups with Velero, visit [Velero#getting-stated](https://velero.io/docs/v1.2.0/examples/)

## Uninstallation

To uninstall the component:

```bash
lokoctl component render-manifest velero | kubectl delete -f -
```
