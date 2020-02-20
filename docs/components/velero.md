[Velero](https://github.com/heptio/velero) is a tool, which allows you to back up
and restore your Kubernetes cluster resources and persistent volumes.

## Lokomotive component

Velero is available as a component in lokoctl.

## Compatibility

Currently lokoctl supports Velero only on AKS platform. Support for another platforms
will be added in the future.

## Requirements

Requirements for using Velero depends on the platform you want to run it.
Please see platform-specific requirements below.

### General

For day-to-day tasks, `velero` CLI tool is a recommended way to interact with Velero.

You can find how to install it in [official documentation](https://velero.io/docs/master/install-overview/).

### Azure

In order to use Velero on Azure, you need to have Application (service principle) created
for it. This service account needs to have access to storage account with blob storage,
where backups will be stored.

Please follow [velero-plugin-for-microsoft-azure#setup](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure#setup) to set it up.

## Configuration

The velero lokoctl component currently supports the following options:

```
# velero.lokocfg

component "velero" {
  # Required parameters when using provider 'azure'
  #
  # Azure Subscription ID where client application is created.
  # Can be obtained with `az account list`.
  azure_subscription_id = "9e5ac23c-6df8-44c4-9790-6f6decf96268"

  # Azure Tenant ID where your subscription is created.
  # Can be obtained with `az account list`.
  azure_tenant_id       = "78bdc534-b34f-4bda-a6ca-6df52915b0b5"

  # Azure Application Client ID, which will be used to perform Azure operations.
  azure_client_id       = "d44117a8-b69d-437b-9073-e4e3b25e164a"

  # Azure Application Client Secret.
  azure_client_secret   = "c26f9698-a563-409e-87ee-4dcf96007b73"

  # Azure resource group, where PVC Disks are created. Note that AKS creates it's own
  # resource group where all nodes and disks are stored, unless you provide it 'nodeResourceGroup' parameter.
  # If this parameter is wrong, Velero will fail to create PVC snapshots.
  azure_resource_group  = "my-resource-group"

  # Configuration where backups and it's metadata are stored.
  azure_backup_storage_location {

    # Name of the resource group containing the storage account for this backup storage location.
    resource_group  = "my-resource-group"

    # Name of the storage account for this backup storage location.
    storage_account = "mybackupstorageaccount"

    # Name of the storage container to store backups.
    bucket          = "backupscontainer"
  }

  # Optional parameters
  #
  # Configuration where PVC snapshots are stored.
  azure_volume_snapshot_location {

    # Azure Resource Group where snapshots will be stored. By default they will be stored in the same resource
    # group as the cluster, meaning if you destroy the cluster, snapshots will be destroyed as well.
    resource_group = "my-resource-group"

    # Azure API timeout. Defaults to 10m.
    api_timeout    = "10m"
  }

  # Metrics controls if prometheus annotations should be added to velero deployment
  metrics {

    # If this is true, prometheus service annotations will be added to velero pod.
    enabled = false

    # Enables creation of ServieMonitor CR used by prometheus operator. Requires 'enabled' to be 'true'.
    service_monitor = false
  }

  # Currently the only supported provider is 'azure', so this parameter is optional.
  provider = "azure"

  # Namespace where velero workload should be deployed. Defaults to 'velero'.
  namespace = "velero"
}
```

### Installation

After preparing your configuration in a lokocfg file (e.g. `velero.lokocfg`), you
can install the component with

```
lokoctl component install velero
```

velero pod runs in the velero namespace, see `kubectl get pods -n velero`.

You can verify that velero is up and running by executing the following command:

```sh
$ velero version
Client:
	Version: v1.0.0
	Git commit: 72f5cadc3a865019ab9dc043d4952c9bfd5f2ecb
Server:
	Version: v1.0.0
```

## Next steps

Once you have successfully installed and configured velero, you can use it
to backup your workloads.

To see all available backups:
```sh
$ velero backup get
NAME   STATUS      CREATED                          EXPIRES   STORAGE LOCATION   SELECTOR
test   Completed   2019-07-23 17:15:17 +0200 CEST   30d       default            <none>
```

To create full cluster backup:
```sh
$ velero backup create test
Backup request "test" submitted successfully.
Run `velero backup describe test` or `velero backup logs test` for more details.
```

See [Velero#getting-stated](https://heptio.github.io/velero/master/get-started.html) for more examples.
