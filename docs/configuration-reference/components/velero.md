---
title: Velero configuration reference for Lokomotive
weight: 10
---

## Introduction

[Velero](https://github.com/vmware-tanzu/velero) helps you back up and restore your Kubernetes
cluster resources and persistent volumes.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl` deployed.

## Configuration


### Velero on AKS

In order to use Velero on Azure, you need to have Application (Service Principal) created
for it. This service account needs to have access to a storage account with blob storage,
where backups will be stored.

Follow [velero-plugin-for-microsoft-azure#setup](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure#setup) to set it up.

### Example

Velero component configuration example:

```tf
# velero.lokocfg
component "velero" {

  # provider = "azure/openebs/restic/csi"
  # azure {
  #   # Required arguments.
  #   subscription_id = "9e5ac23c-6df8-44c4-9790-6f6decf96268"
  #   tenant_id       = "78bdc534-b34f-4bda-a6ca-6df52915b0b5"
  #   client_id       = "d44117a8-b69d-437b-9073-e4e3b25e164a"
  #   client_secret   = "c26f9698-a563-409e-87ee-4dcf96007b73"
  #   resource_group  = "my-resource-group"
  #
  #   backup_storage_location {
  #     resource_group  = "my-resource-group"
  #     storage_account = "mybackupstorageaccount"
  #     bucket          = "backupscontainer"
  #   }
  #
  #   # Optional parameters
  #   volume_snapshot_location {
  #     resource_group = "my-resource-group"
  #     api_timeout    = "10m"
  #   }
  # }

  # openebs {
  #   credentials = file("cloud-credentails-file")
  #		provider		= "aws"
	#
  #   backup_storage_location {
  #     provider = "aws"
  #     region 	 = "my-region"
  #     bucket 	 = "my-bucket"
  #			name     = "my-backup-location"
  #   }
  #
  #   volume_snapshot_location {
  #			bucket 	 = "my-bucket"
  #			region 	 = "my-region"
  #			provider = "aws"
  #			name 		 = "my-snapshot-location"
  #     prefix   = "backup-prefix"
  #     local    = false
  #
  #     openebs_namespace = "openebs"
  #
  #     s3_url = "mybucket.example.com"
  #   }
  # }

  # restic {
  #   credentials = file("cloud-credentials-file")
  #
  #   require_volume_annotation = true
  #
  #   backup_storage_location {
  #     provider = "aws"
  #     bucket   = "my-bucket"
  #     name     = "my-backup-location"
  #   }
  # }

  # csi {
  #   aws {
  #     credentials = file("./credentials-velero")
  #     backup_storage_location {
  #       bucket              = "my-bucket"
  #       region              = "my-region"
  #       name                = "my-name"
  #       prefix              = "my=prefix"
  #       s3_force_path_style = false
  #       s3_url              = "s3 url"
  #       public_url          = "my-public-s3-url"
  #     }
  #     volume_snapshot_location {
  #       name   = "my-name"
  #       region = "my-region"
  #     }
  #   }
  # }

  # Optional.
  metrics {
    enabled         = false
    service_monitor = false
  }

  namespace = "velero"
}
```

## Attribute reference

Table of all the arguments accepted by the component.

| Argument                                              | Description                                                                                                                                                                          | Default                                           | Type   | Required |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------|--------|----------|
| `namespace`                                           | Namespace to install Velero.                                                                                                                                                         | "velero"                                          | string | false    |
| `provider`                                            | Provider sets which provider block to use for the configuration. Supported values are: `azure`, `openebs`, `restic` and `csi`.                                                       | -                                                 | string | true     |
| `metrics`                                             | Configure Prometheus to scrape Velero metrics. Needs the [Prometheus Operator component](prometheus-operator.md) installed.                                                          | -                                                 | object | false    |
| `metrics.enabled`                                     | Adds Prometheus annotations to Velero deployment if enabled.                                                                                                                         | false                                             | bool   | false    |
| `metrics.service_monitor`                             | Adds ServiceMonitor resource for Prometheus. Requires `metrics.enabled` as true.                                                                                                     | false                                             | bool   | false    |
| `azure`                                               | Configure Azure provider for Velero.                                                                                                                                                 | -                                                 | object | false    |
| `azure.subscription_id`                               | Azure Subscription ID where client application is created. Can be obtained with `az account list`.                                                                                   | -                                                 | string | true     |
| `azure.tenant_id`                                     | Azure Tenant ID where your subscription is created. Can be obtained with `az account list`.                                                                                          | -                                                 | string | true     |
| `azure.client_id`                                     | Azure Application Client ID to perform Azure operations.                                                                                                                             | -                                                 | string | true     |
| `azure.client_secret`                                 | Azure Application Client secret.                                                                                                                                                     | -                                                 | string | true     |
| `azure.resource_group`                                | Azure resource group, where PVC Disks are created. If this argument is wrong, Velero will fail to create PVC snapshots.                                                              | -                                                 | string | true     |
| `azure.backup_storage_location`                       | Configure backup storage location and metadata.                                                                                                                                      | -                                                 | object | true     |
| `azure.backup_storage_location.resource_group`        | Name of the resource group containing the storage account for this backup storage location.                                                                                          | -                                                 | string | true     |
| `azure.backup_storage_location.storage_account`       | Name of the storage account for this backup storage location.                                                                                                                        | -                                                 | string | true     |
| `azure.backup_storage_location.bucket`                | Name of the storage container to store backups.                                                                                                                                      | -                                                 | string | true     |
| `azure.volume_snapshot_location`                      | Configure PVC snapshot location.                                                                                                                                                     | -                                                 | object | false    |
| `azure.volume_snapshot_location.resource_group`       | Azure Resource Group where snapshots will be stored.                                                                                                                                 | Stored in the same resource group as the cluster. | string | false    |
| `azure.volume_snapshot_location.api_timeout`          | Azure API timeout.                                                                                                                                                                   | "10m"                                             | string | false    |
| `openebs`                                             | Configure OpenEBS provider for Velero.                                                                                                                                               | -                                                 | object | false    |
| `openebs.credentials`                                 | Content of cloud provider credentials.                                                                                                                                               | -                                                 | string | true     |
| `openebs.provider`                                    | Cloud provider to use for backup and snapshot storage. Supported values are `gcp` and `aws`.                                                                                         | -                                                 | string | false    |
| `openebs.backup_storage_location`                     | Configure backup storage location.                                                                                                                                                   | -                                                 | object | true     |
| `openebs.backup_storage_location.region`              | Cloud provider region for storing backups.                                                                                                                                           | -                                                 | string | true     |
| `openebs.backup_storage_location.bucket`              | Cloud storage bucket name for storing backups.                                                                                                                                       | -                                                 | string | true     |
| `openebs.backup_storage_location.provider`            | Cloud provider name for storing backups. Overrides `openebs.provider` field for backup storage.                                                                                      | -                                                 | string | false    |
| `openebs.backup_storage_location.name`                | Name for backup location object on the cluster.                                                                                                                                      | -                                                 | string | false    |
| `openebs.volume_snapshot_location`                    | Configure volume snapshot location.                                                                                                                                                  | -                                                 | object | true     |
| `openebs.volume_snapshot_location.bucket`             | Cloud storage bucket name for storing volume snapshots.                                                                                                                              | -                                                 | string | true     |
| `openebs.volume_snapshot_location.region`             | Cloud provider region for storing snapshots.                                                                                                                                         |                                                   | string | true     |
| `openebs.volume_snapshot_location.provider`           | Cloud provider name for storing snapshots. Overrides `openebs.provider` field for backup storage.                                                                                    | -                                                 | string | false    |
| `openebs.volume_snapshot_location.name`               | Name for snapshot location object on the cluster.                                                                                                                                    | -                                                 | string | false    |
| `openebs.volume_snapshot_location.prefix`             | Prefix for snapshot names.                                                                                                                                                           | -                                                 | string | false    |
| `openebs.volume_snapshot_location.local`              | If `true`, backups won't be copied to cloud storage.                                                                                                                                 | false                                             | bool   | false    |
| `openebs.volume_snapshot_location.openebs_namespace`  | Name of the namespace where OpenEBS runs.                                                                                                                                            | -                                                 | string | true     |
| `openebs.volume_snapshot_location.s3_url`             | S3 API URL.                                                                                                                                                                          | -                                                 | string | false    |
| `restic`                                              | Configure Restic provider for Velero.                                                                                                                                                | -                                                 | object | false    |
| `restic.credentials`                                  | Content of cloud provider credentials.                                                                                                                                               | -                                                 | string | true     |
| `restic.require_volume_annotation`                    | Backup all pod volumes without having to apply annotation on the pod when using restic. To exclude volumes add the annotation `backup.velero.io/backup-volumes-excludes` on the pod. | false                                             | bool   | false    |
| `restic.backup_storage_location.provider`             | Cloud provider name for storing backups.                                                                                                                                             | -                                                 | string | false    |
| `restic.backup_storage_location.bucket`               | Cloud storage bucket name for storing backups.                                                                                                                                       | -                                                 | string | true     |
| `restic.backup_storage_location.name`                 | Name for backup location object on the cluster.                                                                                                                                      | "default"                                         | string | false    |
| `restic.backup_storage_location.region`               | Cloud provider region for storing snapshots. Required if `restic.backup_storage_location.provider = aws`.                                                                            | -                                                 | string | false    |
| `csi`                                                 | Configure CSI provider for Velero.                                                                                                                                                   | -                                                 | object | false    |
| `csi.aws`                                             | Configure AWS EBS CSI driver. Needs [`aws_ebs_csi_driver`](../../configuration-reference/components/aws-ebs-csi-driver) component installed.                                         | -                                                 | object | true     |
| `csi.aws.credentials`                                 | Content of AWS credentials.                                                                                                                                                          | -                                                 | string | true     |
| `csi.aws.backup_storage_location`                     | Configure backup storage location.                                                                                                                                                   | -                                                 | object | true     |
| `csi.aws.backup_storage_location.region`              | AWS region for storing snapshots.                                                                                                                                                    | -                                                 | string | true     |
| `csi.aws.backup_storage_location.bucket`              | AWS S3 bucket name for Velero backups.                                                                                                                                               | -                                                 | string | true     |
| `csi.aws.backup_storage_location.name`                | Name for backup location object on the cluster.                                                                                                                                      | -                                                 | string | false    |
| `csi.aws.backup_storage_location.prefix`              | Prefix for backup object names.                                                                                                                                                      | -                                                 | string | false    |
| `csi.aws.backup_storage_location.s3_force_path_style` | Use path-style addressing instead of virtual hosted bucket addressing. Set to `true` if using MinIO.                                                                                 | false                                             | bool   | false    |
| `csi.aws.backup_storage_location.s3_url`              | S3 API URL. Set this field if using MinIO.                                                                                                                                           | -                                                 | string | false    |
| `csi.aws.backup_storage_location.public_url`          | S3 API URL.                                                                                                                                                                          | -                                                 | string | false    |
| `csi.aws.volume_snapshot_location.name`               | Name for volume snapshot location object on the cluster.                                                                                                                             | -                                                 | string | false    |
| `csi.aws.volume_snapshot_location.region`             | AWS S3 bucket name for Velero snapshots.                                                                                                                                             | -                                                 | string | true     |



## Applying

To apply the Velero component:

```bash
lokoctl component apply velero
```

### Post-installation

For day-to-day tasks, the `velero` CLI tool is the recommended way to interact with Velero.

You can find how to install it in the [official documentation](https://velero.io/docs/v1.4/basic-install#install-the-cli).

To learn more on taking backups with Velero, visit [Velero#getting-stated](https://velero.io/docs/v1.4/examples/)

## Deleting

To destroy the component:

```bash
lokoctl component delete velero --delete-namespace
```
