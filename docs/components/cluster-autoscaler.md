The [Cluster Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler)
is a tool that automatically adjusts the size of a Kubernetes cluster when
there are pods that failed to run in the cluster due to insufficient resources
or when there are nodes that have been underutilized for an extended period of
time and pods can be placed on other existing nodes.

## Lokomotive component

The Cluster Autoscaler is available as a component in lokoctl.

## Compatibility

Currently lokoctl supports the Cluster Autoscaler only on the Packet platform.
Support for other platforms will be added in the future.

## Requirements

To install the Cluster Autoscaler you need:

* A running Lokomotive cluster.
* A Packet API token stored in the `PACKET_AUTH_TOKEN` environment variable.

## Configuration

The Cluster Autoscaler lokoctl component currently supports the following
options (for optional parameters, the values shown here are the defaults):

```
# cluster-autoscaler.lokocfg
component "cluster-autoscaler" {
  # Required parameters
  #
  # Name of your cluster
  cluster_name = "cluster"
  # Name of your worker pool
  worker_pool = "pool"

  # Optional parameters
  #
  # The only supported provider is 'packet' so this parameter is optional
  provider = "packet"
  # Namespace where the Cluster Autoscaler will run
  namespace = "kube-system"
  # Minimum number of workers in the worker pool
  min_workers = 1
  # Maximum number of workers in the worker pool
  max_workers = 4
  # How long a node should be unneeded before it is eligible for scale down
  scale_down_unneeded_time = "10m"
  # How long after scale up that scale down evaluation resumes
  scale_down_delay_after_add = "10m"
  # How long an unready node should be unneeded before it is eligible for scale down
  scale_down_unready_time = "20m"

  packet {
    # Required parameters when using provider 'packet'
    #
    # Packet Project ID where your cluster is running in
    project_id = "0b5dc3d2-949a-447d-82cd-43fbdc1ae8c0"
    # Packet Facility where your cluster is running in
    facility = "sjc1"

    # Optional parameters when using provider 'packet'
    #
    # Machine type for workers spawned by the Cluster Autoscaler
    worker_type = "t1.small.x86"
    # Flatcar Container Linux channel to be used by workers spawned by the Cluster Autoscaler
    worker_channel = "stable"
  }
}
```

### Installation

After preparing your configuration in a lokocfg file (e.g.
`cluster-autoscaler.lokocfg`), you can install the component with

```
lokoctl component install cluster-autoscaler
```

By default, the Cluster Autoscaler pod runs in the `kube-system` namespace, see
`kubectl get pods -n kube-system`.

## Next steps

Once you have successfully installed the Cluster Autoscaler, when your cluster
needs more resources it will automatically create new nodes.

Also, when some of your nodes are unneeded for more than
`scale_down_unneeeded_time`, they will be removed from your cluster.

It is recommended that you install the [Calico HostEndpoint controller
component](calico-hostendpoint-controller.md) to secure new nodes created by
the Cluster Autoscaler, otherwise they will be exposed.

## Caveats

If you already have worker nodes on your cluster, they will not be considered
by the Cluster Autoscaler. To do that you need to tag them in advance manually
with the following tags:

```
k8s-cluster-${cluster_name}
k8s-nodepool-${worker_pool}
```
