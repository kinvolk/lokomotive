# bootkube

`bootkube` is a Terraform module that renders [kubernetes-incubator/bootkube](https://github.com/kubernetes-incubator/bootkube) assets for bootstrapping a Kubernetes cluster.

## Audience

`bootkube` is a low-level component of the [Lokomotive](https://github.com/kinvolk/lokomotive-kubernetes) Kubernetes distribution. Use Lokomotive modules to create and manage Kubernetes clusters across supported platforms. Use the bootkube module if you'd like to customize a Kubernetes control plane or build your own distribution.

## Usage

Use the module to declare bootkube assets. Check [variables.tf](variables.tf) for options and [terraform.tfvars.example](terraform.tfvars.example) for examples.

```hcl
module "bootkube" {
  source = "git::https://github.com/kinvolk/lokomotive-kubernetes//bootkube?ref=SHA"

  cluster_name = "example"
  api_servers = ["node1.example.com"]
  etcd_servers = ["node1.example.com"]
  asset_dir = "/home/core/clusters/mycluster"
}
```

Generate the assets.

```sh
terraform init
terraform plan
terraform apply
```

Find bootkube assets rendered to the `asset_dir` path. That's it.

### Comparison

Render bootkube assets directly with bootkube v0.14.0.

```sh
bootkube render --asset-dir=assets --api-servers=https://node1.example.com:6443 --api-server-alt-names=DNS=node1.example.com --etcd-servers=https://node1.example.com:2379
```

Compare assets. Rendered assets may differ slightly from bootkube assets to reflect decisions made by the [Lokomotive](https://github.com/kinvolk/lokomotive-kubernetes) distribution.

```sh
pushd /home/core/mycluster
mv manifests-networking/* manifests
popd
diff -rw assets /home/core/mycluster
```
