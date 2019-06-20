# Packet

In this tutorial, we'll create a Kubernetes v1.14.1 cluster on [Packet](https://packet.net) with [Flatcar Linux](https://www.flatcar-linux.org/). For external DNS, [Route53](https://aws.amazon.com/route53/) will be used.

We'll declare a Kubernetes cluster using the Lokomotive Terraform module. Then apply the changes to create DNS records, controller machines, worker machines and TLS assets.

Controllers are provisioned to run an `etcd-member` peer and a `kubelet` service. Workers run just a `kubelet` service. A one-time [bootkube](https://github.com/kubernetes-incubator/bootkube) bootstrap schedules the `apiserver`, `scheduler`, `controller-manager`, and `coredns` on controllers and schedules `kube-proxy` and `calico` (or `flannel`) on every node. A generated `kubeconfig` provides `kubectl` access to the cluster.

## Requirements

* Packet account, Project ID and [API key](https://support.packet.com/kb/articles/api-integrations) (Note, that the term "Auth Token" is also used to refer to the API key in the packet docs)
* AWS Account and IAM credentials
* AWS Route53 DNS Zone (registered Domain Name or delegated subdomain)
* Terraform v0.11.x and [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct) installed locally

## Terraform Setup

Install [Terraform](https://www.terraform.io/downloads.html) v0.11.x on your system.

```sh
$ terraform version
Terraform v0.11.13
```

Add the [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct) plugin binary for your system to `~/.terraform.d/plugins/`, noting the final name.

```sh
wget https://github.com/poseidon/terraform-provider-ct/releases/download/v0.3.1/terraform-provider-ct-v0.3.1-linux-amd64.tar.gz
tar xzf terraform-provider-ct-v0.3.1-linux-amd64.tar.gz
mv terraform-provider-ct-v0.3.1-linux-amd64/terraform-provider-ct ~/.terraform.d/plugins/terraform-provider-ct_v0.3.1
```

Read [concepts](/lokomotive/architecture/concepts.md) to learn about Terraform, modules, and organizing resources. Change to your infrastructure repository (e.g. `infra`).

```
cd infra/clusters
```

## Provider

### AWS

Login to your AWS IAM dashboard and find your IAM user. Select "Security Credentials" and create an access key. Save the id and secret to a file that can be referenced in configs.

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
  version = "~> 2.8.0"
  alias   = "default"

  region                  = "eu-central-1"
  shared_credentials_file = "/home/user/.config/aws/credentials"
}

provider "ct" {
  version = "0.3.1"
}

provider "local" {
  version = "~> 1.0"
  alias   = "default"
}

provider "null" {
  version = "~> 1.0"
  alias   = "default"
}

provider "template" {
  version = "~> 1.0"
  alias   = "default"
}

provider "tls" {
  version = "~> 1.0"
  alias   = "default"
}

provider "packet" {
  version = "~> 1.2"
  alias   = "default"
}
```

Additional configuration options are described in the `aws` provider [docs](https://www.terraform.io/docs/providers/aws/).

!!! tip
    AWS regions are listed in [docs](http://docs.aws.amazon.com/general/latest/gr/rande.html#ec2_region) or with `aws ec2 describe-regions`.
    The Packet facilities (i.e. data centers) list can be dynamically queried using the [API docs](https://www.packet.com/developers/api/#facilities).

### Packet

Login to your Packet account and obtain the project ID from the `Project Settings` tab. Obtain an API Key from the User settings menu. Note that project level API keys don't have all the necessary permissions for this exercise. The API key can be set in the `providers.tf` file for the `packet` provider as described in the docs [here](https://www.terraform.io/docs/providers/packet/index.html#example-usage). However this is not recommended to avoid accidentally committing API keys to version control. Instead set the env variable `PACKET_AUTH_TOKEN`.

Additional configuration options are describe in the `packet` provider [docs](https://www.terraform.io/docs/providers/packet/).

## Cluster

Define a Kubernetes cluster using the controller module [packet/flatcar-linux/kubernetes](https://github.com/kinvolk/lokomotive-kubernetes/tree/master/packet/flatcar-linux/kubernetes) and the worker module [packet/flatcar-linux/kubernetes/workers](https://github.com/kinvolk/lokomotive-kubernetes/tree/master/packet/flatcar-linux/kubernetes).

```tf
module "controller" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//packet/flatcar-linux/kubernetes?ref=<hash>"

  providers = {
    aws      = "aws.default"
    local    = "local.default"
    null     = "null.default"
    template = "template.default"
    tls      = "tls.default"
    packet   = "packet.default"
  }

  # Route53
  dns_zone    = "packet.example.com"
  dns_zone_id = "Z3PAABBCFAKEC0"

  # configuration
  ssh_keys = [
    "ssh-rsa AAAAB3Nz...",
    "ssh-rsa AAAAB3Nz...",
  ]

  asset_dir = "/home/user/.secrets/clusters/packet"

  # Packet
  cluster_name = "supernova"
  project_id   = "93fake81..."
  facility     = "ams1"

  # This must be the total of all worker pools
  worker_count              = 2
  worker_nodes_hostnames    = "${concat("${module.worker-pool-1.worker_nodes_hostname}", "${module.worker-pool-2.worker_nodes_hostname}")}"

  # optional
  controller_count = 1
  controller_type  = "t1.small.x86"

  management_cidrs = [
    "0.0.0.0/0",       # Instances can be SSH-ed into from anywhere on the internet.
  ]

  # This is different for each project on Packet and depends on the packet facility/region. Check yours from the `IPs & Networks` tab from your Packet.net account. If an IP block is not allocated yet, try provisioning an instance from the console in that region. Packet will allocate a public IP CIDR.
  node_private_cidr = "10.128.156.0/25"
}

module "worker-pool-helium" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//packet/flatcar-linux/kubernetes/workers?ref=<hash>"

  providers = {
    local    = "local.default"
    template = "template.default"
    tls      = "tls.default"
    packet   = "packet.default"
  }

  ssh_keys = [
    "ssh-rsa AAAAB3Nz...",
    "ssh-rsa AAAAB3Nz...",
  ]

  cluster_name = "supernova"
  project_id   = "93fake81..."
  facility     = "ams1"
  pool_name    = "helium"

  count = 2
  type  = "t1.small.x86"

  kubeconfig = "${module.controller.kubeconfig}"

  labels = "node.supernova.io/role=backend,node-role.kubernetes.io/backend="
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
module.aws-tempest.null_resource.bootkube-start: Still creating... (4m50s elapsed)
module.aws-tempest.null_resource.bootkube-start: Still creating... (5m0s elapsed)
module.aws-tempest.null_resource.bootkube-start: Creation complete after 11m8s (ID: 3961816482286168143)

Apply complete! Resources: 98 added, 0 changed, 0 destroyed.
```

In 5-10 minutes, the Kubernetes cluster will be ready.

## Verify

[Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) on your system. Use the generated `kubeconfig` credentials to access the Kubernetes cluster and list nodes.

```
$ export KUBECONFIG=/home/user/.secrets/clusters/tempest/auth/kubeconfig
$ kubectl get nodes
NAME                       STATUS  ROLES              AGE  VERSION
supernova-controller-0     Ready   controller,master  10m  v1.14.1
supernova-helium-worker-0  Ready   node               10m  v1.14.1
supernova-helium-worker-1  Ready   node               10m  v1.14.1
```

## Going Further

Learn about [maintenance](/lokomotive/topics/maintenance/).

## Variables

Check the [variables.tf](https://github.com/kinvolk/lokomotive-kubernetes/blob/master/packet/flatcar-linux/kubernetes/variables.tf) source.

### Required

#### Controller module

| Name | Description | Example |
|:-----|:------------|:--------|
| cluster_name | Unique cluster name (prepended to dns_zone) | "tempest" |
| dns_zone | AWS Route53 DNS zone | "aws.example.com" |
| dns_zone_id | AWS Route53 DNS zone id | "Z3PAABBCFAKEC0" |
| ssh_keys | List of SSH public keys for user 'core' | ["ssh-rsa AAAAB3NZ..."] |
| asset_dir | Path to a directory where generated assets should be placed (contains secrets) | "/home/user/.secrets/clusters/tempest" |
| project_id | Project ID obtained from the Packet account | "93fake81-0f3c1-..." |
| facility | Packet Region in which the instance(s) should be deployed | https://www.packet.com/developers/api/#facilities. Eg: "ams1" |
| worker_count | Total number of workers across all worker pools | 2 |
| worker_nodes_hostnames | List of hostnames of all worker nodes | ["foo-pool1-worker-0", "foo-pool1-worker-1"]
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

Clusters create a DNS A record `${cluster_name}.${dns_zone}` to resolve a network load balancer backed by controller instances. This FQDN is used by workers and `kubectl` to access the apiserver(s). In this example, the cluster's apiserver would be accessible at `tempest.aws.example.com`.

You'll need a registered domain name or delegated subdomain on AWS Route53. You can set this up once and create many clusters with unique names.

```tf
resource "aws_route53_zone" "zone-for-clusters" {
  name = "aws.example.com."
}
```

Reference the DNS zone id with `"${aws_route53_zone.zone-for-clusters.zone_id}"`.

!!! tip ""
    If you have an existing domain name with a zone file elsewhere, just delegate a subdomain that can be managed on Route53 (e.g. aws.mydomain.com) and [update nameservers](http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/SOA-NSrecords.html).

### Optional

#### Controller module

| Name | Description | Default | Example |
|:-----|:------------|:--------|:--------|
| controller_count | Number of controllers (i.e. masters) | 1 | 1 |
| controller_type | Type of nodes to provision | "baermetal_0" | "t1.small.x86". See https://www.packet.com/developers/api/#plans for more |
| os_channel | AMI channel for a Flatcar Linux | stable | stable, beta, alpha |
| os_version | Version of a Flatcar Linux release | current | 2079.3.1 |
| networking | Choice of networking provider | "calico" | "calico" or "flannel" |
| network_mtu | CNI interface MTU (calico only) | 1480 | 8981 |
| pod_cidr | CIDR IPv4 range to assign to Kubernetes pods | "10.2.0.0/16" | "10.22.0.0/16" |
| service_cidr | CIDR IPv4 range to assign to Kubernetes services | "10.3.0.0/16" | "10.3.0.0/24" |
| cluster_domain_suffix | FQDN suffix for Kubernetes services answered by coredns. | "cluster.local" | "k8s.example.com" |
| enable_reporting | Enable usage or analytics reporting to upstreams (Calico) | false | true |
| enable_aggregation | Enable the Kubernetes Aggreagation Layer | false | true |

#### Worker module

| Name | Description | Default | Example |
|:-----|:------------|:--------|:--------|
| count | Number of worker nodes | 1 | 3 |
| type | Type of nodes to provision | "baremetal_0" | "t1.small.x86". See https://www.packet.com/developers/api/#plans for more |
| labels | Comma separated labels to be added to the worker nodes | "" | "node.supernova.io/role=backend" |
| os_channel | AMI channel for a Flatcar Linux | stable | stable, beta, alpha |
| os_version | Version of a Flatcar Linux release | current | 2079.3.1 |
| cluster_domain_suffix | FQDN suffix for Kubernetes services answered by coredns. | "cluster.local" | "k8s.example.com" |
| service_cidr | CIDR IPv4 range to assign to Kubernetes services | "10.3.0.0/16" | "10.3.0.0/24" |
| setup_raid | Flag to create a RAID 0 from extra disks on a Packet node | "false" | "true" |
| taints | Comma separated list of custom taints for all workers in the worker pool | "" | "clusterType=staging:NoSchedule,nodeType=storage:NoSchedule" |
| reservation_ids | Map Packet hardware reservation IDs to instances. | {} | { worker-0 = "55555f20-a1fb-55bd-1e11-11af11d11111" } |
| reservation_ids_default | Default hardware reservation ID for nodes not listed in the `reservation_ids` map. | "" | "next-available"|

Documentation about Packet hardware reservation id can be found here: https://support.packet.com/kb/articles/reserved-hardware.
