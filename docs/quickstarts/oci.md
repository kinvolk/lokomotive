
## Introduction

This guide shows how to create a Lokomotive cluster on [OCI](https://www.oracle.com/cloud/). By the
end of this guide, you'll have a basic Lokomotive cluster running on OCI.

Lokomotive runs on top of [Flatcar Container Linux](https://www.flatcar-linux.org/).

The guide uses [Amazon Route 53](https://aws.amazon.com/route53/) as a DNS provider. For more
information on how Lokomotive handles DNS, refer to [this](../concepts/dns.md) document.

[Lokomotive components](../concepts/components.md) complement the "stock" Kubernetes functionality
by adding features such as load balancing, persistent storage and monitoring to a cluster.

## Requirements

* An OCI account.
* An AWS account.
* An AWS
  [access key ID and secret](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html)
  of a user with
  [permissions](https://github.com/kinvolk/lokomotive/blob/master/docs/concepts/dns.md#aws-route-53)
  to edit Route 53 records.
* An AWS Route 53 zone (can be a subdomain).
* An SSH key pair for accessing the cluster nodes.
* Terraform `v0.13.x`
  [installed](https://learn.hashicorp.com/terraform/getting-started/install.html#install-terraform).
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

Build the latest `lokoctl` binary for your platform:

```console
git clone https://github.com/kinvolk/lokomotive
cd lokomotive
git checkout origin/oci
make
```

```
./lokoctl
```

Put it in the `PATH` to access it from anywhere.

### Step 2: Create a cluster configuration

Create a directory for the cluster-related files and navigate to it:

```console
mkdir lokomotive-demo && cd lokomotive-demo
```

Create a file named `cluster.lokocfg` with the following contents:

```hcl
variable "cluster_name" {}
variable "route53_zone" {}
variable "route53_zone_id" {}
variable "region" {}
variable "ssh_pubkey" {}
variable "user_ocid" {}
variable "fingerprint" {}
variable "private_key_path" {}
variable "tenancy_id" {}
variable "compartment_id" {}
variable "ad_number" {}
variable "benchmark_worker_count" {}
variable "benchmark_instance_image_id" {}
variable "benchmark_instance_type" {}
variable "benchmark_worker_cpus" {}
variable "benchmark_worker_memory" {}
variable "amd64_vm_image_id" {}
variable "generic_machine_type" {}
variable "generic_machine_cpus" {}
variable "generic_machine_memory" {}

cluster "oci" {
  asset_dir        = "./assets"
  cluster_name     = var.cluster_name
  dns_zone         = var.route53_zone
  dns_zone_id      = var.route53_zone_id
  region           = var.region
  ssh_pubkeys      = [var.ssh_pubkey]
  user             = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = pathexpand(var.private_key_path)

  tenancy_id     = var.tenancy_id
  compartment_id = var.compartment_id

  controller_count     = 1
  controller_image_id  = var.amd64_vm_image_id
  controller_type      = var.generic_machine_type
  controller_cpus      = var.generic_machine_cpus
  controller_memory    = var.generic_machine_memory
  controller_ad_number = var.ad_number

  os_arch     = "amd64"
  network_mtu = 9001

  worker_pool "general" {
    count            = 2
    image_id         = var.amd64_vm_image_id
    instance_type    = var.generic_machine_type
    worker_cpus      = var.generic_machine_cpus
    worker_memory    = var.generic_machine_memory
    worker_ad_number = var.ad_number
    ssh_pubkeys      = [var.ssh_pubkey]
    extra_volume_size = 150

    clc_snippets = [
      <<EOF
---
storage:
  files:
    - path: /etc/docker/daemon.json
      filesystem: root
      mode: 0644
      contents:
        inline: |
          {
            "mtu": 9000
          }
EOF
    ]
  }

  worker_pool "benchmark" {
    count            = var.benchmark_worker_count
    image_id         = var.benchmark_instance_image_id
    instance_type    = var.benchmark_instance_type
    worker_cpus      = var.benchmark_worker_cpus
    worker_memory    = var.benchmark_worker_memory
    worker_ad_number = var.ad_number
    ssh_pubkeys      = [var.ssh_pubkey]

    taints = {
      "role" = "benchmark:NoSchedule"
    }

    labels = {
      "role" = "benchmark"
    }

    clc_snippets = [
      <<EOF
---
storage:
  files:
    - path: /etc/docker/daemon.json
      filesystem: root
      mode: 0644
      contents:
        inline: |
          {
            "mtu": 9000
          }
EOF
    ]
  }
}

component "openebs-operator" {
  ndm_selector_label = "beta.kubernetes.io/arch"
  ndm_selector_value = "amd64"
}

component "openebs-storage-class" {
  storage-class "openebs-test-sc" {
    replica_count = 1
    default       = true
  }
}

component "prometheus-operator" {
  prometheus {
    watch_labeled_service_monitors = "false"
    watch_labeled_prometheus_rules = "false"
    storage_size                   = "100Gi"

    external_labels = {
      "cluster" = var.cluster_name
    }

    node_selector = {
      "beta.kubernetes.io/arch" = "amd64"
    }
  }

  alertmanager_node_selector = {
    "beta.kubernetes.io/arch" = "amd64"
  }
}
```

Create a file named `lokocfg.vars` with the following contents:

```tf
cluster_name                = "testlokomotive"

# OCI User specific information
user_ocid                   = "ocid1.user..."
fingerprint                 = "a1:b2:..."

private_key_path            = "/absolute/path/to/the/key/file/downloaded/from/the/oci/dashboard"
tenancy_id                  = "ocid1.tenancy...."
compartment_id              = "ocid1.compartment...."
ssh_pubkey                  = "ssh-rsa ... user@domain.com"

route53_zone                = "your-domain.net"
route53_zone_id             = "A1BCD23EF4GHI5"

region                      = "us-ashburn-1"
ad_number                   = "2"

# Machine configurations.
benchmark_worker_count      = "1"
benchmark_instance_image_id = "ocid1.image.oc1...."
benchmark_instance_type     = "BM.Standard.A1.160"
benchmark_worker_cpus       = "160"
benchmark_worker_memory     = "1024"

amd64_vm_image_id = "ocid1.image.oc1.iad...."
generic_machine_type = "VM.Standard.E3.Flex"
generic_machine_cpus = 5
generic_machine_memory = 16
```

Replace the parameters above using the following information:

- `route53_zone` - a Route 53 zone name. A subdomain will be created under this zone in the
  following format: `<cluster_name>.<zone>`
- `ssh_pubkey` - A string representing the *contents* of the public SSH keys which should be
  authorized on cluster nodes.

### Step 3: Deploy the cluster

Add a private key corresponding to one of the public keys specified in `ssh_pubkeys` to your
`ssh-agent`:

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
module.oci-lokomotive-demo.null_resource.copy-controller-secrets: Still creating... (8m30s elapsed)
module.oci-lokomotive-demo.null_resource.copy-controller-secrets: Still creating... (8m40s elapsed)
...
```

In case the deployment process seems to hang at the `copy-controller-secrets` phase for a long
time, check the following:

- Verify the correct private SSH key was added to `ssh-agent`.
- Verify that you can SSH into the created controller node from the machine running `lokoctl`.


### Permission issues

If the deployment fails due to insufficient permissions on AWS, ensure the IAM user associated with
the AWS API credentials has permissions to create records on Route 53.

[releases]: https://github.com/kinvolk/lokomotive/releases
