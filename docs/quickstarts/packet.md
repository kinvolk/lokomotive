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

## Verification

Use the generated `kubeconfig` file to access the cluster:

```console
export KUBECONFIG=$(pwd)/assets/cluster-assets/auth/kubeconfig
kubectl get nodes
```

Sample output:

```
NAME                            STATUS   ROLES    AGE   VERSION
lokomotive-demo-controller-0      Ready    <none>   33m   v1.17.4
lokomotive-demo-pool-1-worker-0   Ready    <none>   33m   v1.17.4
lokomotive-demo-pool-1-worker-1   Ready    <none>   33m   v1.17.4
```

Verify all pods are ready:

```console
kubectl get pods -A
```

Verify you can access httpbin:

```console
kubectl -n httpbin port-forward svc/httpbin 8080

# In a new terminal.
curl http://localhost:8080/get
```

Sample output:

```
{
  "args": {}, 
  "headers": {
    "Accept": "*/*", 
    "Host": "localhost:8080", 
    "User-Agent": "curl/7.70.0"
  }, 
  "origin": "127.0.0.1", 
  "url": "http://localhost:8080/get"
}
```

## Using the cluster

At this point you should have access to a Lokomotive cluster and can use it to deploy applications.

If you don't have any Kubernetes experience, you can check out the [Kubernetes
Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/deploy-app/deploy-intro/) tutorial.

>NOTE: Lokomotive uses a relatively restrictive Pod Security Policy by default. This policy
>disallows running containers as root. Refer to the
>[Pod Security Policy documentation](../concepts/securing-lokomotive-cluster.md#cluster-wide-pod-security-policy)
>for more details.

## Cleanup

To destroy the cluster, execute the following command:

```console
lokoctl cluster destroy -v
```

Confirm you want to destroy the cluster by typing `yes` and hitting **Enter**.

You can now safely delete the directory created for this guide if you no longer need it.

## Troubleshooting

### Stuck at "copy controller secrets"

```
...
module.packet-lokomotive-demo.null_resource.copy-controller-secrets: Still creating... (8m30s elapsed)
module.packet-lokomotive-demo.null_resource.copy-controller-secrets: Still creating... (8m40s elapsed)
...
```

In case the deployment process seems to hang at the `copy-controller-secrets` phase for a long
time, check the following:

- Verify the correct private SSH key was added to `ssh-agent`.
- Verify that you can SSH into the created controller node from the machine running `lokoctl`.

### Packet provisioning failed

Sometimes the provisioning of servers on Packet may fail, in which case the following error is
shown:

```
Error: provisioning time limit exceeded; the Packet team will investigate
```

In this case, retrying the deployment by re-running `lokoctl cluster apply -v` may help.

### Insufficient capacity on Packet

Sometimes there may not be enough hardware available at a given Packet facility for a given machine
type, in which case the following error is shown:

```
The facility ams1 has no provisionable t1.small.x86 servers matching your criteria
```

In this case, either select a different node type and/or Packet facility, or wait for a while until
more capacity becomes available. You can check the current capacity status on the Packet
[API](https://www.packet.com/developers/api/capacity/).

### Permission issues

If the deployment fails due to insufficient permissions on Packet, verify your Packet API key has
permissions to the right Packet project.

If the deployment fails due to insufficient permissions on AWS, ensure the IAM user associated with
the AWS API credentials has permissions to create records on Route 53.

## Next steps

In this guide you used port forwarding to communicate with a sample application on the cluster.
However, in real-world cases you may want to expose your applications to the internet.
[This](../how-to-guides/ingress-with-contour-metallb.md) guide explains how to use MetalLB and
Contour to expose applications on Packet clusters to the internet.

[releases]: https://github.com/kinvolk/lokomotive/releases
