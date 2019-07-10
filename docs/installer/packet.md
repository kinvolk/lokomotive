# Lokomotive Packet installation guide

This guide walks through a Lokomotive installation on [Packet](https://packet.net).

## Requirements

* An API token to a Packet account
* A Packet project ID
* AWS Route53 DNS Zone (registered domain name or delegated subdomain)
* Terraform v0.11.x
* [terraform-provider-ct](https://github.com/coreos/terraform-provider-ct) installed locally
* An SSH key pair for management access

## Install a Cluster

Create a `my-cluster.lokocfg` file to define your cluster and, optionally,
components that should be installed. Example:

```
variable "packet_token" {
	type = "string"
}

cluster "packet" {
	asset_dir = "/tmp/lokoctl-assets"
	auth_token = "${var.packet_token}"
	aws_creds_path = "${pathexpand("~/.aws/credentials")}"
	aws_region = "eu-central-1"
	cluster_name = "test"
	controller_count = 1
	dns_zone = "k8s.example.com"
	dns_zone_id = "XXX"
	facility = "ams1"
	project_id = "aaa-bbb-ccc-ddd"
	ssh_pubkeys = [
		"ssh-rsa AAAA...",
	]
	management_cidrs = ["123.45.67.89/32"]
	node_private_cidr = "XX.XX.XX.0/24"

	# Define one or more worker pools
	worker_pool "pool-1" {
	  # Define the number of worker nodes (required)
	  count = 1

	  # Define an instance type (optional)
	  # node_type = "t1.small.x86"

	  #  Define a Flatcar Linux channel (optional; 'stable', 'beta' or 'alpha')
	  # os_channel = "stable"

	  # Define a Flatcar Linux version (optional)
	  # os_version = "current"
	}
}

component "ingress-nginx" {
}
```

Quick note:

The asset directory should be kept for the lifetime of the cluster.
The path cannot be relative at the moment.

`management_cidrs` is the list of IPv4 CIDRs authorised to access or manage the cluster.

For `node_private_cidr`, if you do not know the actual private IP address CIDR that will be assigned to the nodes, you can use the project blocks on https://app.packet.net/projects/<PROJECT_ID>/network as a guide.

Next,

Create a `lokocfg.vars` file and define all needed variables. Example:

```
packet_token = "XXX"
```

To apply the configuration, run

```
lokoctl cluster install
```

Terraform will generate Bootkube assets to the directory specified with `asset_dir`. Terraform will
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

## Destroying the cluster

```bash
cd <asset_dir>/terraform/
terraform destroy
```
