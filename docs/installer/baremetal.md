# Lokomotive bare-metal installation guide

This guide walks through a bare metal installation of Lokomotive utilizing PXE.

Note that the guide is tailored for a matchbox setup and not generally
useful for PXE environments.

## Requirements

A minimum of two machines are required to run Lokomotive.

* Machines with 2GB RAM, 30GB disk, PXE-enabled NIC, IPMI
* PXE-enabled [network boot](https://coreos.com/matchbox/docs/latest/network-setup.html) environment
* Matchbox v0.6+ deployment with API enabled
* Matchbox credentials `client.crt`, `client.key`, `ca.crt`
* Terraform v0.11.x, [terraform-provider-matchbox](https://github.com/poseidon/terraform-provider-matchbox), and [terraform-provider-ct](https://github.com/poseidon/terraform-provider-ct) installed locally

Note that the machines should only be powered on *after* starting the
installation, see below.

### Machines

* Mac addresses collected from each machine.

    For machines with multiple PXE-enabled NICs, pick one of the MAC addresses. MAC addresses will be used to match machines to profiles during network boot.

* DNS A (or AAAA) record for each node's default interface.

    Cluster nodes will be configured to refer to the control plane and themselves by these fully qualified names and they will be used in generated TLS certificates.

* SSH access to all machines

## Installing the cluster

Create a `my-cluster.lokocfg` file to define your cluster and, optionally,
components that should be installed. Example:

```
cluster "bare-metal" {
  asset_dir = "${pathexpand("~/.lokoctl/mercury")}"
  ssh_pubkey = "${pathexpand("~/.ssh/id_rsa.pub")}"
  cached_install = "true"
  matchbox_ca_path = "${pathexpand("~/matchbox-certs/ca.crt")}"
  matchbox_client_cert_path = "${pathexpand("~/matchbox-certs/client.crt")}"
  matchbox_client_key_path = "${pathexpand("~/matchbox-certs/client.key")}"
  matchbox_endpoint = "matchbox.example.com:8081"
  matchbox_http_endpoint = "http://matchbox.example.com:8080"
  cluster_name = "mercury"
  k8s_domain_name = "node1.example.com"
  controller_domains = [
    "node1.example.com",
  ]
  controller_macs = [
    "52:54:00:a1:9c:ae",
  ]
  controller_names = [
    "node1",
  ]
  worker_domains = [
    "node2.example.com",
    "node3.example.com",
  ]
  worker_macs = [
    "52:54:00:b2:2f:86",
    "52:54:00:c3:61:77",
  ]
  worker_names = [
    "node2",
    "node3",
  ]
}

component "contour" {}
```

>NOTE: The asset directory should be kept for the lifetime of the cluster. For more information
>regarding the asset directory, see [here](general.md#asset-directory).

To apply the configuration, run

```
lokoctl cluster install
```

Apply will then loop until it can successfully copy credentials to each machine and start the one-time Kubernetes bootstrap service.

**Proceed to Power on the PXE machines while this loops.**

### Bootstrap

Wait for the bootkube-start step to finish bootstrapping the Kubernetes control plane. This may take 5-15 minutes depending on your network.

```
...
module.bare-metal-mercury.null_resource.bootkube-start: Still creating... (2m40s elapsed)
module.bare-metal-mercury.null_resource.bootkube-start: Still creating... (2m50s elapsed)
module.bare-metal-mercury.null_resource.bootkube-start: Still creating... (3m0s elapsed)
module.bare-metal-mercury.null_resource.bootkube-start: Still creating... (3m10s elapsed)
module.bare-metal-mercury.null_resource.bootkube-start: Creation complete after 3m18s (ID: 7372450104484005399)

Apply complete! Resources: 59 added, 0 changed, 0 destroyed.

Your configurations are stored in /path/to/assets/directory
```

To watch the install to disk (until machines reboot from disk), SSH to port 2222. e.g

```
ssh -p 2222 core@node1.example.com
```

To watch the bootstrap process in detail, SSH to the first controller and journal the logs.

```
$ ssh core@node1.example.com
$ journalctl -f -u bootkube
bootkube[5]:         Pod Status:        pod-checkpointer        Running
bootkube[5]:         Pod Status:          kube-apiserver        Running
bootkube[5]:         Pod Status:          kube-scheduler        Running
bootkube[5]:         Pod Status: kube-controller-manager        Running
bootkube[5]: All self-hosted control plane components successfully started
bootkube[5]: Tearing down temporary bootstrap control plane...
```

## Verify

Install kubectl on your system. Use the generated `kubeconfig` credentials to access the Kubernetes cluster and list nodes.

```
$ export KUBECONFIG="<asset_dir>/auth/kubeconfig"
$ kubectl get nodes
NAME                STATUS  ROLES              AGE  VERSION
node1.example.com   Ready   controller,master  10m  v1.12.2
node2.example.com   Ready   node               10m  v1.12.2
node3.example.com   Ready   node               10m  v1.12.2
```

List the pods.

```
$ kubectl get pods --all-namespaces
NAMESPACE     NAME                                       READY     STATUS    RESTARTS   AGE
kube-system   calico-node-6qp7f                          2/2       Running   1          11m
kube-system   calico-node-gnjrm                          2/2       Running   0          11m
kube-system   calico-node-llbgt                          2/2       Running   0          11m
kube-system   coredns-1187388186-dj3pd                   1/1       Running   0          11m
kube-system   coredns-1187388186-mx9rt                   1/1       Running   0          11m
kube-system   kube-apiserver-7336w                       1/1       Running   0          11m
kube-system   kube-controller-manager-3271970485-b9chx   1/1       Running   0          11m
kube-system   kube-controller-manager-3271970485-v30js   1/1       Running   1          11m
kube-system   kube-proxy-50sd4                           1/1       Running   0          11m
kube-system   kube-proxy-bczhp                           1/1       Running   0          11m
kube-system   kube-proxy-mp2fw                           1/1       Running   0          11m
kube-system   kube-scheduler-3895335239-fd3l7            1/1       Running   1          11m
kube-system   kube-scheduler-3895335239-hfjv0            1/1       Running   0          11m
kube-system   pod-checkpointer-wf65d                     1/1       Running   0          11m
kube-system   pod-checkpointer-wf65d-node1.example.com   1/1       Running   0          11m
```

## Destroying the cluster

```bash
lokoctl cluster destroy --confirm
```

You will then need to manually delete the assets directory

```bash
rm -r <asset_dir>/<cluster_name>
```