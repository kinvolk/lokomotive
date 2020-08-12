# Lokomotive KVM libvirt quickstart guide

## Contents

* [Introduction](#introduction)
* [Requirements](#requirements)
* [Step 1: Install lokoctl](#step-1-install-lokoctl)
* [Step 2: Prepare the VM Image](#step-2-prepare-the-vm-image)
* [Step 3: libvirtd Setup](#step-3-libvirtd-setup)
* [Step 4: User Setup](#step-4-user-setup)
* [Step 5: Create a cluster configuration](#step-5-create-a-cluster-configuration)
* [Step 6: Check ssh-agent](#step-6-check-ssh-agent)
* [Step 7: Deploy the cluster](#step-7-deploy-the-cluster)
* [Verification](#verification)
* [Using the cluster](#using-the-cluster)
* [Cleanup](#cleanup)
* [Troubleshooting](#troubleshooting)

## Introduction

This guide shows how to create a Lokomotive cluster with Flatcar Container Linux VMs on your local machine.
By the end of this guide, you'll have a basic Lokomotive cluster running that consist of controller and worker nodes.

Controllers are provisioned to run an `etcd-member` peer and a `kubelet` service.
Workers run just a `kubelet` service. A one-time [bootkube](https://github.com/kubernetes-incubator/bootkube)
bootstrap process schedules the `apiserver`, `scheduler`, `controller-manager`, and `coredns` pods on the controllers
and schedules `kube-proxy` and `calico` on every node.
A generated `kubeconfig` provides `kubectl` access to the cluster.

## Requirements

* A Linux host OS.
* At least 4 GB of available RAM.
* An SSH key pair.
* Terraform `v0.12.x`
  [installed](https://learn.hashicorp.com/terraform/getting-started/install.html#install-terraform).
* [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct),
  and [terraform-provider-libvirt](https://github.com/dmacvicar/terraform-provider-libvirt)
  installed locally, e.g., as two binaries `~/.terraform.d/plugins/terraform-provider-ct_v0.5.0`
  and `~/.terraform.d/plugins/terraform-provider-libvirt_v0.6.1`.
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed.


## Steps

### Step 1: Install lokoctl


`lokoctl` is the command-line interface for managing Lokomotive clusters.

Download the latest `lokoctl` binary:

```console
export release=$(curl -s https://api.github.com/repos/kinvolk/lokomotive/releases | jq -r '.[0].name' | tr -d v)
curl -LO "https://github.com/kinvolk/lokomotive/releases/download/v${release}/lokoctl_${release}_linux_amd64.tar.gz"
```

Extract the binary and copy it to a place under your `$PATH`:

```console
tar zxvf lokoctl_${release}_linux_amd64.tar.gz
cp lokoctl_${release}_linux_amd64/lokoctl ~/.local/bin/
rm -rf lokoctl_${release}_linux_amd64*
```

### Step 2: Prepare the VM Image

Download a Flatcar Container Linux image. The example below uses the Alpha Flatcar Container Linux channel.
You need to increase its size depending on your use case because the initial size
is only enough for some small pods to be loaded. Starting pods with normal-sized
container images may already fail when the disk space is filled.
That's why the example below directly increases the image size by 5 GB to make
sure that a couple of container images fit on the node.

```console
# On Fedora/RHEL/CentOS
sudo dnf install wget bzip2 qemu-img
# On Ubuntu/Debian
sudo apt install wget bzip2 qemu-utils
wget https://alpha.release.flatcar-linux.net/amd64-usr/current/flatcar_production_qemu_image.img.bz2
bunzip2 flatcar_production_qemu_image.img.bz2
qemu-img resize flatcar_production_qemu_image.img +5G
```

Do not use this image file with QEMU directly because it needs to stay clean for Ignition to run on first boot.

### Step 3: libvirtd Setup

Ensure that libvirtd is set up.

```console
# On Fedora/RHEL/CentOS
sudo dnf install libvirt-daemon
# On Ubuntu/Debian
sudo apt install libvirt-daemon
sudo systemctl start libvirtd
```

#### AppArmor and libvirt issues

If you use Ubuntu or Debian with AppArmor, you need to disable AppArmor because it disallows
using temporary paths for the storage pool. Open `/etc/libvirt/qemu.conf` and add the
following line below any existing commented `#security_driver = "selinux"` line:

```
security_driver = "none"
```

Restart libvirtd using `sudo systemctl restart libvirtd`.

### Step 4: User Setup

Ensure that libvirt is accessible for the current user because the user needs to create
system-wide virtual machines with libvirt:

```console
groups
```

If you see some group names like `myuser wheel postgres docker` but not `libvirt` you need to
add the current user to the libvirt group.

```console
sudo usermod -a -G libvirt $(whoami)
newgrp libvirt
```

If `newgrp` was used you have to continue to use this terminal session or run `newgrp libvirt`
in every new terminal sessons or log your user out and in again.

### Step 5: Create a cluster configuration

Create a directory for the cluster-related files and navigate to it:

```console
mkdir lokomotive-demo && cd lokomotive-demo
```

Create a file called `cluster.lokocfg` with these contents but remember to change and
check the image path and check the SSH public key path:

```hcl
cluster "kvm-libvirt" {
  asset_dir = pathexpand("./assets")
  # Your SSH public key
  ssh_pubkeys = [file(pathexpand("~/.ssh/id_rsa.pub"))]
  cluster_name = "vmcluster"
  machine_domain = "vmcluster.k8s"
  # Path to the prepared image file. Note the absolute path notation with the three slashes.
  os_image = "file:///var/tmp/lokomotive-demo/flatcar_production_qemu_image.img"

  worker_pool "one" {
    count = 1
  }
}
```

Again, check the contents of `cluster.lokocfg`.
Check if you need to change the SSH key entry to match your `~/.ssh/id_*.pub` and
change the `os_image` path to point to the image location from the VM image preparation.

The rest of the parameters may be left as-is. For more information about the configuration options
see the [configuration reference](../configuration-reference/platforms/kvm-libvirt.md).


### Step 6: Check ssh-agent

Check that your SSH key is known to `ssh-agent` so that it is unlocked if it was secured with a passphrase:

```console
ssh-add -L
```

This should normally be the case when the key was created with the GNOME _Passwords and Keys_ tool.
Otherwise add the key:

```console
ssh-add ~/.ssh/id_rsa
```

### Step 7: Deploy the cluster

Deploy the cluster with `lokoctl`:

```console
lokoctl cluster apply
INFO[0000] Initializing Terraform working directory      phase=infrastructure

You can find the logs in "assets/terraform/logs/3213353.log"
INFO[0001] Applying Terraform configuration. This creates infrastructure so it might take a long time...  phase=infrastructure

You can find the logs in "assets/terraform/logs/3213406.log"

Your configurations are stored in ./assets

Now checking health and readiness of the cluster nodes ...

Node                                    Ready    Reason          Message

vmcluster-controller-0.vmcluster.k8s    True     KubeletReady    kubelet is posting ready status
vmcluster-one-worker-0.vmcluster.k8s    True     KubeletReady    kubelet is posting ready status

Success - cluster is healthy and nodes are ready!
INFO[0498] Applying component configuration              args="[]" command="lokoctl cluster apply"
```

The deployment process typically takes 3-6 minutes, the Lokomotive cluster is be ready, depending on your internet connection and hardware.

## Verification

Use the generated `kubeconfig` credentials to access the Kubernetes cluster and list nodes:

```console
export KUBECONFIG=$(pwd)/assets/cluster-assets/auth/kubeconfig
kubectl get nodes
NAME                                   STATUS   ROLES               AGE     VERSION
vmcluster-controller-0.vmcluster.k8s   Ready    controller,master   5h28m   v1.14.1
vmcluster-one-worker-0.vmcluster.k8s   Ready    node                5h28m   v1.14.1
```

List the pods.

```console
kubectl get pods --all-namespaces
kube-system   calico-node-x5cgr                          1/1     Running   0          5h28m
kube-system   calico-node-zp2bl                          1/1     Running   0          5h28m
kube-system   coredns-5644c585c9-58w2r                   1/1     Running   2          5h28m
kube-system   coredns-5644c585c9-vtknk                   1/1     Running   1          5h28m
kube-system   kube-apiserver-8fxjs                       1/1     Running   3          5h28m
kube-system   kube-controller-manager-865f9f995d-cj8hp   1/1     Running   1          5h28m
kube-system   kube-controller-manager-865f9f995d-jtmqf   1/1     Running   1          5h28m
kube-system   kube-proxy-jxsz4                           1/1     Running   0          5h28m
kube-system   kube-proxy-r24jn                           1/1     Running   0          5h28m
kube-system   kube-scheduler-57c444dcc8-54zpq            1/1     Running   1          5h28m
kube-system   kube-scheduler-57c444dcc8-6gv94            1/1     Running   0          5h28m
kube-system   pod-checkpointer-8qdrt                     1/1     Running   0          5h28m
```


## Using the cluster

At this point you should have access to a Lokomotive cluster and can use it to deploy applications.

If you don't have any Kubernetes experience, you can check out the [Kubernetes
Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/deploy-app/deploy-intro/) tutorial.

When you don't need the cluster for some time, shut the VMs down in `virt-manager` to free up RAM.

>NOTE: Lokomotive uses a relatively restrictive Pod Security Policy by default. This policy
>disallows running containers as root. Refer to the
>[Pod Security Policy documentation](../concepts/securing-lokomotive-cluster.md#cluster-wide-pod-security-policy)
>for more details.

## Cleanup

To destroy the cluster, execute the following command:

```console
lokoctl cluster destroy
```

Confirm you want to destroy the cluster by typing `yes` and hitting **Enter**.

You can now safely delete the directory created for this guide if you no longer need it.

## Troubleshooting

If you want to log in using SSH for debugging purposes, use the Flatcar Container Linux default user `core`.

You can look up the node IP address in `virt-manager` or resolve them from the hostnames,
using the libvirt DNS server:

```console
host vmcluster-controller-0.vmcluster.k8s 192.168.192.1
Using domain server:
Name: 192.168.192.1
Address: 192.168.192.1#53
Aliases: 

vmcluster-controller-0.vmcluster.k8s has address 192.168.192.10
```

On each node list the containers with `docker ps` and follow the system logs with `journalctl -f`.

