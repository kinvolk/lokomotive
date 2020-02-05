# Packet

In this tutorial, we'll create a Kubernetes cluster on [Packet](https://packet.net) with [Flatcar Container Linux](https://www.flatcar-linux.org/).

We'll declare a Kubernetes cluster using the Lokomotive Terraform module. Then apply the changes to create DNS records, controller machines, worker machines and TLS assets.

Controllers are provisioned to run an `etcd-member` peer and a `kubelet` service. Workers run just a `kubelet` service. A one-time [bootkube](https://github.com/kubernetes-incubator/bootkube) bootstrap schedules the `apiserver`, `scheduler`, `controller-manager`, and `coredns` on controllers and schedules `kube-proxy` and `calico` (or `flannel`) on every node. A generated `kubeconfig` provides `kubectl` access to the cluster.

## Requirements

* Packet account, Project ID and [API key](https://support.packet.com/kb/articles/api-integrations) (Note, that the term "Auth Token" is also used to refer to the API key in the packet docs)
* [DNS Zone](#dns-zone)
* Terraform v0.12.x and [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct) installed locally

## Terraform Setup

Install [Terraform](https://www.terraform.io/downloads.html) v0.12.x on your system.

```sh
$ terraform version
Terraform v0.12.17
```

Add the [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct) plugin binary for your system to `~/.terraform.d/plugins/`, noting the final name.

```sh
wget https://github.com/poseidon/terraform-provider-ct/releases/download/v0.4.0/terraform-provider-ct-v0.4.0-linux-amd64.tar.gz
tar xzf terraform-provider-ct-v0.4.0-linux-amd64.tar.gz
mv terraform-provider-ct-v0.4.0-linux-amd64/terraform-provider-ct ~/.terraform.d/plugins/terraform-provider-ct_v0.4.0
```

Read [concepts](/docs/architecture/concepts.md) to learn about Terraform, modules, and organizing resources. Change to your infrastructure repository (e.g. `infra`).

```
cd infra/clusters
```

## Provider

```
[default]
aws_access_key_id = xxx
aws_secret_access_key = yyy
```

!!! tip
    Other standard AWS authentication methods can be used instead of specifying `shared_credentials_file` under the provider's config. For more information see the [docs](https://www.terraform.io/docs/providers/aws/#authentication).

Configure the AWS provider to use your access key credentials in a `providers.tf` file.

```
provider "aws" {
  version = "2.31.0"

  region                  = "eu-central-1"
  shared_credentials_file = "/home/user/.config/aws/credentials"
}
```



!!! tip
    The Packet facilities (i.e. data centers) list can be dynamically queried using the [API docs](https://www.packet.com/developers/api/#facilities).

### Packet

Login to your Packet account and obtain the project ID from the `Project Settings` tab. Obtain an API Key from the User settings menu. Note that project level API keys don't have all the necessary permissions for this exercise. The API key can be set in the `providers.tf` file for the `packet` provider as described in the docs [here](https://www.terraform.io/docs/providers/packet/index.html#example-usage). However this is not recommended to avoid accidentally committing API keys to version control. Instead set the env variable `PACKET_AUTH_TOKEN`.

Additional configuration options are describe in the `packet` provider [docs](https://www.terraform.io/docs/providers/packet/).

## Cluster

Define a Kubernetes cluster using the controller module [packet/flatcar-linux/kubernetes](https://github.com/kinvolk/lokomotive-kubernetes/tree/master/packet/flatcar-linux/kubernetes) and the worker module [packet/flatcar-linux/kubernetes/workers](https://github.com/kinvolk/lokomotive-kubernetes/tree/master/packet/flatcar-linux/kubernetes/workers).

```tf
locals {
  project_id   = "93fake81..."
  cluster_name = "supernova"
  facility     = "ams1"
}

module "controller" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//packet/flatcar-linux/kubernetes?ref=<hash>"

  # DNS configuration
  dns_zone    = "myclusters.example.com"

  # configuration
  ssh_keys = [
    "ssh-rsa AAAAB3Nz...",
    "ssh-rsa AAAAB3Nz...",
  ]

  asset_dir = "/home/user/.secrets/clusters/packet"

  # Packet
  cluster_name = local.cluster_name
  project_id   = local.project_id
  facility     = local.facility

  # optional
  controller_count = 1
  controller_type  = "t1.small.x86"

  management_cidrs = [
    "0.0.0.0/0",       # Instances can be SSH-ed into from anywhere on the internet.
  ]

  # This is different for each project on Packet and depends on the packet facility/region.
  # Check yours from the `IPs & Networks` tab from your Packet.net account.
  # If an IP block is not allocated yet, try provisioning an instance from the console in
  # that region. Packet will allocate a public IP CIDR.
  # Note: Packet does not guarantee this CIDR to be stable if there are no servers deployed in the project and region
  node_private_cidr = "10.128.156.0/25"
}

# DNS module that creates the required DNS entries for the cluster.
# More details in DNS Zone section of this document.
module "dns" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//dns/route53?ref=<hash>"

  entries = module.controller.dns_entries
  aws_zone_id = "Z1_FAKE" # Z23ABC4XYZL05B for instance
}

module "worker-pool-helium" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//packet/flatcar-linux/kubernetes/workers?ref=<hash>"

  ssh_keys = [
    "ssh-rsa AAAAB3Nz...",
    "ssh-rsa AAAAB3Nz...",
  ]

  cluster_name = local.cluster_name
  project_id   = local.project_id
  facility     = local.facility
  pool_name    = "helium"

  worker_count = 2
  type  = "t1.small.x86"

  kubeconfig = module.controller.kubeconfig

  labels = "node.supernova.io/role=backend"
}
```

Reference the [variables docs](#variables) or the [variables.tf source of the controller module](https://github.com/kinvolk/lokomotive-kubernetes/blob/master/packet/flatcar-linux/kubernetes/variables.tf) and the [variables.tf source of the worker module](https://github.com/kinvolk/lokomotive-kubernetes/blob/master/packet/flatcar-linux/kubernetes/workers/variables.tf)

## ssh-agent

Initial bootstrapping requires `bootkube.service` be started on one controller node. Terraform uses `ssh-agent` to automate this step. Add your SSH private key to `ssh-agent`.

```sh
ssh-add ~/.ssh/id_rsa
ssh-add -L
```

## Apply

Initialize the config directory if this is the first use with Terraform.

```sh
terraform init
```

Plan the resources to be created.

```sh
$ terraform plan
Plan: 98 to add, 0 to change, 0 to destroy.
```

Apply the changes to create the cluster.

```sh
$ terraform apply
...
module.controller.null_resource.bootkube-start: Still creating... (4m50s elapsed)
module.controller.null_resource.bootkube-start: Still creating... (5m0s elapsed)
module.controller.null_resource.bootkube-start: Creation complete after 11m8s (ID: 3961816482286168143)

Apply complete! Resources: 98 added, 0 changed, 0 destroyed.
```

In 5-10 minutes, the Kubernetes cluster will be ready.

## Verify

[Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) on your system. Use the generated `kubeconfig` credentials to access the Kubernetes cluster and list nodes.

```
$ export KUBECONFIG=/home/user/.secrets/clusters/packet/auth/kubeconfig
$ kubectl get nodes
NAME                       STATUS  ROLES              AGE  VERSION
supernova-controller-0     Ready   controller,master  10m  v1.14.1
supernova-helium-worker-0  Ready   node               10m  v1.14.1
supernova-helium-worker-1  Ready   node               10m  v1.14.1
```

## Going Further

Learn about [maintenance](../topics/maintenance.md).

## ARM Support and Hybrid Clusters

Lokomotive and Flatcar currently include an alpha-quality preview for the Packet arm64 server types `c1.large.arm` and `c2.large.arm`.
They can be used for both worker and controller nodes.
Besides specifying them in `controller_type`/`type` you need to configure some additional variables
in the respective controller/worker module:

```
os_arch = "arm64"
os_channel = "alpha"
ipxe_script_url = "https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/arm64-usr/packet.ipxe"
```

The `os_channel` variable is needed as long as the content of the iPXE script refers to `alpha.release…` as `base-url`.
Adjust `os_channel` to the named channel there or remove it once the script uses `stable.release…` since `stable` is the default.
The iPXE boot variable can be removed once Flatcar is available for installation in the Packet OS menu for the ARM servers.

If you have a hybrid cluster with both x86 and ARM nodes, you need to either use Docker multiarch images such as the standard
`debian:latest` or `python:latest` images, or restrict the Pods to the nodes of the correct architecture with an entry like
this for ARM (or with `amd64` for x86) in your YAML deployment:

```
nodeSelector:
  kubernetes.io/arch: arm64
```

An example on how to build multiarch images yourself is [here](https://github.com/kinvolk/calico-hostendpoint-controller/#building).


## Variables

Check the [variables.tf](https://github.com/kinvolk/lokomotive-kubernetes/blob/master/packet/flatcar-linux/kubernetes/variables.tf) source.

### Required

#### Controller module

| Name | Description | Example |
|:-----|:------------|:--------|
| cluster_name | Unique cluster name (prepended to dns_zone) | "tempest" |
| dns_zone | DNS zone | "myclusters.example.com" |
| ssh_keys | List of SSH public keys for user 'core' | ["ssh-rsa AAAAB3NZ..."] |
| asset_dir | Path to a directory where generated assets should be placed (contains secrets) | "/home/user/.secrets/clusters/tempest" |
| project_id | Project ID obtained from the Packet account | "93fake81-0f3c1-..." |
| facility | Packet Region in which the instance(s) should be deployed | https://www.packet.com/developers/api/#facilities. Eg: "ams1" |
| management_cidrs | List of CIDRs to allow SSH access to the nodes | ["153.79.80.1/16", "59.60.10.1/32"] |
| node_private_cidr | Private CIDR obtained from Packet for the project and facility | 10.128.16.32/25 |

#### Worker module

| Name | Description | Example |
| :----|:-----------:|:--------|
| ssh_keys | List of SSH public keys for user 'core' | ["ssh-rsa AAAAB3NZ..."] |
| cluster_name | Unique cluster name. Must be same as the value used in the controller module | "tempest" |
| project_id | Project ID obtained from the Packet account. Must be same as the value used in the controller module | "93fake81-0f3c1-..." |
| facility | Packet Region in which the instance(s) should be deployed | https://www.packet.com/developers/api/#facilities. Eg: "ams1" |
| pool_name | Name of the worker pool. Used in setting hostname | "helium" |
| kubeconfig | Kubeconfig to be used in worker pools | "${module.controller.kubeconfig} |

#### DNS Zone

Clusters create few DNS A records to resolve to controller instances. For example `${cluster_name}.${dns_zone}` is used by workers and `kubectl` to access the apiserver(s). In this example, the cluster's apiserver would be accessible at `tempest.myclusters.example.com`.

In order to create such DNS entries you'll need a registered domain/subdomain and to define a DNS module that takes as input the DNS entries required by the controller module.

```tf
module "dns" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//dns/<provider>?ref=<hash>"

  # DNS entries required for the cluster to work.
  entries = module.controller.dns_entries

  # Specific configuration for this DNS provider.
}
```

Lokomotive implements support for some [DNS providers](../dns/), if your provider is not supported you'll need to implement the module yourself or set the DNS entries by hand:

```bash
  # Create controller nodes (you could also create worker nodes to save time)
  terraform apply -target=module.controller.null_resource.dns_entries

  # Get list of DNS entries to be created
  terraform state show module.controller.null_resource.dns_entries

  # Create the DNS entries by hand

  # Finish deploying the cluster
  terraform apply
```

### Optional

#### Controller module

| Name | Description | Default | Example |
|:-----|:------------|:--------|:--------|
| controller_count | Number of controllers (i.e. masters) | 1 | 1 |
| controller_type | Type of nodes to provision | "baremetal_0" | "t1.small.x86". See https://www.packet.com/cloud/servers/ for more |
| os_channel | Flatcar Container Linux channel to install from | "stable" | "stable", "beta", "alpha", "edge" |
| os_arch    | Flatcar Container Linux architecture to install | "amd64"  | "amd64", "arm64" |
| os_version | Version of a Flatcar Container Linux release, only for iPXE | "current" | "2191.5.0" |
| ipxe_script_url | URL that contains iPXE script to boot Flatcar on the node over PXE | "" | https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/amd64-usr/packet.ipxe, https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/arm64-usr/packet.ipxe |
| networking | Choice of networking provider | "calico" | "calico" or "flannel" |
| network_mtu | CNI interface MTU (calico only) | 1480 | 8981 |
| pod_cidr | CIDR IPv4 range to assign to Kubernetes pods | "10.2.0.0/16" | "10.22.0.0/16" |
| service_cidr | CIDR IPv4 range to assign to Kubernetes services | "10.3.0.0/16" | "10.3.0.0/24" |
| cluster_domain_suffix | FQDN suffix for Kubernetes services answered by coredns. | "cluster.local" | "k8s.example.com" |
| enable_reporting | Enable usage or analytics reporting to upstreams (Calico) | false | true |
| enable_aggregation | Enable the Kubernetes Aggreagation Layer | true | false |
| reservation_ids | Map Packet hardware reservation IDs to instances. | {} | { controller-0 = "55555f20-a1fb-55bd-1e11-11af11d11111" } |
| reservation_ids_default | Default hardware reservation ID for nodes not listed in the `reservation_ids` map. | "" | "next-available"|
| certs_validity_period_hours | Validity of all the certificates in hours | 8760 | 17520 |
| controller_clc_snippets [[1]](#clc-snippets-limitation) | Controller Container Linux Config snippets | [] | [example](../advanced/customization.md#usage) |


#### Worker module

| Name | Description | Default | Example |
|:-----|:------------|:--------|:--------|
| worker_count | Number of worker nodes | 1 | 3 |
| type | Type of nodes to provision | "baremetal_0" | "t1.small.x86". See https://www.packet.com/cloud/servers/ for more |
| labels | Comma separated labels to be added to the worker nodes | "" | "node.supernova.io/role=backend" |
| os_channel | Flatcar Container Linux channel to install from | "stable" | "stable", "beta", "alpha", "edge" |
| os_arch    | Flatcar Container Linux architecture to install | "amd64"  | "amd64", "arm64" |
| os_version | Version of a Flatcar Container Linux release, only for iPXE | "current" | "2191.5.0" |
| ipxe_script_url | URL that contains iPXE script to boot Flatcar on the node over PXE | "" | https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/amd64-usr/packet.ipxe, https://raw.githubusercontent.com/kinvolk/flatcar-ipxe-scripts/arm64-usr/packet.ipxe |
| cluster_domain_suffix | FQDN suffix for Kubernetes services answered by coredns. | "cluster.local" | "k8s.example.com" |
| service_cidr | CIDR IPv4 range to assign to Kubernetes services | "10.3.0.0/16" | "10.3.0.0/24" |
| setup_raid | Flag to create a RAID 0 from extra disks on a Packet node | false | true |
| setup_raid_hdd    | Flag to create a RAID 0 from extra Hard Disk Drives (HDD) only. Has no effect if `setup_raid` is `true`   | false | true |
| setup_raid_ssd    | Flag to create a RAID 0 from extra Solid State Drives (SSD) only. Has no effect if `setup_raid` is `true` | false | true |
| setup_raid_ssd_fs | Flag to create a file system on RAID 0 created using flag `setup_raid_ssd`. Has no effect if `setup_raid` is `true` | true | false |
| taints | Comma separated list of custom taints for all workers in the worker pool | "" | "clusterType=staging:NoSchedule,nodeType=storage:NoSchedule" |
| reservation_ids | Map Packet hardware reservation IDs to instances. | {} | { worker-0 = "55555f20-a1fb-55bd-1e11-11af11d11111" } |
| reservation_ids_default | Default hardware reservation ID for nodes not listed in the `reservation_ids` map. | "" | "next-available"|
| clc_snippets [[1]](#clc-snippets-limitation) | Worker Container Linux Config snippets | [] | [example](../advanced/customization.md#usage) |

Documentation about Packet hardware reservation id can be found here: https://support.packet.com/kb/articles/reserved-hardware.

#### CLC Snippets Limitation

The CLC snippepts are passsed using the user-data mechanishm. The size of it affects the time Packet needs to deploy a node in a severe way, for instance a difference of 64kB increases the deployment time by 5 minutes and a user-data bigger than 128kB could cause the deployment to timeout. Lokomotive consumes about 18kB of user-data on Packet.
Please also consider that the different vendors have different limits for the user-data, i.e. the same snippets could not work on differet providers.

See [issue #111](https://github.com/kinvolk/lokomotive-kubernetes/issues/111) for more details.

## Post-installation modification

Currently the only tested ways to modify a cluster are:

* Adding new worker pools, done by adding a new worker module.
* Scaling a worker pool by changing the `worker_count` to delete or add nodes, even to 0 (but the worker pool definition has to be kept and the total number of workers must be > 0).
* Changing the instance type of a worker pool by altering `type`, e.g., from `t1.small.x86` to `c1.small.x86`, which will recreate the nodes, causing downtime since they are destroyed first and then created again.

This list may be expanded in the future but for now other changes are not supported but can be done at your own risk.
