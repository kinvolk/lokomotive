# Run conformance tests on Lokomotive

## Run the tests

Create a lokomotive cluster. In this case I followed the [packet tutorial](../flatcar-linux/packet.md) and my terraform files were the following.

**Note:** This example uses `master` as the reference for lokomotive-kubernetes but for a conformance test intended for submission you should use the corresponding version tag.

```hcl
# providers.tf

provider "aws" {
  version = "~> 2.25.0"
  alias = "default"

  region = "eu-central-1"
  shared_credentials_file = "/home/user/.aws/credentials"
}

provider "ct" {
  version = "0.3.1"
}

provider "local" {
  version = "~> 1.2"
  alias = "default"
}

provider "null" {
  version = "~> 2.1"
  alias = "default"
}

provider "template" {
  version = "~> 2.1"
  alias = "default"
}

provider "tls" {
  version = "~> 2.0"
  alias = "default"
}

provider "packet" {
  version = "~> 1.2"
  alias   = "default"
}
```

```hcl
# cluster.tf

module "controller" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//packet/flatcar-linux/kubernetes?ref=master"

  providers = {
    aws      = "aws.default"
    local    = "local.default"
    null     = "null.default"
    template = "template.default"
    tls      = "tls.default"
    packet   = "packet.default"
  }

  # Route53
  dns_zone    = "..."
  dns_zone_id = "..."

  # configuration
  ssh_keys = [
    "...",
  ]

  asset_dir = "./assets/conformance-cluster"

  # Packet
  cluster_name = "conformance-cluster"
  project_id   = "..."
  facility     = "ams1"

  # This must be the total of all worker pools
  worker_count              = 2
  worker_nodes_hostnames    = "${module.worker.worker_nodes_hostname}"

  # optional
  controller_count = 1
  controller_type  = "t1.small.x86"

  management_cidrs = [
    "0.0.0.0/0",       # Instances can be SSH-ed into from anywhere on the internet.
  ]

  # This is different for each project on Packet and depends on the packet facility/region. Check yours from the `IPs & Networks` tab from your Packet.net account. If an IP block is not allocated yet, try provisioning an instance from the console in that region. Packet will allocate a public IP CIDR.
  node_private_cidr = "..."
}

module "worker" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//packet/flatcar-linux/kubernetes/workers?ref=master"

  providers = {
    local    = "local.default"
    template = "template.default"
    tls      = "tls.default"
    packet   = "packet.default"
  }

  ssh_keys = [
    "...",
  ]

  # Packet
  cluster_name = "conformance-cluster"
  project_id   = "..."
  facility     = "ams1"

  pool_name    = "worker"

  count = 2
  type  = "t1.small.x86"

  kubeconfig = "${module.controller.kubeconfig}"

  labels = "node.conformance.io/role=backend,node-role.kubernetes.io/backend="
}
```

Once the cluster is created, configure your kubectl and make sure you can connect to it with kubectl:

```
$ export KUBECONFIG="$PWD/assets/conformance-cluster/auth/kubeconfig"
$ kubectl get nodes
NAME                                  STATUS   ROLES               AGE   VERSION
conformance-cluster-controller-0      Ready    controller,master   71s   v1.15.3
conformance-cluster-worker-worker-0   Ready    backend,node        71s   v1.15.3
conformance-cluster-worker-worker-1   Ready    backend,node        71s   v1.15.3
```

Download latest binary release of sonobuoy:

```
wget https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.16.2/sonobuoy_0.16.2_linux_amd64.tar.gz
wget https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.16.2/sonobuoy_0.16.2_checksums.txt
sha256sum --ignore-missing -c sonobuoy_0.16.2_checksums.txt
tar xvf sonobuoy_0.16.2_linux_amd64.tar.gz
```

Run sonobuoy in conformance mode:

```
./sonobuoy run --mode=certified-conformance
```

Wait for sonobuoy to finish (might take a long time). You can check its status by running:

```
./sonobuoy status
```

Once the run shows as `completed`, copy the output to a local directory:

```
outfile=$(./sonobuoy retrieve)
```

Now you can extract the results snapshot:

```
mkdir ./results; tar xzf $outfile -C ./results
```

Finally, clean up the Kubernetes objects created by Sonobuoy:

```
./sonobuoy delete
```

## Make a PR to get into k8s-conformance

Check the tests pass by going to the end of the file `results/plugins/e2e/results/global/e2e.log` and, if so, create a PR with the results as explained in the [documentation](https://github.com/cncf/k8s-conformance/blob/master/instructions.md#uploading).

For `PRODUCT.yaml`, you can use the following values:

```
vendor: Kinvolk GmbH
name: Lokomotive Kubernetes
version: $LOKOMOTIVE_VERSION
website_url: https://github.com/kinvolk/lokomotive-kubernetes
documentation_url: https://github.com/kinvolk/lokomotive-kubernetes/blob/master/docs/index.md
product_logo_url: https://raw.githubusercontent.com/kinvolk/lokomotive-kubernetes/master/docs/images/lokomotive-logo.svg
type: distribution
description: Lokomotive is an open source project by Kinvolk which distributes pure upstream Kubernetes via Terraform.
```

For the instructions on how to reproduce you can use the first section of this document.
