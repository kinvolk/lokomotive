# Lokomotive <img align="right" width=384 src="docs/images/lokomotive-logo.svg">

Lokomotive is an open source project by [Kinvolk](https://kinvolk.io/) which distributes pure upstream Kubernetes.

## Features

* Kubernetes v1.17.0 (upstream, via [kubernetes-incubator/bootkube](https://github.com/kubernetes-incubator/bootkube))
* Single or multi-master, [Calico](https://www.projectcalico.org/) or [flannel](https://github.com/coreos/flannel) networking
* On-cluster etcd with TLS, [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/)-enabled, [network policy](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
* Advanced features like [worker pools](docs/advanced/worker-pools.md) and [snippets](docs/advanced/customization.md#flatcar-linux) customization

## Modules

Lokomotive provides a Terraform Module for each supported operating system and platform. Flatcar Container Linux is a mature and reliable choice.

| Platform      | Operating System        | Terraform Module | Status |
|---------------|-------------------------|------------------|--------|
| AWS           | Flatcar Container Linux | [aws/flatcar-linux/kubernetes](docs/flatcar-linux/aws.md) | stable |
| Azure         | Flatcar Container Linux | [azure/flatcar-linux/kubernetes](docs/flatcar-linux/azure.md) | alpha |
| Bare-Metal    | Flatcar Container Linux | [bare-metal/flatcar-linux/kubernetes](docs/flatcar-linux/bare-metal.md) | stable |
| Packet        | Flatcar Container Linux | [packet/flatcar-linux/kubernetes](docs/flatcar-linux/packet.md) | beta |

## Documentation

* Architecture [concepts](docs/architecture/concepts.md) and [operating-systems](docs/architecture/operating-systems.md)
* Tutorials for [AWS](docs/flatcar-linux/aws.md), [Azure](docs/flatcar-linux/azure.md), [Bare-Metal](docs/flatcar-linux/bare-metal.md) and [Packet](docs/flatcar-linux/packet.md)

## Usage

Define a Kubernetes cluster by using the Terraform module for your chosen platform and operating system. Here's a minimal example.

```tf
module "aws-tempest" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//aws/flatcar-linux/kubernetes?ref=master"

  providers = {
    aws = aws.default
    local = local.default
    null = null.default
    template = template.default
    tls = tls.default
  }

  # AWS
  cluster_name = "yavin"
  dns_zone     = "example.com"
  dns_zone_id  = "Z3PAABBCFAKEC0"

  # configuration
  ssh_keys = [
    "ssh-rsa AAAAB3Nz...",
    "ssh-rsa AAAAB3Nz...",
  ]

  asset_dir          = "/home/user/.secrets/clusters/yavin"

  # optional
  worker_count = 2
  worker_type  = "t3.small"
}
```

Initialize modules, plan the changes to be made, and apply the changes.

```sh
$ terraform init
$ terraform plan
Plan: 64 to add, 0 to change, 0 to destroy.
$ terraform apply
Apply complete! Resources: 64 added, 0 changed, 0 destroyed.
```

In 4-8 minutes (varies by platform), the cluster will be ready. This AWS example creates a `yavin.example.com` DNS record to resolve to a network load balancer backed by controller instances.

```sh
$ export KUBECONFIG=/home/user/.secrets/clusters/yavin/auth/kubeconfig
$ kubectl get nodes
NAME                                       ROLES              STATUS  AGE  VERSION
yavin-controller-0.c.example-com.internal  controller,master  Ready   6m   v1.14.1
yavin-worker-jrbf.c.example-com.internal   node               Ready   5m   v1.14.1
yavin-worker-mzdm.c.example-com.internal   node               Ready   5m   v1.14.1
```

List the pods.

```
$ kubectl get pods --all-namespaces
NAMESPACE     NAME                                      READY  STATUS    RESTARTS  AGE
kube-system   calico-node-1cs8z                         2/2    Running   0         6m
kube-system   calico-node-d1l5b                         2/2    Running   0         6m
kube-system   calico-node-sp9ps                         2/2    Running   0         6m
kube-system   coredns-1187388186-dkh3o                  1/1    Running   0         6m
kube-system   kube-apiserver-zppls                      1/1    Running   0         6m
kube-system   kube-controller-manager-3271970485-gh9kt  1/1    Running   0         6m
kube-system   kube-controller-manager-3271970485-h90v8  1/1    Running   1         6m
kube-system   kube-proxy-117v6                          1/1    Running   0         6m
kube-system   kube-proxy-9886n                          1/1    Running   0         6m
kube-system   kube-proxy-njn47                          1/1    Running   0         6m
kube-system   kube-scheduler-3895335239-5x87r           1/1    Running   0         6m
kube-system   kube-scheduler-3895335239-bzrrt           1/1    Running   1         6m
kube-system   pod-checkpointer-l6lrt                    1/1    Running   0         6m
kube-system   pod-checkpointer-l6lrt-controller-0       1/1    Running   0         6m
```

## Try Flatcar Container Linux Edge

[Flatcar Container Linux Edge](https://kinvolk.io/blog/2019/05/introducing-the-flatcar-linux-edge-channel/) is a [Flatcar Container Linux](https://www.flatcar-linux.org/) channel that includes experimental bleeding-edge features.

To try it just add the following configuration option to the example above.

```
os_image = "flatcar-edge"
```

## Help

Ask questions on the IRC #lokomotive-k8s channel on [freenode.net](http://freenode.net/).
