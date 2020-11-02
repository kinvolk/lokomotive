# Velero

Velero is an open source tool to safely backup and restore, perform disaster recovery, and migrate Kubernetes cluster resources and persistent volumes.

Velero has two main components: a CLI, and a server-side Kubernetes deployment.

## Installing the Velero CLI

See the different options for installing the [Velero CLI](https://velero.io/docs/v1.5/basic-install/#install-the-cli).

## Installing the Velero server

### Velero version

This helm chart installs Velero version v1.5.1 https://github.com/vmware-tanzu/velero/tree/v1.5.1. See the [#Upgrading](#upgrading) section for information on how to upgrade from other versions.

### Provider credentials

When installing using the Helm chart, the provider's credential information will need to be appended into your values. The easiest way to do this is with the `--set-file` argument, available in Helm 2.10 and higher. See your cloud provider's documentation for the contents and creation of the `credentials-velero` file.

### Installing

The default configuration values for this chart are listed in values.yaml.

See Velero's full [official documentation](https://velero.io/docs/v1.5/basic-install/). More specifically, find your provider in the Velero list of [supported providers](https://velero.io/docs/v1.5/supported-providers/) for specific configuration information and examples.


#### Using Helm 3

First, create the namespace: `kubectl create namespace <YOUR NAMESPACE>`

##### Option 1) CLI commands

Note: you may add the flag `--skip-crds` if you don't want to install the CRDs.

Specify the necessary values using the --set key=value[,key=value] argument to helm install. For example,

```bash
helm install vmware-tanzu/velero --namespace <YOUR NAMESPACE> \
--set-file credentials.secretContents.cloud=<FULL PATH TO FILE> \
--set configuration.provider=<PROVIDER NAME> \
--set configuration.backupStorageLocation.name=<BACKUP STORAGE LOCATION NAME> \
--set configuration.backupStorageLocation.bucket=<BUCKET NAME> \
--set configuration.backupStorageLocation.config.region=<REGION> \
--set configuration.volumeSnapshotLocation.name=<VOLUME SNAPSHOT LOCATION NAME> \
--set configuration.volumeSnapshotLocation.config.region=<REGION> \
--set image.repository=velero/velero \
--set image.tag=v1.5.1 \
--set image.pullPolicy=IfNotPresent \
--set initContainers[0].name=velero-plugin-for-aws \
--set initContainers[0].image=velero/velero-plugin-for-aws:v1.1.0 \
--set initContainers[0].volumeMounts[0].mountPath=/target \
--set initContainers[0].volumeMounts[0].name=plugins \
--generate-name
```

##### Option 2) YAML file

Add/update the necessary values by changing the values.yaml from this repository, then run:

```bash
helm install vmware-tanzu/velero --namespace <YOUR NAMESPACE> -f values.yaml --generate-name
```
##### Upgrade the configuration

If a value needs to be added or changed, you may do so with the `upgrade` command. An example:

```bash
helm upgrade vmware-tanzu/velero <RELEASE NAME> --namespace <YOUR NAMESPACE> --reuse-values --set configuration.provider=<NEW PROVIDER>
```

#### Using Helm 2

##### Tiller cluster-admin permissions

A service account and the role binding prerequisite must be added to Tiller when configuring Helm to install Velero:

```
kubectl create sa -n kube-system tiller
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole cluster-admin --serviceaccount kube-system:tiller
helm init --service-account=tiller --wait --upgrade
```

##### Option 1) CLI commands

Note: you may add the flag `--set installCRDs=false` if you don't want to install the CRDs.

Specify the necessary values using the --set key=value[,key=value] argument to helm install. For example,

```bash
helm install vmware-tanzu/velero --namespace <YOUR NAMESPACE> \
--set-file credentials.secretContents.cloud=<FULL PATH TO FILE> \
--set configuration.provider=aws \
--set configuration.backupStorageLocation.name=<BACKUP STORAGE LOCATION NAME> \
--set configuration.backupStorageLocation.bucket=<BUCKET NAME> \
--set configuration.backupStorageLocation.config.region=<REGION> \
--set configuration.volumeSnapshotLocation.name=<VOLUME SNAPSHOT LOCATION NAME> \
--set configuration.volumeSnapshotLocation.config.region=<REGION> \
--set image.repository=velero/velero \
--set image.tag=v1.5.1 \
--set image.pullPolicy=IfNotPresent \
--set initContainers[0].name=velero-plugin-for-aws \
--set initContainers[0].image=velero/velero-plugin-for-aws:v1.1.0 \
--set initContainers[0].volumeMounts[0].mountPath=/target \
--set initContainers[0].volumeMounts[0].name=plugins 
```

##### Option 2) YAML file

Add/update the necessary values by changing the values.yaml from this repository, then run:

```bash
helm install vmware-tanzu/velero --namespace <YOUR NAMESPACE> -f values.yaml
```

##### Upgrade the configuration

If a value needs to be added or changed, you may do so with the `upgrade` command. An example:

```bash
helm upgrade vmware-tanzu/velero <RELEASE NAME> --reuse-values --set configuration.provider=<NEW PROVIDER> 
```

## Upgrading

### Upgrading to v1.5

The [instructions found here](https://velero.io/docs/v1.5/upgrade-to-1.5/) will assist you in upgrading from version v1.4.x to v1.5.


### Upgrading to v1.4

The [instructions found here](https://velero.io/docs/v1.4/upgrade-to-1.4/) will assist you in upgrading from version v1.3.x to v1.4.

### Upgrading to v1.3.1

The [instructions found here](https://velero.io/docs/v1.3.1/upgrade-to-1.3/) will assist you in upgrading from version v1.2.0 or v1.3.0 to v1.3.1.

### Upgrading to v1.2.0

The [instructions found here](https://velero.io/docs/v1.2.0/upgrade-to-1.2/) will assist you in upgrading from version v1.0.0 or v1.1.0 to v1.2.0.

### Upgrading to v1.1.0

The [instructions found here](https://velero.io/docs/v1.1.0/upgrade-to-1.1/) will assist you in upgrading from version v1.0.0 to v1.1.0.

## Uninstall Velero

Note: when you uninstall the Velero server, all backups remain untouched.

### Using Helm 2

```bash
helm delete <RELEASE NAME> --purge
```

### Using Helm 3

```bash
helm delete <RELEASE NAME> -n <YOUR NAMESPACE>
```
