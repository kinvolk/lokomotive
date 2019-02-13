# Packet Installation

This guide walks through a Lokomotive installation on [Packet](https://packet.net).

## Requirements

* An API token to a Packet account
* A Packet project ID
* AWS Route53 DNS Zone (registered domain name or delegated subdomain)
* Terraform v0.11.x
* [terraform-provider-ct](https://github.com/coreos/terraform-provider-ct) installed locally
* An SSH key pair for management access

## Install a Cluster

Get [lokoctl](https://github.com/kinvolk/lokoctl) and build it by running `make` in the project's
root.

Set the `PACKET_AUTH_TOKEN` environment variable with the Packet API token:

```
export PACKET_AUTH_TOKEN=xxxxxxxx
```

Run `./lokoctl install packet` with the required flags specified:

```
./lokoctl install packet \
    --assets /tmp/lokoctl-assets \
    --aws-region eu-central-1 \
    --cluster-name my-cluster \
    --aws-creds ~/.aws/credentials \
    --dns-zone myclusters.example.com \
    --dns-zone-id Z3PAABBCFAKEC0 \
    --facility ams1 \
    --project-id 4cff83ac-de23-432a-b01b-b2950dabc76e \
    --ssh-public-key ~/.ssh/my-cluster.pub \
    --worker-count 1
```

Use `-h` for information regarding additional flags.

Terraform will generate Bootkube assets to the directory specified with `--assets`. Terraform will
then create the machines on Packet and loop until it can successfully copy credentials to each
machine and start the one-time Kubernetes bootstrap service.

```
...
module.packet-my-cluster.null_resource.copy-controller-secrets: Creating...
module.packet-my-cluster.null_resource.copy-controller-secrets: Provisioning with 'file'...
module.packet-my-cluster.null_resource.copy-controller-secrets: Still creating... (10s elapsed)
module.packet-my-cluster.null_resource.copy-controller-secrets: Still creating... (20s elapsed)
module.packet-my-cluster.null_resource.copy-controller-secrets: Still creating... (30s elapsed)
module.packet-my-cluster.null_resource.copy-controller-secrets: Still creating... (40s elapsed)
...
```

### Bootstrap

Wait for the bootkube-start step to finish bootstrapping the Kubernetes control plane. This may
take 5-15 minutes.

```
...
module.packet-johannes-test.null_resource.bootkube-start: Still creating... (4m50s elapsed)
module.packet-johannes-test.null_resource.bootkube-start: Still creating... (5m0s elapsed)
module.packet-johannes-test.null_resource.bootkube-start: Still creating... (5m10s elapsed)
module.packet-johannes-test.null_resource.bootkube-start: Still creating... (5m20s elapsed)
module.packet-johannes-test.null_resource.bootkube-start: Creation complete after 5m26s (ID: 6276739637382861631)

Apply complete! Resources: 56 added, 0 changed, 0 destroyed.

Your configurations are stored in /tmp/lokoctl-assets
```

To watch the instances during the initial OS provisioning, you can use the Packet
[out-of-band console service](https://support.packet.com/kb/articles/sos-serial-over-ssh). For
example:

```
ssh 89cd1d28-32ca-432a-812c-ff0fc38fcbda@sos.ams1.packet.net
```

To watch the bootstrap process in detail, SSH to the first controller machine and watch the logs
using `journalctl`:

```
ssh core@147.1.2.3
journalctl -f -u bootkube
```

Sample output:

```
bootkube[5]:         Pod Status:        pod-checkpointer        Running
bootkube[5]:         Pod Status:          kube-apiserver        Running
bootkube[5]:         Pod Status:          kube-scheduler        Running
bootkube[5]:         Pod Status: kube-controller-manager        Running
bootkube[5]: All self-hosted control plane components successfully started
bootkube[5]: Tearing down temporary bootstrap control plane...
```

## Verify

Install `kubectl` on your system. Use the generated `kubeconfig` credentials to access the
Kubernetes cluster and list the nodes:

```
export KUBECONFIG=/tmp/lokoctl-assets/auth/kubeconfig
kubectl get nodes
```

Sample output:

```
NAME                         STATUS   ROLES               AGE   VERSION
my-cluster-controller-0   Ready    controller,master   85m   v1.13.1
my-cluster-worker-0       Ready    node                85m   v1.13.1
```

List the pods:

```
kubectl get pods --all-namespaces
```

Sample output:

```
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

## Clean up
Run `terraform destroy` inside the `terraform` directory in the assets directory:
```
cd /tmp/lokoctl-assets/terraform
terraform destroy
```
