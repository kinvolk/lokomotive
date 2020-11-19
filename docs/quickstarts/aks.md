---
title: Lokomotive AKS quickstart guide
weight: 10
---

## Introduction

This quickstart guide walks through the steps needed to create a Lokomotive cluster on AKS.

By the end of this guide, you'll have a production-ready Kubernetes cluster running on Azure AKS.

_Note: Lokomotive on AKS currently provides Kubernetes 1.16 as opposed to other platforms, which provide 1.18. This is because of limitations of Azure platform._

## Requirements

* Basic understanding of Kubernetes concepts.
* [Azure account](https://azure.microsoft.com/en-us/free/).
* `kubectl` installed locally to access the Kubernetes cluster.

## Steps

### Step 1: Install lokoctl

lokoctl is a command-line interface for Lokomotive.

To install `lokoctl`, follow the instructions in the [lokoctl installation](../installer/lokoctl.md)
guide.

### Step 2: Set up a working directory

It's better to start fresh in a new working directory, as the state of the cluster is stored in this
directory.

This also makes the cleanup task easier.

```console
mkdir -p lokomotive-infra/myakscluster
cd lokomotive-infra/myakscluster
```

### Step 3: Set up Azure API credentials

To create an AKS resource in Azure, you need to be authenticatied to Azure API. Follow
[Authenticating to Azure](https://www.terraform.io/docs/providers/azurerm/index.html#authenticating-to-azure)
to set up environment variables required for authentication.

### Step 4: Prepare AKS credentials

An AKS cluster requires a set of service principal credentials to run, as it talks to Azure API to create Load Balancers,
Disks and other objects. Depending on your level of privileges in Azure, there are different ways to provide them.

#### Azure AD Application Creator (full automation)

If you are an Azure AD administrator or if your Azure user has permissions to create Azure AD Applications then
you don't need to prepare anything manually. When configuring a cluster, set the `application_name` property
to e.g. the cluster name and `lokoctl` will create Azure AD application for you, together with the associated
service principal and credentials. Those credentials will be automatically used for running AKS.

#### Subscription collaborator

If you are a user with full administrative access to your subscription, then you need to ask your administrator to create
Azure AD application for you and provide you a Service Principal Client ID and a Client secret, which will be used by AKS cluster.

You can then provide them to the configuration using either `LOKOMOTIVE_AKS_CLIENT_ID` and `LOKOMOTIVE_AKS_CLIENT_SECRET` environment
variables or via `client_id` and `client_secret` parameters. See [AKS attribute reference](../configuration-reference/platforms/aks.md#attribute-reference) for more details.

#### Resource group collaborator

If your Azure user has only access to a single Resource Group, you must set the `manage_resource_group` property to `false`,
as otherwise `lokoctl` will try to create a Resource Group for you.

You also need Service Principal credentials, as explained in [#subscription-collabolator](#subscription-collabolator).

### Step 5: Define cluster configuration

To create a Lokomotive cluster, you need to define a configuration.

A [production-ready configuration](../../examples/aks-production) is already provided for ease of
use. Copy the example configuration to the working directory and modify accordingly.

The provided configuration installs the Lokomotive cluster and the following components:

* [prometheus-operator](../configuration-reference/components/prometheus-operator.md)
* [cert-manager](../configuration-reference/components/cert-manager.md)
* [contour](../configuration-reference/components/contour.md)

You can configure the components as per your requirements.

Lokomotive can store Terraform state [locally](../configuration-reference/backend/local.md)
or remotely within an [AWS S3 bucket](../configuration-reference/backend/s3.md). By default, Lokomotive
stores Terraform state locally.

Create a variables file named `lokocfg.vars` in the working directory to set values for variables
defined in the configuration file.

```console
#lokocfg.vars
state_s3_bucket = "name-of-the-s3-bucket-to-store-the-cluster-state"
lock_dynamodb_table = "name-of-the-dynamodb-table-for-state-locking"

cert_manager_email = "email-address-used-for-cert-manager-component"
grafana_admin_password = "password-for-grafana"
```

**NOTE**: You can separate component configurations from cluster configuration in separate
configuration files if doing so fits your needs.

Example:
```console
$ ls lokomotive-infra/myakscluster
cluster.lokocfg  prometheus-operator.lokocfg  lokocfg.vars
```

For advanced cluster configurations and more information refer to the [AKS configuration
guide](../configuration-reference/platforms/aks.md).

### Step 6: Create Lokomotive cluster

Run the following command to create the cluster:

```console
lokoctl cluster apply
```
Once the command finishes, your Lokomotive cluster details are stored in the path you've specified
under `asset_dir`.

## Verification

A successful installation results in the output:

```console
azurerm_kubernetes_cluster.aks: Still creating... [8m0s elapsed]
azurerm_kubernetes_cluster.aks: Still creating... [8m10s elapsed]
azurerm_kubernetes_cluster.aks: Still creating... [8m20s elapsed]
azurerm_kubernetes_cluster.aks: Creation complete after 8m24s [id=/subscriptions/55555555-4444-3333-2222-1111111111/resourcegroups/ci1586244933-fg/providers/Microsoft.ContainerService/managedClusters/ci1586244933-fg]
local_file.kubeconfig: Creating...
local_file.kubeconfig: Creation complete after 0s [id=f96468e341a652192af7508836430241e6f49df1]

Apply complete! Resources: 3 added, 0 changed, 0 destroyed.

Outputs:

initialized = true

Your configurations are stored in /root/lokoctl-assets

Now checking health and readiness of the cluster nodes ...

Node                               Ready    Reason          Message

aks-default-31666422-vmss000000    True     KubeletReady    kubelet is posting ready status. AppArmor enabled
aks-default-31666422-vmss000001    True     KubeletReady    kubelet is posting ready status. AppArmor enabled

Success - cluster is healthy and nodes are ready!
```

Use the generated `kubeconfig` file to access the Kubernetes cluster and list nodes.

```console
export KUBECONFIG=./lokomotive-assets/cluster-assets/auth/kubeconfig
kubectl get nodes
```

## Using the cluster

At this point you have access to the Kubernetes cluster and can use it!
If you don't have Kubernetes experience you can check out the [Kubernetes
Basics official
documentation](https://kubernetes.io/docs/tutorials/kubernetes-basics/deploy-app/deploy-intro/)
to learn about its usage.

## Cleanup

To destroy the Lokomotive cluster, execute the following command:

```console
lokoctl cluster destroy --confirm
```

You can safely delete the working directory created for this quickstart guide if you no longer
require it.

## Conclusion

After walking through this guide, you've learned how to set up a Lokomotive cluster on AKS.

## Next steps

You can now start deploying your workloads on the cluster.

For more information on installing supported Lokomotive components, you can visit the [component
configuration references](../configuration-reference/components).
