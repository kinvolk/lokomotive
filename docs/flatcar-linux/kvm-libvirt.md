# KVM libvirt

In this tutorial, we'll create a Kubernetes cluster with Flatcar Container Linux inside VMs on our local machine.

We'll declare the cluster using the Lokomotive Terraform module for the KVM libvirt platform.

Controllers are provisioned to run an `etcd-member` peer and a `kubelet` service.
Workers run just a `kubelet` service. A one-time [bootkube](https://github.com/kubernetes-incubator/bootkube)
bootstrap schedules the `apiserver`, `scheduler`, `controller-manager`, and `coredns` on controllers
and schedules `kube-proxy` and `calico` (or `flannel`) on every node.
A generated `kubeconfig` provides `kubectl` access to the cluster.

## Requirements

* `qemu-img`, `wget`, and `bunzip2` binaries installed
* At least 4 GB of free RAM
* Running libvirtd system service, see [libvirtd Setup](#libvirtd-setup)
* A user in the `libvirt` group (to create virtual machines with libvirt and access `/dev/kvm`), see [User Setup]
* Terraform v0.11.x, [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct),
  and [terraform-provider-libvirt](https://github.com/dmacvicar/terraform-provider-libvirt)
  installed locally

## Prepare VM Image

Download the Flatcar Container Linux image. The example below uses the "Edge" Flatcar Container Linux channel.
You need to increase its size depending on your use case because the initial size
is only enough for some small pods to be loaded. Starting pods with normal-sized
container images may already fail because the disk space will be filled.
That's why the example below directly increases the image size by 5 GB to make
sure that a couple of container images fit on the node.

```sh
$ sudo dnf install wget bzip2 qemu-img  # or sudo apt install wget bzip2 qemu-utils
$ wget https://edge.release.flatcar-linux.net/amd64-usr/current/flatcar_production_qemu_image.img.bz2
$ bunzip2 flatcar_production_qemu_image.img.bz2
$ qemu-img resize flatcar_production_qemu_image.img +5G
```

## libvirtd Setup

Ensure that libvirtd is set up.
The examples in this guide use Fedora's `dnf`
but for example in Ubuntu you would use `apt` as stated in the comment in the same line.

```sh
$ sudo dnf install libvirt-daemon  # or sudo apt install libvirt-daemon
$ sudo systemctl start libvirtd
```

### Apparmor and libvirt issues

If you use Ubuntu or Debian with apparmor, you need to disable apparmor because it disallows
to use temporary paths for the storage pool. Open `/etc/libvirt/qemu.conf` and add the
following line below any existing commented `#security_driver = "selinux"` line.

```
security_driver = "none"
```

Now restart the libvirtd through `sudo systemctl restart libvirtd`.

## User Setup

Ensure that libvirt is accessible for the current user:

```sh
$ sudo usermod -a -G libvirt $(whoami)
$ newgrp libvirt
```

## Terraform Setup

Install [Terraform](https://www.terraform.io/downloads.html) v0.11.x on your system.

```sh
$ terraform version
Terraform v0.11.13
```

Add the [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct) plugin binary for your system
to `~/.terraform.d/plugins/`, noting the `_v0.3.1` suffix.

```sh
wget https://github.com/poseidon/terraform-provider-ct/releases/download/v0.4.0/terraform-provider-ct-v0.4.0-linux-amd64.tar.gz
tar xzf terraform-provider-ct-v0.4.0-linux-amd64.tar.gz
mv terraform-provider-ct-v0.4.0-linux-amd64/terraform-provider-ct ~/.terraform.d/plugins/terraform-provider-ct_v0.4.0
```

Download the tar file for your distribution from the [release page](https://github.com/dmacvicar/terraform-provider-libvirt/releases):

```sh
wget https://github.com/dmacvicar/terraform-provider-libvirt/releases/download/v0.6.0/terraform-provider-libvirt-0.6.0+git.1569597268.1c8597df.Fedora_28.x86_64.tar.gz
# or, e.g., https://github.com/dmacvicar/terraform-provider-libvirt/releases/download/v0.6.0/terraform-provider-libvirt-0.6.0+git.1569597268.1c8597df.Ubuntu_18.04.amd64.tar.gz
tar xzf terraform-provider-libvirt-0.6.0+git.1569597268.1c8597df.Fedora_28.x86_64.tar.gz
mv terraform-provider-libvirt ~/.terraform.d/plugins/terraform-provider-libvirt_v0.6.0
```

Read [concepts](/docs/architecture/concepts.md) to learn about Terraform, modules, and organizing resources if the following confuses you.

A Terraform project should be created in a new clean directory.
Assuming there is a parent directory `infra` for all the terraform projects, we create a new subdirectory `vmcluster`.

```sh
mkdir vmcluster
cd vmcluster
```

## Provider

Create a file named `provider.tf` with the following contents:

```tf
provider "ct" {
  version = "0.4.0"
}

provider "local" {
  version = "~> 1.2"
  alias   = "default"
}

provider "null" {
  version = "~> 2.1"
  alias   = "default"
}

provider "template" {
  version = "~> 2.1"
  alias   = "default"
}

provider "tls" {
  version = "~> 2.0"
  alias   = "default"
}

provider "libvirt" {
  version = "~> 0.6.0"
  uri     = "qemu:///system"
  alias   = "default"
}
```

## Cluster

Follow the below steps to define a Kubernetes cluster using the module
[kvm-libvirt/flatcar-linux/kubernetes](https://github.com/kinvolk/lokomotive-kubernetes/tree/master/kvm-libvirt/flatcar-linux/kubernetes).

Create a file called `mycluster.tf` with these contents but remember to change the `path/to` and `AAAA...` values:

```tf
module "controller" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//kvm-libvirt/flatcar-linux/kubernetes"

  providers = {
    local    = "local.default"
    null     = "null.default"
    template = "template.default"
    tls      = "tls.default"
    libvirt  = "libvirt.default"
  }

  # Path to where the image was prepared, note the triple slash for the absolute path
  os_image_unpacked = "file:///path/to/flatcar_production_qemu_image.img"

  # Your SSH public key
  ssh_keys = [
    "ssh-rsa AAAA...",
  ]

  # Where you want the terraform assets to be saved for KUBECONFIG
  asset_dir = "/path/to/clusters/asset"

  machine_domain = "vmcluster.k8s"
  cluster_name = "vmcluster"
  node_ip_pool = "192.168.192.0/24"

  controller_count = 1

}

module "worker-pool-one" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//kvm-libvirt/flatcar-linux/kubernetes/workers"

  providers = {
    local    = "local.default"
    template = "template.default"
    tls      = "tls.default"
    libvirt  = "libvirt.default"
  }

  ssh_keys = "${module.controller.ssh_keys}"

  machine_domain = "${module.controller.machine_domain}"
  cluster_name = "${module.controller.cluster_name}"
  libvirtpool = "${module.controller.libvirtpool}"
  libvirtbaseid = "${module.controller.libvirtbaseid}"

  pool_name = "one"

  count = 1

  kubeconfig = "${module.controller.kubeconfig}"

  labels = "node.supernova.io/role=backend"
}
```

Now check the contents of `mycluster.tf`.
Change at least the `ssh_keys` entry to match the content of your `~/.ssh/id_rsa.pub` and
change the `os_image_unpacked` path to point to the image location from the VM image preparation.

Reference the [variables docs](#variables) or the
[variables.tf](https://github.com/kinvolk/lokomotive-kubernetes/blob/master/kvm-libvirt/flatcar-linux/kubernetes/variables.tf)
source.

## ssh-agent

Initial bootstrapping requires `bootkube.service` to be started on one controller node.
Terraform uses `ssh-agent` to automate this step.
Add your SSH private key to `ssh-agent` so that it is unlocked if it was secured with a passphrase
(you can skip this if the key was created with the GNOME Passwords and Keys tool).

```sh
exec /usr/bin/ssh-agent $SHELL
ssh-add ~/.ssh/id_rsa
ssh-add -L
```

## Apply

Initialize the config directory since this is the first use with Terraform.

```sh
terraform init
```

Plan the resources to be created.

```sh
$ terraform plan
Plan: 68 to add, 0 to change, 0 to destroy.
```

Apply the changes to create the cluster (type `yes` when prompted).

```sh
$ terraform apply
...
module.controller.null_resource.bootkube-start: Still creating... (2m50s elapsed)
module.controller.null_resource.bootkube-start: Still creating... (3m0s elapsed)
module.controller.null_resource.bootkube-start: Creation complete after 3m20s

Apply complete! Resources: 68 added, 0 changed, 0 destroyed.
```

In 3-6 minutes, the Kubernetes cluster will be ready, depending on your downlink speed and system performance.

## Verify that it works

[Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) on your system.
Use the generated `kubeconfig` credentials to access the Kubernetes cluster and list nodes.

```
$ export KUBECONFIG=/path/to/clusters/asset/auth/kubeconfig  # matching the asset_dir from the config file
$ kubectl get nodes
NAME                                   STATUS   ROLES               AGE     VERSION
vmcluster-controller-0.vmcluster.k8s   Ready    controller,master   5h28m   v1.14.1
vmcluster-one-worker-0.vmcluster.k8s   Ready    node                5h28m   v1.14.1
```

List the pods.

```
$ kubectl get pods --all-namespaces
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

## Variables

Check the
[variables.tf](https://github.com/kinvolk/lokomotive-kubernetes/blob/master/kvm-libvirt/flatcar-linux/kubernetes/variables.tf)
source.

### Controller

#### Required

| Name | Description | Example |
|:-----|:------------|:--------|
| asset_dir | Path to a directory where generated assets should be placed (contains secrets) | "/home/user/infra/assets" |
| cluster_name | Unique cluster name | "vmcluster" |
| machine_domain | DNS zone | "vmcluster.k8s" |
| os_image_unpacked | Path to unpacked Flatcar Container Linux image (probably after a qemu-img resize IMG +5G) | "file:///home/user/infra/flatcar_production_qemu_image.img" |
| ssh_keys | List of SSH public keys for user 'core' | ["ssh-rsa AAAAB3NZ..."] |

#### Optional
| Name | Description | Default | Example |
|:-----|:------------|:--------|:--------|
| cluster_domain_suffix | Queries for domains with the suffix will be answered by coredns | "cluster.local" | "k8s.example.com" |
| controller_count | Number of controller VMs | "1" | "1" |
| enable_aggregation | Enable the Kubernetes Aggregation Layer | "false" | "true" |
| enable_reporting | Enable usage or analytics reporting to upstreams (Calico) | "false" | "true" |
| network_mtu | CNI interface MTU (applies to calico only) | "1480" | "8981" |
| network_ip_autodetection_method | Method to autodetect the host IPv4 address (applies to calico only) | "first-found" or "can-reach=192.168.192.1" |
| networking | Choice of networking provider | "calico" | "calico" or "flannel" |
| node_ip_pool | Unique VM IP CIDR (different per cluster) | "192.168.192.0/24" | | "192.168.13.0/24" |
| pod_cidr | CIDR IPv4 range to assign Kubernetes pods | "10.1.0.0/16" | "10.22.0.0/16" |
| service_cidr | CIDR IPv4 range to assign Kubernetes services. The 1st IP will be reserved for kube_apiserver, the 10th IP will be reserved for coredns. | "10.2.0.0/16" | "10.3.0.0/24" |
| virtual_cpus | Number of virtual CPUs | "1" | "2" |
| virtual_memory | Virtual RAM in MB | "2048" | "4096" |
| certs_validity_period_hours | Validity of all the certificates in hours | "8760" | "17520" |
| controller_clc_snippets | Controller Container Linux Config snippets | [] | [example](../advanced/customization.md#usage) |

### Worker

#### Required

| Name | Description | Example |
|:-----|:------------|:--------|
| cluster_name | Cluster to join | "${module.controller.cluster_name}" |
| kubeconfig | Kubeconfig file generated for the controller | "${module.controller.kubeconfig}" |
| libvirtbaseid | The cluster's OS volume to use for copy-on-write | "${module.controller.libvirtbaseid}" |
| libvirtpool | The cluster's disk volume pool to use | "${module.controller.libvirtpool}" |
| machine_domain | DNS zone | "${module.controller.machine_domain}" |
| os_image_unpacked | Path to unpacked Flatcar Container Linux image (probably after a qemu-img resize IMG +5G) | "file:///home/user/infra/flatcar_production_qemu_image.img" |
| pool_name | Worker pool name | "one" |
| ssh_keys | List of SSH public keys for user 'core' | "${module.controller.ssh_keys}" |

#### Optional
| Name | Description | Default | Example |
|:-----|:------------|:--------|:--------|
| count | Number of worker VMs | "1" | "3" |
| cluster_domain_suffix | The cluster's suffix answered by coredns | "cluster.local" | "k8s.example.com" |
| labels | Custom label to assign to worker nodes. Provide comma separated key=value pairs as labels." | "" | "foo=oof,bar=,baz=zab" |
| service_cidr | CIDR IPv4 range to assign Kubernetes services. The 1st IP will be reserved for kube_apiserver, the 10th IP will be reserved for coredns. | "10.2.0.0/16" | "10.3.0.0/24" |
| virtual_cpus | Number of virtual CPUs | "1" | "2" |
| virtual_memory | Virtual RAM in MB | "2048" | "4096" |
| clc_snippets | Worker Container Linux Config snippets | [] | [example](../advanced/customization.md#usage) |

#### Closing Notes

The generated kubeconfig contains the node's public IP. If you want to use the names,
you have to resolve them over the libvirt DNS server `192.168.192.1`.

If you need a reproducible setup, please add ?ref=HASH to the end of
each `git::` path in the `source = "git::…"` imports.

You can consider to change the Terraform modules provided by Lokomotive
if they don't meet your demands
(change the `source = "git:…"` import to your fork or a relative path with `source = "./libvirt-kubernetes"`).
Additional configuration options you may specify there for libvirt are described in the `libvirt` provider
[docs](https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/website/docs/index.html.markdown).
