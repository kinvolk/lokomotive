# Contour Ingress Controller

## Requirements

The Contour Ingress component has different requirements on different
platforms. The reason for this is that an Ingress Controller needs traffic to be
routed to their ingress pods, and the network configurations needed to achieve
that differ on each platform.

Currently the following platforms are supported:
 * [Packet](#Requirements-to-run-on-Packet)


### Requirements to run on Packet

To run on [Packet](https://packet.com), the requirements are:

 * [MetalLB component](/docs/components/metallb.md) installed and configured

A typical symptom of not having MetalLB installed and configured correctly is
having the service in the `heptio-contour` namespace in pending state and
Contour won't be reachable from the internet. If that is the case, you probably
want to revisit your MetalLB configuration.

If Contour and all required components were installed correctly you should see a
external IP assigned to the service in the `heption-contour` namespace.

## Installation

Contour may be installed either as a
[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
or a
[DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/).
To deploy Contour as a DaemonSet, include in any file with the `.lokocfg`
extension the following:

```hcl
component "contour" {
	install_mode = "daemonset"
}
```

To deploy Contour as a Deployment, use `install_mode = "deployment"` instead.

Using DaemonSet `install_mode` spreads the load on all the cluster nodes, with
the obvious consequence of consuming resources on all cluster nodes, while
deployment mode uses only the desired replicas but each pod might have more
traffic load. This is a trade-off and it is up to your use case to choose one or
the other.

Then install the component by running:

```bash
lokoctl component install contour
```

For more information on Contour check the upstream [documentation](https://github.com/heptio/contour/tree/master/docs).
