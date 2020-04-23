# Lokomotive Packet quickstart guide

## Contents

* [Introduction](#introduction)
* [Requirements](#requirements)
* [Step 1: Install lokoctl](#step-1-install-lokoctl)
* [Step 2: Set up a working directory](#step-2-set-up-a-working-directory)
* [Step 3: Set up credentials from environment variables](#step-3-set-up-credentials-from-environment-variables)
* [Step 4: Define cluster configuration](#step-4-define-cluster-configuration)
* [Step 5: Create Lokomotive cluster](#step-5-create-lokomotive-cluster)
* [Verification](#verification)
* [Cleanup](#cleanup)
* [Troubleshooting](#troubleshooting)
* [Conclusion](#conclusion)
* [Next steps](#next-steps)

## Introduction

This guide shows how to create a Lokomotive cluster on [Packet](https://www.packet.com/). By the
end of this guide, you'll have a basic Lokomotive cluster running on Packet with a demo application
deployed.

The guide uses `t1.small.x86` as the Packet device type for all created nodes. This is also the
default device type.

Lokomotive runs on top of [Flatcar Container Linux](https://www.flatcar-linux.org/). This guide
uses the `stable` channel.

The guide uses [Amazon Route 53](https://aws.amazon.com/route53/) as a DNS provider. For more
information on how Lokomotive handles DNS, refer to [this](../concepts/dns.md) document.

[Lokomotive components](../concepts/components.md) complement the "stock" Kubernetes functionality
by adding features such as load balancing, persistent storage and monitoring to a cluster. To keep
this guide short you will deploy a single component - `httpbin` - which serves as a demo
application to verify the cluster behaves as expected.

## Requirements

* A Packet account with a project created and
  [local BGP](https://www.packet.com/developers/docs/network/advanced/local-and-global-bgp/)
  enabled.
* A Packet project ID.
* A Packet
  [user level API key](https://www.packet.com/developers/docs/API/getting-started/)
  with access to the relevant project.
* An AWS account.
* An AWS
  [access key ID and secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html)
  of a user with
  [permissions](https://github.com/kinvolk/lokomotive/blob/master/docs/concepts/dns.md#aws-route-53)
  to edit Route 53 records.
* An AWS Route 53 zone (can be a subdomain).
* An SSH key pair for accessing the cluster nodes.
* Terraform `v0.12.x`
  [installed](https://learn.hashicorp.com/terraform/getting-started/install.html#install-terraform).
* The `ct` Terraform provider `v0.5.0`
  [installed](https://github.com/poseidon/terraform-provider-ct/blob/v0.5.0/README.md#install).
* `kubectl` [installed](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

>NOTE: The `kubectl` version used to interact with a Kubernetes cluster needs to be compatible with
>the version of the Kubernetes control plane. Ideally you should install a `kubectl` binary whose
>version is identical to the Kubernetes control plane included with a Lokomotive release. However,
>some degree of version "skew" is tolerated - see the Kubernetes
>[version skew policy](https://kubernetes.io/docs/setup/release/version-skew-policy/) document for
>more information. You can determine the version of the Kubernetes control plane included with a
>Lokomotive release by looking at the [release notes][releases].

## Steps

### Step 1: Install lokoctl

`lokoctl` is the command-line interface for managing Lokomotive clusters.

Download the latest `lokoctl` binary for your platform:

```console
export os=linux  # For macOS, use `os=darwin`.

export release=$(curl -s https://api.github.com/repos/kinvolk/lokomotive/releases | jq -r '.[0].name' | tr -d v)
curl -LO "https://github.com/kinvolk/lokomotive/releases/download/v${release}/lokoctl_${release}_${os}_amd64.tar.gz"
```

Extract the binary and copy it to a place under your `$PATH`:

```console
tar zxvf lokoctl_${release}_${os}_amd64.tar.gz
sudo cp lokoctl_${release}_${os}_amd64/lokoctl /usr/local/bin
rm -rf lokoctl_${release}_${os}_amd64*
```

### Step 2: Create a cluster configuration

Create a directory for the cluster-related files and navigate to it:

```console
mkdir lokomotive-demo && cd lokomotive-demo
```

Create a file named `cluster.lokocfg` with the following contents:

```hcl
cluster "packet" {
  asset_dir        = "./assets"
  cluster_name     = "lokomotive-demo"

  dns {
    zone     = "example.com"
    provider = "route53"
  }

  facility = "ams1"
  project_id = "89273817-4f44-4b41-9f0c-cb00bf538542"

  ssh_pubkeys       = ["ssh-rsa AAAA..."]
  management_cidrs  = ["0.0.0.0/0"]
  node_private_cidr = "10.0.0.0/8"

  controller_count = 1

  worker_pool "pool-1" {
    count       = 2
  }
}

# A demo application.
component "httpbin" {
  ingress_host = "httpbin.example.com"
}
```

Replace the parameters above using the following information:

- `dns.zone` - a Route 53 zone name. A subdomain will be created under this zone in the following
  format: `<cluster_name>.<zone>`
- `project_id` - the Packet project ID to deploy the cluster in.
- `ssh_pubkeys` - A list of strings representing the *contents* of the public SSH keys which should
  be authorized on cluster nodes.

The rest of the parameters may be left as-is. For more information about the configuration options
see the [configuration reference](../configuration-reference/platforms/packet.md).

### Step 3: Deploy the cluster

>NOTE: If you have the AWS CLI installed and configured for an AWS account, you can skip setting
>the `AWS_*` variables below. `lokoctl` follows the standard AWS authentication methods, which
>means it will use the `default` AWS CLI profile if no explicit credentials are specified.
>Similarly, environment variables such as `AWS_PROFILE` can be used to instruct `lokoctl` to use a
>specific AWS CLI profile for AWS authentication.

Set up your Packet and AWS credentials in your shell:

```console
export PACKET_AUTH_TOKEN=k84jfL83kJF849B776Nle4L3980fake
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7FAKE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYFAKE
```

Add a private key corresponding to one of the public keys specified in `ssh_pubkeys` to your `ssh-agent`:

```bash
ssh-add ~/.ssh/id_rsa
ssh-add -L
```

Deploy the cluster:

```console
lokoctl cluster apply -v
```

The deployment process typically takes about 15 minutes. Upon successful completion, an output
similar to the following is shown:

```
Your configurations are stored in ./assets

Now checking health and readiness of the cluster nodes ...

Node                             Ready    Reason          Message                            
                                                                                             
lokomotive-demo-controller-0       True     KubeletReady    kubelet is posting ready status    
lokomotive-demo-pool-1-worker-0    True     KubeletReady    kubelet is posting ready status    
lokomotive-demo-pool-1-worker-1    True     KubeletReady    kubelet is posting ready status    

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

**Note**: Lokomotive sets up a pretty restrictive Pod Security Policy that
disallows running containers as root by default, check the [Pod Security Policy
documentation](../concepts/securing-lokomotive-cluster.md#cluster-wide-pod-security-policy)
for more details.

## Cleanup

To destroy the Lokomotive cluster, execute the following command:

```console
lokoctl cluster destroy --confirm
```

You can safely delete the working directory created for this quickstart guide if you no longer
require it.

## Troubleshooting

### Stuck at copy controller secrets

If there is an execution error or no progress beyond the output provided below:

```console
...
module.packet-mycluster.null_resource.copy-controller-secrets: Still creating... (8m30s elapsed)
module.packet-mycluster.null_resource.copy-controller-secrets: Still creating... (8m40s elapsed)
...
```

The error probably happens because the `ssh_pubkeys` provided in the configuration is missing in the
`ssh-agent`.

To rectify the error, you need to:

1. Follow the steps [to add the SSH key to the
   ssh-agent](https://help.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent#adding-your-ssh-key-to-the-ssh-agent).
2. Retry [Step 5](#step-5-create-lokomotive-cluster).

### Packet provisioning failed

For failed machine provisioning on Packet end, retry [Step 5](#step-5-create-lokomotive-cluster).

### Insufficient availability of nodes types on Packet

In the event of failed Packet provisioning due to machines of type `controller_type` or
`workers_type` not available.  You can check the Packet API [capacity
endpoint](https://www.packet.com/developers/api/capacity/) to get the current capacity and decide on
changing the facility or the machine type.

### Permission issues

  * If the failure is due to insufficient permissions on Packet, check the permission on the Packet
    console.
  * This generally happens if user is using `Project Level API Key` and not `User Level API Key`.

### Failed installation of components that require disk storage

For components that require disk storage such as [Openebs storage
class](../configuration-reference/components/openebs-storage-class.md), [Prometheus
Operator](../configuration-reference/components/prometheus-operator.md) machine types with spare disks
should be used.

## Conclusion

After walking through this guide, you've learned how to set up a Lokomotive cluster on Packet.

## Next steps

You can now start deploying your workloads on the cluster.

For more information on installing supported Lokomotive components, you can visit the [component
configuration references](../configuration-reference/components).
[releases]: https://github.com/kinvolk/lokomotive/releases
