# Lokomotive Bare Metal quickstart guide

## Contents

* [Introduction](#introduction)
* [Requirements](#requirements)
* [Step 1: Install lokoctl](#step-1-install-lokoctl)
* [Step 2: Set up a working directory](#step-2-set-up-a-working-directory)
* [Step 3: Machines, DNS and Matchbox set up](#step-3-machines-dns-and-matchbox-set-up)
* [Step 4: Define Cluster Configuration](#step-4-define-cluster-configuration)
* [Step 5: Create Lokomotive cluster](#step-5-create-lokomotive-cluster)
* [Verification](#verification)
* [Cleanup](#cleanup)
* [Troubleshooting](#troubleshooting)
* [Conclusion](#conclusion)
* [Next steps](#next-steps)

## Introduction

This quickstart guide walks through the steps needed to create a Lokomotive cluster on bare metal with
Flatcar Container Linux utilizing PXE.

By the end of this guide, you'll have a working Kubernetes cluster with 1 controller node and 2
worker nodes.

## Requirements

* Basic understanding of Kubernetes concepts.
* Terraform v0.12.x, [terraform-provider-matchbox](https://github.com/poseidon/terraform-provider-matchbox)
and [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct) v0.5.0 installed locally.
* Machines with at least 2GB RAM, 30GB disk, PXE-enabled NIC and IPMI.
* PXE-enabled [network boot](https://coreos.com/matchbox/docs/latest/network-setup.html) environment.
* Matchbox v0.6+ deployment with API enabled.
* Matchbox credentials `client.crt`, `client.key`, `ca.crt`.
* An SSH key pair for management access.
* `kubectl` installed locally to access the Kubernetes cluster.

Note that the machines should only be powered on after starting the installation, see below.

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
mkdir -p lokomotive-infra/mybaremetalcluster
cd lokomotive-infra/mybaremetalcluster
```

### Step 3: Machines, DNS and Matchbox set up

#### Machines

Mac addresses collected from each machine.

For machines with multiple PXE-enabled NICs, pick one of the MAC addresses. MAC addresses will be
used to match machines to profiles during network boot.

Example:

```console
52:54:00:a1:9c:ae (node1)
52:54:00:b2:2f:86 (node2)
52:54:00:c3:61:77 (node3)
```

#### DNS

Create DNS A (or AAAA) record for each node's default interface.

Cluster nodes will be configured to refer to the control plane and themselves by these fully
qualified names and they will be used in generated TLS certificates.

Example:

```console
node1.example.com (node1)
node2.example.com (node2)
node3.example.com (node3)
```

#### Matchbox

One of the requirements is to have [Matchbox](https://github.com/poseidon/matchbox) with TLS enabled
deployed.

Verify the Matchbox read-only HTTP endpoints are accessible.

```console
curl http://matchbox.example.com:8080
matchbox
```

Verify your TLS client certificate and key can be used to access the Matchbox API.

```console
openssl s_client -connect matchbox.example.com:8081 \
  -CAfile /path/to/matchbox/ca.crt \
  -cert /path/to/matchbox/client.crt \
  -key /path/to/matchbox/client.key
```


### Step 4: Define Cluster Configuration

To create a Lokomotive cluster, we need to define a configuration.

Create a file with the extension `.lokocfg` with the contents below:

```tf
# baremetalcluster.lokocfg
cluster "bare-metal" {
  # Change the location where lokoctl stores the cluster assets.
  asset_dir = "./lokomotive-assets"

  # Cluster name.
  cluster_name = baremetalcluster

  # SSH Public keys.
  ssh_pubkeys = [
    "ssh-rsa AAAAB3Nz...",
  ]
  # Whether the operating system should PXE boot and install from matchbox /assets cache.
  cached_install = "true"

  # Matchbox CA crt path.
  matchbox_ca_path = pathexpand("/path/to/matchbox/ca.crt")

  # Matchbox client crt path.
  matchbox_client_cert_path = pathexpand("/path/to/matchbox/client.crt")

  # Matchbox client key path.
  matchbox_client_key_path = pathexpand("/path/to/matchbox/client.key")

  # Matchbox https endpoint.
  matchbox_endpoint = "matchbox.example.com:8081"

  # Matchbox HTTP read-only endpoint.
  matchbox_http_endpoint = "http://matchbox.example.com:8080"

  # Domain name.
  k8s_domain_name = "node1.example.com"

  # FQDN of controller nodes.
  controller_domains = [
    "node1.example.com",
  ]

  # MAC addresses of controllers.
  controller_macs = [
    "52:54:00:a1:9c:ae",
  ]

  # Names of the controller nodes.
  controller_names = [
    "node1",
  ]

  # FQDN of worker nodes.
  worker_domains = [
    "node2.example.com",
    "node3.example.com",
  ]

  # Mac addresses of worker nodes.
  worker_macs = [
    "52:54:00:b2:2f:86",
    "52:54:00:c3:61:77",
  ]

  # Names of the worker nodes.
  worker_names = [
    "node2",
    "node3",
  ]
}

```

For advanced cluster configurations and more information refer to the [Bare Metal configuration
guide](../configuration-reference/platform/baremetal.md).

### Step 5: Create Lokomotive Cluster

Add a private key corresponding to one of the public keys specified in `ssh_pubkeys` to your `ssh-agent`:

```bash
ssh-add ~/.ssh/id_rsa
ssh-add -L
```

Run the following command to create the cluster:

```console
lokoctl cluster apply
```

**Proceed to Power on the PXE machines while this loops.**

Once the command finishes, your Lokomotive cluster details are stored in the path you've specified
under `asset_dir`.

## Verification

A successful installation results in the output:

```console
module.baremetal-baremetalcluster.null_resource.bootkube-start: Still creating... [4m10s elapsed]
module.baremetal-baremetalcluster.null_resource.bootkube-start: Still creating... [4m20s elapsed]
module.baremetal-baremetalcluster.null_resource.bootkube-start: Creation complete after 4m25s [id=1122239320434737682]

Apply complete! Resources: 74 added, 0 changed, 0 destroyed.

Your configurations are stored in /home/imran/lokoctl-assets/mycluster

Now checking health and readiness of the cluster nodes ...

Node                                          Ready    Reason          Message

node1.example.com                             True     KubeletReady    kubelet is posting ready status
node2.example.com                             True     KubeletReady    kubelet is posting ready status
node3.example.com                             True     KubeletReady    kubelet is posting ready status

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

### Stuck At Copy Controller Secrets

If there is an execution error or no progress beyond the output provided below:

```console
...
module.baremetal-baremetalcluster.null_resource.copy-controller-secrets: Still creating... (8m30s elapsed)
module.baremetal-baremetalcluster.null_resource.copy-controller-secrets: Still creating... (8m40s elapsed)
...
```

The error probably happens because the `ssh_pubkeys` provided in the configuration is missing in the
`ssh-agent`.

To rectify the error, you need to:

1. Follow the steps [to add the SSH key to the
   ssh-agent](https://help.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent#adding-your-ssh-key-to-the-ssh-agent).
2. Retry [Step 5](#step-5-create-lokomotive-cluster).

## Conclusion

After walking through this guide, you've learned how to set up a Lokomotive cluster on Bare Metal.

## Next steps

You can now start deploying your workloads on the cluster.

For more information on installing supported Lokomotive components, you can visit the [component configuration
guides](../configuration-reference/components).
