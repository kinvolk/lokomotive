---
title: Deploy configuration using GitOps on an Azure Arc enabled Lokomotive cluster
weight: 10
---

## Introduction

Azure Arc offers simplified management, faster app development, and consistent Azure services.

With Azure Arc, you can:

- Centrally manage a wide range of resources, including
[Windows](https://azure.microsoft.com/en-in/campaigns/windows-server/) and
[Linux](https://azure.microsoft.com/en-in/overview/linux-on-azure/) servers, SQL server,
[Kubernetes](https://azure.microsoft.com/en-in/services/kubernetes-service/) clusters and [Azure
services](https://azure.microsoft.com/en-in/services/azure-arc/hybrid-data-services/).

- Establish central visibility in the [Azure portal](https://azure.microsoft.com/en-in/features/azure-portal/)
and enable multi-environment search with Azure Resource Graph.

- Meet [governance](https://azure.microsoft.com/en-in/solutions/governance/) and compliance standards for
apps, infrastructure and data with [Azure Policy](https://azure.microsoft.com/en-in/services/azure-policy/).

- Delegate access and manage security policies for resources using role-based access control (RBAC) and [Azure
Lighthouse](https://azure.microsoft.com/en-in/services/azure-lighthouse/).

- Organise and inventory assets through a variety of Azure scopes, such as management groups, subscriptions,
resource groups and tags.

This guide provides the steps for onboarding a Lokomotive cluster on Azure Arc.

## Learning objectives

By the end of this guide, the following things would be accomplished:
* Onboard a Lokomotive cluster on Azure Arc.
* Create Azure Arc Kubernetes configuration for the GitOps agent.
* Watch the GitOps agent deploy Kubernetes resources from the git repository.

## Prerequisites

* A Lokomotive cluster.

* Azure command-line interface `az` installed on the local machine.

* `jq` installed on the system.

* Register providers for Azure Arc enabled Kubernetes:
   ```bash
   az provider register --namespace Microsoft.Kubernetes
   az provider register --namespace Microsoft.KubernetesConfiguration
   az provider register --namespace Microsoft.ExtendedLocation
   ```

   Monitor the registration process. Registration may take up to 10 minutes.

   ```bash
   az provider show -n Microsoft.Kubernetes -o table
   az provider show -n Microsoft.KubernetesConfiguration -o table
   az provider show -n Microsoft.ExtendedLocation -o table
   ```

   **NOTE**: This registration is needed only once per tenant.

* Install `k8s-configuration` extension for Azure CLI.

  ```bash
  az extension add --name k8s-configuration
  ```

* Create a resource group:

   ```bash
   RG_NAME="AzureArcTest"
   az group create --name "${RG_NAME}" --location EastUS --output table
   ```

* Create a service principal, its credentials and assign roles to the service principal.

  ```bash
  # Create a ServicePrincipal and its credentials.
  SP_NAME=azure-arc-onboarding-service-principal
  az ad sp create-for-rbac -n "${SP_NAME}" --skip-assignment -o jsonc > /tmp/sp-cred.json
  SP_ID=$(az ad sp list -o tsv --query='[0].objectId' --display-name "${SP_NAME}")

  # Get Subscription ID.
  SUB_ID=$(az account show --query id --output tsv)

  # Assign "Kubernetes Cluster - Azure Arc Onboarding" Role by its identifier.
  az role assignment create --assignee "${SP_ID}" \
    --role "Kubernetes Cluster - Azure Arc Onboarding" \
    --scope /subscriptions/${SUB_ID}/resourcegroups/${RG_NAME}

  # Assign "Microsoft.Kubernetes connected cluster" Role by its identifier.
  az role assignment create --assignee "${SP_ID}" \
    --role "Microsoft.Kubernetes connected cluster role" \
    --scope /subscriptions/${SUB_ID}/resourcegroups/${RG_NAME}
  ```

## Steps

### Step 1: Configure [azure-arc-onboarding](../configuration-reference/components/azure-arc-onboarding.md) Lokomotive component.

#### Config

Generate a file named `azure-arc-onboarding.lokocfg` with the following command:

```bash
# Copy values from /tmp/sp-cred.json created in the prerequisites section.
APPLICATION_ID=$(jq -r .appId /tmp/sp-cred.json)
APPLICATION_PASSWORD=$(jq -r .password /tmp/sp-cred.json)
TENANT_ID=$(jq -r .tenant /tmp/sp-cred.json)
CLUSTER_NAME=mercury

cat <<EOF > azure-arc-onboarding.lokocfg
component "azure-arc-onboarding" {
  application_client_id = "${APPLICATION_ID}"
  application_password  = "${APPLICATION_PASSWORD}"
  tenant_id             = "${TENANT_ID}"
  resource_group        = "${RG_NAME}"
  cluster_name          = "${CLUSTER_NAME}"
}
EOF
```

Ensure that none of the fields is empty.

Check out the component's [configuration
reference](../configuration-reference/components/azure-arc-onboarding.md) for more information.

#### Deploy the component

Execute the following command to deploy the `azure-arc-onboarding` component:

```bash
lokoctl component apply azure-arc-onboarding
```

Verify the pod in the `azure-arc-onboarding` namespace is in the `Completed` state (this may take a few minutes):

```console
$ kubectl -n azure-arc-onboarding get pods -n azure-arc-onboarding
NAME                       READY   STATUS      RESTARTS   AGE
azure-arc-register-4s54j   0/1     Completed   0          81m
```

Azure Arc installs various Helm charts in the `azure-arc` namespace, verify that all the deployments are in
`Running` state:

```console
$ kubectl -n azure-arc get pods
NAME                                         READY   STATUS    RESTARTS   AGE
cluster-metadata-operator-7cff574c4f-ld8cv   2/2     Running   0          83m
clusterconnect-agent-6dfd867c68-629xg        3/3     Running   0          83m
clusteridentityoperator-fd498bf96-dqlhr      2/2     Running   0          83m
config-agent-bd647558d-f46xn                 2/2     Running   0          83m
controller-manager-8676dcdc6-mv68p           2/2     Running   0          83m
extension-manager-bcfd5b597-lsk85            2/2     Running   0          83m
flux-logs-agent-6596f58c56-55l7r             1/1     Running   0          83m
kube-aad-proxy-97bf4bcf8-g9djl               2/2     Running   0          83m
metrics-agent-5b9b94754f-qf5q2               2/2     Running   0          83m
resource-sync-agent-f8c7c6b6b-zvhx4          2/2     Running   0          83m
```

### Step 3: Create a Kubernetes configuration

#### Create a configuration

Once the Lokomotive cluster is onboarded on Azure Arc, we can proceed ahead with creating a configuration for
deploying the GitOps agent.

```bash
az k8s-configuration create \
  --name cluster-config \
  --cluster-name "${CLUSTER_NAME}" \
  --resource-group "${RG_NAME}" \
  --operator-instance-name cluster-config \
  --operator-namespace cluster-config \
  --repository-url https://github.com/ipochi/azure-arc-nginx-demo \
  --scope cluster \
  --cluster-type connectedClusters \
  --operator-params='--git-branch=main'
```

- `name` - the name of the Kubernetes configuration you want.
- `cluster-name` - the name of the cluster provided in `azure-arc-onboarding` component configuration.
- `resource-group` - Azure Resource group name provided in the `azure-arc-onboarding` component configuration.
- `operator-instance-name` - the name of the Operator instance.
- `repository-url` - the name of the git repository.
- `scope` - scope of the operator. Accepted values: `cluster` or `namespace`.
- `cluster-type` - Arc cluster type. Accepted values: `connectedClusters` or `managedClusters`.
   **NOTE**: If the Lokomotive cluster is deployed on AKS, then the value of `cluster-type` is `managedClusters`
- `operator-params` - Additional parameters to pass to the Operator. By default, the GitOps agent tracks and
   syncs with the `master` branch. In this case, we are changing default the branch to `main`.

This command creates two Deployments in the provided namespace `cluster-config`, corresponding to the GitOps
agent that Azure Arc uses:

```console
$ kubectl get deployments -n cluster-config
NAME                         READY   UP-TO-DATE   AVAILABLE   AGE
cluster-config               1/1     1            1           103m
memcached-cluster-config     1/1     1            1           103m

kubectl get pods -n cluster-config
NAME                                          READY   STATUS    RESTARTS   AGE
cluster-config-5f68dd9bd5-xsmc9               1/1     Running   0          5d18h
memcached-cluster-config-857d47969f-v2g8q     1/1     Running   0          5d18h
```

Once the `cluster-config` pods are in `Running` state, you'll notice the new resources created by the GitOps
agent in the namespace `demo`. The manifests for these resources are pulled from the provided git repository.

```console
$ kubectl get all -n demo
NAME                           READY   STATUS    RESTARTS   AGE
pod/my-nginx-5bcdc5784-fb87k   1/1     Running   0          105m
pod/my-nginx-5bcdc5784-g2gt2   1/1     Running   0          105m

NAME               TYPE       CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
service/my-nginx   NodePort   10.3.250.153   <none>        8080:32500/TCP   105m

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/my-nginx   2/2     2            2           105m

NAME                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/my-nginx-5bcdc5784    2         2         2       105m
replicaset.apps/my-nginx-5f4dfbff8c   0         0         0       105m
```

If you used your git repository URL, try pushing some changes to your repository. You'll notice the changes
propagate instantly without the need for any manual intervention.

## Cleanup

Before removing the Lokomotive cluster from Azure arc, it is good practice to remove the GitOps agent.

```bash
az k8s-configuration delete \
  --name cluster-config \
  --cluster-name "${CLUSTER_NAME}" \
  --resource-group "${RG_NAME}" \
  --cluster-type connectedclusters \
```

Next, we remove the Lokomotive cluster from Azure Arc, by deleting the component:

```bash
lokoctl component delete azure-arc-onboarding
```

Finally, if the Lokomotive cluster is not needed, destroy the cluster:

```bash
lokoctl cluster destroy --confirm
```

**NOTE**: Store Service principal credentials somewhere safe and delete the temporary credentials file:

```bash
rm -rf /tmp/sp-cred.json
```

## Additional resources

- `azure-arc-onboarding` component [configuration reference](../configuration-reference/components/azure-arc-onboarding.md) guide.
- Azure Arc docs:

  - [Configurations and GitOps with Azure Arc enabled clusters](https://docs.microsoft.com/en-us/azure/azure-arc/kubernetes/conceptual-configurations).
  - [Using Azure Policy to apply GitOps configurations at scale](https://docs.microsoft.com/en-us/azure/azure-arc/kubernetes/use-azure-policy).
