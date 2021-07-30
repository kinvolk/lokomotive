---
title: Lokomotive Tinkerbell quickstart guide
weight: 10
---

## Introduction

This guide shows how to create a Lokomotive cluster using [Tinkerbell]. By the
end of this guide, you'll have a basic Lokomotive cluster running on bare metal machines provisioned
using Tinkerbell with a demo application deployed.

Lokomotive runs on top of [Flatcar Container Linux](https://www.flatcar-linux.org/). This guide
uses the `stable` channel.

This guide uses experimental Tinkerbell sandbox running locally using [libvirt] virtual machines.

Lokomotive can store Terraform state [locally](../configuration-reference/backend/local.md)
or remotely within an [AWS S3 bucket](../configuration-reference/backend/s3.md). By default, Lokomotive
stores Terraform state locally.

[Lokomotive components](../concepts/components.md) complement the "stock" Kubernetes functionality
by adding features such as load balancing, persistent storage and monitoring to a cluster. To keep
this guide short you will deploy a single component - `httpbin` - which serves as a demo
application to verify the cluster behaves as expected.

## Requirements

* [libvirt] installed locally and running.
* Host machine with at least 25 GB of free disk space available locally.
* Host machine with at least 4 CPU threads.
* Host machine with at least 16 GB of RAM.
* An SSH key pair for accessing the cluster nodes.
* Terraform `v0.13.x`
  [installed](https://learn.hashicorp.com/terraform/getting-started/install.html#install-terraform).
* The `libvirt` Terraform provider `v0.6.2`
  [installed](https://github.com/dmacvicar/terraform-provider-libvirt#installing).
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

export release=$(curl -s https://api.github.com/repos/kinvolk/lokomotive/releases | jq -r '.[0].name')
curl -LO "https://github.com/kinvolk/lokomotive/releases/download/${release}/lokoctl_${release}_${os}_amd64.tar.gz"
```

Extract the binary and copy it to a place under your `$PATH`:

```console
tar zxvf lokoctl_${release}_${os}_amd64.tar.gz
sudo cp lokoctl_${release}_${os}_amd64/lokoctl /usr/local/bin
rm -rf lokoctl_${release}_${os}_amd64*
```

### Step 2: Download Flatcar Container Linux image

Before we deploy the cluster, we need to download a Flatcar Container Linux image, which will be used as a
base OS image for the Tinkerbell provisioner server.

```bash
sudo dnf install wget bzip2 # or sudo apt install wget bzip2
wget https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_qemu_image.img.bz2
bunzip2 flatcar_production_qemu_image.img.bz2
```

After you download the image, get absolute path of it using the following command:

```bash
realpath flatcar_production_qemu_image.img
```

It will be needed for next step.

### Step 3: Create a cluster configuration

Create a directory for the cluster-related files and navigate to it:

```console
mkdir lokomotive-demo && cd lokomotive-demo
```

Create a file named `cluster.lokocfg` with the following contents:

```hcl
cluster "tinkerbell" {
  asset_dir               = "./assets"
  name                    = "demo"
  dns_zone                = "example.com"
  ssh_public_keys         = ["ssh-rsa AAAA..."]
  controller_ip_addresses = ["10.17.3.4"]

  //os_channel       = "stable"
  //os_version       = "current"

  experimental_sandbox {
    pool_path          = "/opt/pool"
    flatcar_image_path = "/opt/flatcar_production_qemu_image.img"
    hosts_cidr         = "10.17.3.0/24"
  }

  worker_pool "pool1" {
    ip_addresses    = ["10.17.3.5"]

    //os_channel    = "stable"
    //os_version    = "current"
  }
}

# A demo application.
component "httpbin" {
  ingress_host = "httpbin.example.com"
}
```

Replace the parameters above using the following information:

- `ssh_public_keys` - A list of strings representing the *contents* of the public SSH keys which should
  be authorized on cluster nodes.
- `pool_path` - Absolute path where VM disk images will be stored. Can be set output of `echo $(pwd)/pool` command.
- `flatcar_image_path` - This should be set to the value obtained from previous step.

The rest of the parameters may be left as-is. For more information about the configuration options
see the [configuration reference](../configuration-reference/platforms/tinkerbell.md).

### Step 4: Deploy the cluster

Add a private key corresponding to one of the public keys specified in `ssh_public_keys` to your `ssh-agent`:

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

If some provisioning error occurs, the best strategy is to destroy resources which has been created
and start over. See [Cleanup](#cleanup) section for instruction how to clean things up.

### Stuck at "copy controller secrets"

```
...
module.equinixmetal-lokomotive-demo.null_resource.copy-controller-secrets: Still creating... (8m30s elapsed)
module.equinixmetal-lokomotive-demo.null_resource.copy-controller-secrets: Still creating... (8m40s elapsed)
...
```

In case the deployment process seems to hang at the `copy-controller-secrets` phase for a long
time (more than 15 minutes), check the following:

- Verify the correct private SSH key was added to `ssh-agent`.
- Verify that you can SSH into the created controller node from the machine running `lokoctl`.

### Cluster node provisioning failed

If cluster machine did not start with Flatcar, you need to investigate why provisioning failed. This can be
done by opening SSH session to provisioner machine and then accessing `tink` CLI using the following commands:

```sh
ssh root@10.17.3.2
source tink/.env && docker-compose -f tink/deploy/docker-compose.yml exec tink-cli sh
```

#### Checking hardware entries

With `tink` CLI available, you can check if hardware has been registered properly:

```sh
tink hardware list
```

It should return output similar to this:

```console
+--------------------------------------+-------------------+------------+----------+
| ID                                   | MAC ADDRESS       | IP ADDRESS | HOSTNAME |
+--------------------------------------+-------------------+------------+----------+
| 829af385-7d96-e35a-f9d4-8c6495ccd2f3 | 52:13:a0:c9:24:87 | 10.17.3.4  |          |
| 9f38e568-e5c4-6c2a-0859-6ea22b2650bb | 52:3b:0a:10:13:b8 | 10.17.3.5  |          |
+--------------------------------------+-------------------+------------+----------+
```

You should find here the IP addresses which are specified in the configuration. If the IP addresses are missing,
created workflows which installs cluster nodes did not run as they didn't find the machine to run on.

#### Checking workflows

For each cluster node, a Tinkerbell workflow should be created, which installs the OS on the node disk. You can see created
workflows using the following command:

```sh
tink workflow list
```

Sample output:

```console
+--------------------------------------+--------------------------------------+---------------------------+-------------------------------+-------------------------------+
| WORKFLOW ID                          | TEMPLATE ID                          | HARDWARE DEVICE           | CREATED AT                    | UPDATED AT                    |
+--------------------------------------+--------------------------------------+---------------------------+-------------------------------+-------------------------------+
| dc9669e2-b57c-4e08-9655-6d8d2a1d3791 | 42bb7fd3-240a-4df5-a46e-dd8820f6c231 | {"device_1": "10.17.3.4"} | 2020-10-14 10:17:58 +0000 UTC | 2020-10-14 10:17:58 +0000 UTC |
| b58bd418-1586-48fc-b640-09c43c1b83d7 | 2e83e46d-edfa-41e7-b4fa-4a1741031a8c | {"device_1": "10.17.3.5"} | 2020-10-14 10:17:58 +0000 UTC | 2020-10-14 10:17:58 +0000 UTC |
+--------------------------------------+--------------------------------------+---------------------------+-------------------------------+-------------------------------
```

You can check the state of each individual workflow using the following sample commands:

```sh
tink workflow state dc9669e2-b57c-4e08-9655-6d8d2a1d3791
tink workflow events dc9669e2-b57c-4e08-9655-6d8d2a1d3791
```

Sample output:

```console
/ # tink workflow state dc9669e2-b57c-4e08-9655-6d8d2a1d3791
+----------------------+--------------------------------------+
| FIELD NAME           | VALUES                               |
+----------------------+--------------------------------------+
| Workflow ID          | dc9669e2-b57c-4e08-9655-6d8d2a1d3791 |
| Workflow Progress    | 66%                                  |
| Current Task         | flatcar-install                      |
| Current Action       | reboot                               |
| Current Worker       | 829af385-7d96-e35a-f9d4-8c6495ccd2f3 |
| Current Action State | ACTION_IN_PROGRESS                   |
+----------------------+--------------------------------------+
/ # tink workflow events dc9669e2-b57c-4e08-9655-6d8d2a1d3791
+--------------------------------------+-----------------+-----------------+----------------+---------------------------------+--------------------+
| WORKER ID                            | TASK NAME       | ACTION NAME     | EXECUTION TIME | MESSAGE                         |      ACTION STATUS |
+--------------------------------------+-----------------+-----------------+----------------+---------------------------------+--------------------+
| 829af385-7d96-e35a-f9d4-8c6495ccd2f3 | flatcar-install | dump-ignition   |              0 | Started execution               | ACTION_IN_PROGRESS |
| 829af385-7d96-e35a-f9d4-8c6495ccd2f3 | flatcar-install | dump-ignition   |              4 | Finished Execution Successfully |     ACTION_SUCCESS |
| 829af385-7d96-e35a-f9d4-8c6495ccd2f3 | flatcar-install | flatcar-install |              0 | Started execution               | ACTION_IN_PROGRESS |
| 829af385-7d96-e35a-f9d4-8c6495ccd2f3 | flatcar-install | flatcar-install |             98 | Finished Execution Successfully |     ACTION_SUCCESS |
| 829af385-7d96-e35a-f9d4-8c6495ccd2f3 | flatcar-install | reboot          |              0 | Started execution               | ACTION_IN_PROGRESS |
+--------------------------------------+-----------------+-----------------+----------------+---------------------------------+--------------------+
```

If everything went well, you should see the `ACTION_IN_PROGRESS` state on the `reboot` action as in the example above. This is expected, as this action
reboots the node, so it has no time to send information about successful execution back to Tinkerbell server.

### Error opening Flatcar image file

```
Error: Error while determining image type for /root/assets/terraform/flatcar_production_qemu_image.img: Error while opening /root/assets/terraform/flatcar_production_qemu_image.img: open /root/assets/terraform/flatcar_production_qemu_image.img: no such file or directory
```

If you see an error similar to this, it means that the `flatcar_image_path` parameter is not set correctly.
Please make sure that its value points to a valid Flatcar image.


### Cannot access storage pool

```
Error: Error creating libvirt domain: virError(Code=38, Domain=18, Message='Cannot access storage file '/root/pool/provisioner' (as uid:64055, gid:108): Permission denied')
```

Libvirt is not able to use some paths as storage pools. If an error similar to this occurs, try
changing the path in the `pool_path` attribute to some other value or make sure that the UID and GID
mentioned in the error message has permissions to write to the configured path.

### Domain already exist with UUID ...

```
Error: Error defining libvirt domain: virError(Code=9, Domain=20, Message='operation failed: domain 'tinkerbell-sandbox-demo-provisioner' already exists with uuid 55ce0fc5-a961-4625-8e45-71dcc9bd4f43')
```

If cluster creation failed for you previously, you may see this
error message on consecutive runs.

This error is caused by the libvirt Terraform provider not being able to
properly clean up machines which has failed to create.

If this occurs, run the following command to see which machines are created:

```sh
virsh list --all
```

Then, remove the machine using example command:

```sh
virsh undefine <machine name>
```

[releases]: https://github.com/kinvolk/lokomotive/releases
[Tinkerbell]: https://tinkerbell.org/
[libvirt]: https://libvirt.org/
