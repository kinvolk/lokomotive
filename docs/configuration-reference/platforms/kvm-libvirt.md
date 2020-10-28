# Lokomotive KVM libvirt configuration reference

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Destroying](#destroying)

## Introduction

This configuration reference provides information on configuring a Lokomotive cluster on KVM libvirt VMs
with all the configuration options available to the user.

## Prerequisites

* Terraform providers and libvirt setup from the [quickstart guide](../../quickstarts/kvm-libvirt.md)
* `lokoctl` [installed locally.](../../installer/lokoctl.md)
* `kubectl` installed locally to access the Kubernetes cluster.

### Configuration

To create a Lokomotive cluster, we need to define a configuration.

Example configuration file:

```tf

cluster "kvm-libvirt" {

  asset_dir = pathexpand("./assets")

  cluster_name = "vmcluster"

  machine_domain = "vmcluster.k8s"

  os_image = "file:///home/myuser/Downloads/flatcar_production_qemu_image.img"

  ssh_pubkeys = [file(pathexpand("~/.ssh/id_rsa.pub"))]

  controller_count = 1

  node_ip_pool = "192.168.192.0/24"

  disable_self_hosted_kubelet = false

  kube_apiserver_extra_flags = []

  controller_virtual_cpus = 1
  controller_virtual_memory = 2048

  controller_clc_snippets = []

  network_mtu = 1480
  network_ip_autodetection_method = "first-found"

  pod_cidr = "10.1.0.0/16"
  service_cidr = "10.2.0.0/16"

  cluster_domain_suffix = "cluster.local"

  enable_reporting = false
  enable_aggregation = true

  certs_validity_period_hours = 8760

  enable_tls_bootstrap = true

  worker_pool "worker-pool-1" {
    count = 2

    virtual_cpus = 1
    virtual_memory = 2048

    labels = "foo=oof,bar=,baz=zab"

    clc_snippets = [
  <<EOF
systemd:
  units:
  - name: helloworld.service
    dropins:
      - name: 10-helloworld.conf
        contents: |
          [Install]
          WantedBy=multi-user.target
EOF
        ,
  ]

  }
}
```


## Attribute reference

| Argument                              | Description                                                                                                                                                                                                                                                                       |       Default      |     Type     | Required |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:------------------:|:------------:|:--------:|
| `asset_dir`                           | Location where Lokomotive stores cluster assets.                                                                                                                                                                                                                                  |          -         |    string    |   true   |
| `certs_validity_period_hours`         | Validity of all the certificates in hours.                                                                                                                                                                                                                                        |        8760        |    number    |  false   |
| `cluster_domain_suffix`               | Cluster's DNS domain.                                                                                                                                                                                                                                                             |   "cluster.local"  |    string    |  false   |
| `cluster_name`                        | Name of the cluster.                                                                                                                                                                                                                                                              |          -         |    string    |   true   |
| `controller_clc_snippets`             | Controller Flatcar Container Linux Config snippets.                                                                                                                                                                                                                               |          []        | list(string) |  false   |
| `controller_count`                    | Number of controller nodes.                                                                                                                                                                                                                                                       |          1         |    number    |  false   |
| `controller_virtual_cpus`             | Number of virtual CPUs for the controller VMs.                                                                                                                                                                                                                                    |          1         |      int     |  false   |
| `controller_virtual_memory`           | Virtual RAM in MB for the controller VMs.                                                                                                                                                                                                                                         |         2048       |      int     |  false   |
| `disable_self_hosted_kubelet`         | Disable self-hosting the kubelet as pod on the cluster.                                                                                                                                                                                                                           |        false       |     bool     |  false   |
| `enable_aggregation`                  | Enable the Kubernetes Aggregation Layer.                                                                                                                                                                                                                                          |        true        |     bool     |  false   |
| `enable_reporting`                    | Enables usage or analytics reporting to upstream.                                                                                                                                                                                                                                 |        false       |     bool     |  false   |
| `kube_apiserver_extra_flags`          | Extra flags to pass to the kube-apiserver binary.                                                                                                                                                                                                                                 |          []        | list(string) |  false   |
| `machine_domain`                      | DNS zone of the cluster, used by nodes to find each other as HOSTNAME.machine_domain.                                                                                                                                                                                             |          -         |    string    |   true   |
| `network_mtu`                         | CNI interface MTU                                                                                                                                                                                                                                                                 |        1480        |    number    |  false   |
| `node_ip_pool`                        | Unique VM IP CIDR.                                                                                                                                                                                                                                                                | "192.168.192.0/24" |    string    |  false   |
| `os_image`                            | Path to unpacked Flatcar Container Linux image flatcar_production_qemu_image.img (probably after a qemu-img resize IMG +5G).                                                                                                                                                      |          -         |    string    |   true   |
| `pod_cidr`                            | CIDR IPv4 range to assign Kubernetes pods.                                                                                                                                                                                                                                        |    "10.2.0.0/16"   |    string    |  false   |
| `service_cidr`                        | CIDR IPv4 range to assign Kubernetes services.                                                                                                                                                                                                                                    |    "10.3.0.0/16"   |    string    |  false   |
| `ssh_pubkeys`                         | List of SSH public keys for user `core`. Each element must be specified in a valid OpenSSH public key format, as defined in RFC 4253 Section 6.6, e.g. "ssh-rsa AAAAB3N...".                                                                                                      |          -         | list(string) |   true   |
| `enable_tls_bootstrap`                | Enable TLS bootstrapping for Kubelet.                                                                                                                                                                                                                                             |        true        |     bool     |  false   |
| `worker_pool.clc_snippets`            | Flatcar Container Linux Config snippets for nodes in the worker pool.                                                                                                                                                                                                             |          []        | list(string) |  false   |
| `worker_pool`                         | Configuration block for worker pools. There can be more than one.                                                                                                                                                                                                                 |          -         | list(object) |   true   |
| `worker_pool.count`                   | Number of workers in the worker pool. Can be changed afterwards to add or delete workers.                                                                                                                                                                                         |          1         |    number    |   true   |
| `worker_pool.labels`                  | Custom labels to assign to worker nodes.                                                                                                                                                                                                                                          |          -         |    string    |  false   |
| `worker_pool.virtual_cpus`            | List of tags that will be propagated to nodes in the worker pool.                                                                                                                                                                                                                 |          -         | map(string)  |  false   |
| `worker_pool.virtual_memory`          | Disable BGP on nodes. Nodes won't be able to connect to Packet BGP peers.                                                                                                                                                                                                         |        false       |     bool     |  false   |


## Applying

To create the cluster, execute the following command:

```console
lokoctl cluster apply
```

## Destroying

To destroy the Lokomotive cluster, execute the following command:

```console
lokoctl cluster destroy --confirm
```

