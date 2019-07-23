# Lokomotive Packet installation guide

This guide walks through a Lokomotive installation on [Packet](https://packet.net).

## Requirements

* An API token to a Packet account (do not use the project token but a profile token)
* A Packet project ID
* Local BGP enabled
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
	# Change asset folder
	asset_dir = "${pathexpand("~/lokoctl-assets/mycluster")}"
	auth_token = "${var.packet_token}"
	aws_creds_path = "${pathexpand("~/.aws/credentials")}"
	# Change according to your AWS DNS zone
	aws_region = "eu-central-1"
	# Change cluster name
	cluster_name = "test"
	controller_count = 1
	# Change AWS DNS zone
	dns_zone = "k8s.example.com"
	# and zone ID
	dns_zone_id = "XXX"
	# Change Packet server location
	facility = "ams1"
	# Change Packet project ID
	project_id = "aaa-bbb-ccc-ddd"
	# Change management SSH public key
	ssh_pubkeys = [
		"ssh-rsa AAAA...",
	]
	# Change to your external IP address to allow it for management access
	management_cidrs = ["my.ip.ad.dr/32"]
	# Change to internal Packet IPs to allow cluster communication
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

component "contour" {
  install_mode = "daemonset"
}
```

Quick note:

The asset directory should be kept for the lifetime of the cluster.
The path cannot be relative at the moment.

`management_cidrs` is the list of IPv4 CIDRs authorised to access or manage the cluster.
Find your current external IP with `curl -4 icanhazip.com` and put it there.
You can put `0.0.0.0/0` to allow any address if you cannot predict your IP address.

For `node_private_cidr`, if you do not know the actual private IP address CIDR that
will be assigned to the nodes, you can copy the project blocks from https://app.packet.net/projects/<PROJECT_ID>/network.
If you don't know the exact block in advance you can put `10.0.0.0/8` to allow any address
from the internal packet network for node-to-node communication. Using `0.0.0.0/0` does not
work because calico tries to find its network interface based on reaching that.

You also have to specify the `project_id` variable in the configuration file (as seen in the URL of the Packet web interface),
add SSH public keys to the list (take the content from `~/.ssh/id_rsa.pub`),
and change the `dns_zone` and `dns_zone_id` values matching those you set up in AWS.

Next,

create a `lokocfg.vars` file and define all needed variables. Example:

```
packet_token = "XXX"
```

When you store your configuration in a git repository, do not include the `lokocfg.vars` file which holds
the Packet authentification token (Consider adding it to `.gitignore`).

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
