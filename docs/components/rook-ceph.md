# Rook Ceph

[Rook](https://rook.io) is a storage orchestrator for Kubernetes. This component installs a Ceph cluster managed by the Rook operator.

## Requirements

The Rook component deployed on the cluster.

## Installation

#### `.lokocfg` file

Declare the component's config:

```
component "rook-ceph" {
  namespace = "rook-test"

  monitor_count = 3

  node_selector {
    key      = "node-role.kubernetes.io/storage"
    operator = "Exists"
  }

  node_selector {
    key      = "storage.lokomotive.io"
    operator = "In"

    # If the `operator` is set to `"In"`, `values` should be specified.
    values = [
      "foo",
    ]
  }

  toleration {
    key      = "storage.lokomotive.io"
    operator = "Equal"
    value    = "rook-ceph"
    effect   = "NoSchedule"
  }
}
```

Run the following command:

```bash
lokoctl component install rook-ceph
```

## Limitations

The Ceph cluster needs to be deployed in the same namespace as the Rook operator at the moment. Additional `Roles` and `RoleBindings` need to be created if deploying across separate namespaces is desired.

## Next steps

Once the Ceph cluster is ready, an object store can be deployed to start writing to Ceph. More information is available here: https://rook.io/docs/rook/v1.0/ceph-object-store-crd.html

## Cleanup

After removing the component from the cluster, make sure to delete `/var/lib/rook` from hostpath of all worker nodes for a clean reinstallation.

## Argument Reference

| Argument | Explanation | Default | Required |
|----------|-------------|---------|----------|
| `namespace` | Namespace to deploy the ceph cluster into. Must be same as the rook operator. | rook | false |
| `monitor_count` | Number of ceph monitors to deploy. Odd number like 3 or 5 is recommended which should also be sufficient for most cases. | 1 | false |
| `node_selector` | Node selectors for deploying the ceph cluster pods. | - | false |
| `toleration` | Tolerations that the ceph cluster's pods will tolerate | - | false |
| `metadata_device` | Name of the device to store the metadata on each storage machine. **Note**: Provide just the name of the device and skip prefixing `/dev/`. | - | "md127" |
