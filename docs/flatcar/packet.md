# Packet

In this tutorial, we'll create a Kubernetes v1.13.1 cluster on [Packet](https://packet.net) with [Flatcar Linux](https://www.flatcar-linux.org/). For external DNS, [Route53](https://aws.amazon.com/route53/) will be used.

We'll declare a Kubernetes cluster using the Typhoon Terraform module. Then apply the changes to create DNS records, controller machines, worker machines and TLS assets.

Controllers are provisioned to run an `etcd-member` peer and a `kubelet` service. Workers run just a `kubelet` service. A one-time [bootkube](https://github.com/kubernetes-incubator/bootkube) bootstrap schedules the `apiserver`, `scheduler`, `controller-manager`, and `coredns` on controllers and schedules `kube-proxy` and `calico` (or `flannel`) on every node. A generated `kubeconfig` provides `kubectl` access to the cluster.

!!! note
    Currently Route53 is the only supported external DNS provider for Packet deployments and is automatically configured. Further providers may be added in the future.

## Requirements

* Packet account and [API key](https://support.packet.com/kb/articles/api-integrations)
* AWS Account and IAM credentials
* AWS Route53 DNS Zone (registered Domain Name or delegated subdomain)
* Terraform v0.11.x and [terraform-provider-ct](https://github.com/coreos/terraform-provider-ct) installed locally

## Terraform Setup

Install [Terraform](https://www.terraform.io/downloads.html) v0.11.x on your system.

```sh
$ terraform version
Terraform v0.11.11
```

Add the [terraform-provider-ct](https://github.com/coreos/terraform-provider-ct) plugin binary for your system to `~/.terraform.d/plugins/`, noting the final name.

```sh
wget https://github.com/coreos/terraform-provider-ct/releases/download/v0.3.0/terraform-provider-ct-v0.3.0-linux-amd64.tar.gz
tar xzf terraform-provider-ct-v0.3.0-linux-amd64.tar.gz
mv terraform-provider-ct-v0.3.0-linux-amd64/terraform-provider-ct ~/.terraform.d/plugins/terraform-provider-ct_v0.3.0
```

Read [concepts](/architecture/concepts/) to learn about Terraform, modules, and organizing resources. Change to your infrastructure repository (e.g. `infra`).

```
cd infra/clusters
```

## Provider

Login to your AWS IAM dashboard and find your IAM user. Select "Security Credentials" and create an access key. Save the id and secret to a file that can be referenced in configs.

```
[default]
aws_access_key_id = xxx
aws_secret_access_key = yyy
```

!!! tip
    Other standard AWS authentication methods can be used instead of specifying `shared_credentials_file` under the provider's config. For more information see the [docs](https://www.terraform.io/docs/providers/aws/#authentication).

Login to your Packet account and generate a project-level API key under "Project Settings" or a user-level API key under your user settings. Set the key as an environment variable which Terraform will automatically pick up.

```
export PACKET_AUTH_TOKEN=<api_key>
```

Configure the AWS provider to use your access key credentials in a `providers.tf` file. Alternatively, omit `shared_credentials_file` to use a different form of AWS authentication.

```tf
provider "aws" {
  version = "~> 1.57.0"
  alias   = "default"

  region                  = "eu-central-1"
  shared_credentials_file = "/home/user/.config/aws/credentials"
}

provider "ct" {
  version = "0.3.0"
}

provider "local" {
  version = "~> 1.0"
  alias = "default"
}

provider "null" {
  version = "~> 1.0"
  alias = "default"
}

provider "template" {
  version = "~> 1.0"
  alias = "default"
}

provider "tls" {
  version = "~> 1.0"
  alias = "default"
}

provider "packet" {
  version = "~> 1.4"
  alias = "default"
}
```

Additional configuration options are described in the `aws` provider [docs](https://www.terraform.io/docs/providers/aws/) and the `packet` provider [docs](https://www.terraform.io/docs/providers/packet/).

!!! tip
    AWS regions are listed in [docs](http://docs.aws.amazon.com/general/latest/gr/rande.html#ec2_region) or with `aws ec2 describe-regions`.
    The Packet facilities (i.e. data centers) list can be dynamically queried using the [API docs](https://www.packet.com/developers/api/#facilities).

## Cluster

Define a Kubernetes cluster using the module `packet/flatcar-linux/kubernetes`.

!!! todo
    Update the module's `source` reference to a permanent Git ref.

```tf
module "packet-lithium" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//packet/flatcar-linux/kubernetes"

  providers = {
    aws = "aws.default"
    local = "local.default"
    null = "null.default"
    template = "template.default"
    tls = "tls.default"
    packet = "packet.default"
  }

  # Route53
  dns_zone     = "aws.example.com"
  dns_zone_id  = "Z3PAABBCFAKEC0"

  # configuration
  ssh_keys = ["ssh-rsa AAAAB3Nz...", "ssh-rsa AAAAB3Nz..."]
  asset_dir = "/home/user/.secrets/clusters/lithium"

  # Packet
  cluster_name = "lithium"
  project_id = "4cff83ac-de23-432a-b01b-b2950dabc76e"
  facility = "ams1"

  # optional
  controller_count = 1
  worker_count = 2
  worker_type  = "baremetal_0"
}
```

Reference the [variables docs](#variables) or the [variables.tf](https://github.com/kinvolk/lokomotive-kubernetes/blob/master/packet/flatcar-linux/kubernetes/variables.tf) source.

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
module.packet-lithium.null_resource.bootkube-start: Still creating... (4m50s elapsed)
module.packet-lithium.null_resource.bootkube-start: Still creating... (5m0s elapsed)
module.packet-lithium.null_resource.bootkube-start: Creation complete after 11m8s (ID: 3961816482286168143)

Apply complete! Resources: 98 added, 0 changed, 0 destroyed.
```

In 5-10 minutes, the Kubernetes cluster will be ready.

## Verify

[Install kubectl](https://coreos.com/kubernetes/docs/latest/configure-kubectl.html) on your system. Use the generated `kubeconfig` credentials to access the Kubernetes cluster and list nodes.

```
$ export KUBECONFIG=/home/user/.secrets/clusters/lithium/auth/kubeconfig
$ kubectl get nodes
NAME           STATUS  ROLES              AGE  VERSION
ip-10-0-3-155  Ready   controller,master  10m  v1.13.1
ip-10-0-26-65  Ready   node               10m  v1.13.1
ip-10-0-41-21  Ready   node               10m  v1.13.1
```

List the pods.

```
$ kubectl get pods --all-namespaces
NAMESPACE     NAME                                      READY  STATUS    RESTARTS  AGE              
kube-system   calico-node-1m5bf                         2/2    Running   0         34m              
kube-system   calico-node-7jmr1                         2/2    Running   0         34m              
kube-system   calico-node-bknc8                         2/2    Running   0         34m              
kube-system   coredns-1187388186-wx1lg                  1/1    Running   0         34m              
kube-system   coredns-1187388186-qjnvp                  1/1    Running   0         34m
kube-system   kube-apiserver-4mjbk                      1/1    Running   0         34m              
kube-system   kube-controller-manager-3597210155-j2jbt  1/1    Running   1         34m              
kube-system   kube-controller-manager-3597210155-j7g7x  1/1    Running   0         34m              
kube-system   kube-proxy-14wxv                          1/1    Running   0         34m              
kube-system   kube-proxy-9vxh2                          1/1    Running   0         34m              
kube-system   kube-proxy-sbbsh                          1/1    Running   0         34m              
kube-system   kube-scheduler-3359497473-5plhf           1/1    Running   0         34m              
kube-system   kube-scheduler-3359497473-r7zg7           1/1    Running   1         34m              
kube-system   pod-checkpointer-4kxtl                    1/1    Running   0         34m              
kube-system   pod-checkpointer-4kxtl-ip-10-0-3-155      1/1    Running   0         33m
```

## Going Further

Learn about [maintenance](/topics/maintenance/) and [addons](/addons/overview/).

!!! note
    On Container Linux clusters, install the `CLUO` addon to coordinate reboots and drains when nodes auto-update. Otherwise, updates may not be applied until the next reboot.

## Variables

Check the [variables.tf](https://github.com/kinvolk/lokomotive-kubernetes/blob/master/packet/flatcar-linux/kubernetes/variables.tf) source.

### Required

| Name | Description | Example |
|:-----|:------------|:--------|
| cluster_name | Unique cluster name (prepended to dns_zone) | "lithium" |
| dns_zone | AWS Route53 DNS zone | "aws.example.com" |
| dns_zone_id | AWS Route53 DNS zone id | "Z3PAABBCFAKEC0" |
| ssh_keys | A list of SSH public keys for user 'core' | ["ssh-rsa AAAAB3NZ...", "ssh-rsa AAAAB3NZ..."] |
| asset_dir | Path to a directory where generated assets should be placed (contains secrets) | "/home/user/.secrets/clusters/lithium" |
| project_id | Packet project ID | "4cff83ac-de23-432a-b01b-b2950dabc76e" |
| facility | Packet facility (data center) in which to deploy the cluster | "ams1" |
| management_cidrs | List of IPv4 CIDRs authorized to access or manage the cluster | ["1.2.3.4/32"] |
| node_private_cidr | Private IPv4 CIDR of the nodes used to allow inter-node traffic | "10.80.123.128/25" |

#### DNS Zone

Clusters create a DNS A record `${cluster_name}.${dns_zone}` to resolve controller instances. This FQDN is used by workers and `kubectl` to access the apiserver(s). In this example, the cluster's apiserver would be accessible at `lithium.aws.example.com`.

You'll need a registered domain name or delegated subdomain on AWS Route53. You can set this up once and create many clusters with unique names.

```tf
resource "aws_route53_zone" "zone-for-clusters" {
  name = "aws.example.com."
}
```

Reference the DNS zone id with `"${aws_route53_zone.zone-for-clusters.zone_id}"`.

!!! tip ""
    If you have an existing domain name with a zone file elsewhere, just delegate a subdomain that can be managed on Route53 (e.g. aws.mydomain.com) and [update nameservers](http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/SOA-NSrecords.html).

#### Project ID

Your Packet project ID can be obtained by navigating to "Project Settings" on the Packet console.

!!! warning
    Deployments with multiple controllers haven't been tested yet!

### Optional

| Name | Description | Default | Example |
|:-----|:------------|:--------|:--------|
| controller_count | Number of controllers (i.e. masters) | 1 | 1 |
| worker_count | Number of workers | 1 | 3 |
| controller_type | Packet server type for controllers | "baremetal_0" | See below |
| worker_type | Packet server type for workers | "baremetal_0" | See below |
| ipxe_script_url | Custom iPXE script to use for booting the machines | See below | See below |
| networking | Choice of networking provider | "calico" | "calico" or "flannel" |
| network_mtu | CNI interface MTU (calico only) | 1480 | 8981 |
| network_ip_autodetection_method | Method to detect host IPv4 address (calico-only) | first-found | can-reach=10.0.0.1 |
| pod_cidr | CIDR IPv4 range to assign to Kubernetes pods | "10.2.0.0/16" | "10.22.0.0/16" |
| service_cidr | CIDR IPv4 range to assign to Kubernetes services | "10.3.0.0/16" | "10.3.0.0/24" |
| cluster_domain_suffix | FQDN suffix for Kubernetes services answered by coredns. | "cluster.local" | "k8s.example.com" |
| enable_reporting | Enable usage or analytics reporting to upstreams (Calico) | false | true |

Check the list of valid Packet [plans](https://www.packet.com/developers/api/#plans) (server types).

The latest stable [release](https://www.flatcar-linux.org/releases/) of Flatcar Linux will be used by default for the nodes. To use a different OS channel or a pervious release, host an iPXE script in an internet-accessible location and set `ipxe_script_url` to point at the correct URL.
