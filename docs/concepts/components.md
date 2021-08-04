---
title: Lokomotive component
weight: 10
---

A Lokomotive component is a Kubernetes workload which adds functionality to a Lokomotive cluster.
Without components, a Lokomotive cluster is just a barebone Kubernetes cluster. In some cases (e.g.
bare metal environments), a Lokomotive cluster may even be unusable without certain components.
Components take care of tasks such as load balancing, monitoring, authentication, storage and
others.

`lokoctl` is a command line interface for Lokomotive and its *components*.

## Categories of Lokomotive components

### User Authentication

Lokomotive provides the [Dex](../configuration-reference/components/dex.md) and
[Gangway](../configuration-reference/components/gangway.md) components for user authentication via
OpenID Connect (OIDC). With these components, you can securely manage access to the Kubernetes
cluster and resources.

Lokomotive also provides the [cert-manager](../configuration-reference/components/cert-manager.md)
component for automating the management and issuance of TLS certificates from various issuing
sources.

### Monitoring/Metrics

Lokomotive provides a
[prometheus-operator](../configuration-reference/components/prometheus-operator.md) component that
creates, configures and manages [Prometheus](https://prometheus.io/) atop Kubernetes.

Lokomotive also provides a [metrics-server](../configuration-reference/components/metrics-server.md)
component responsible for collecting resource metrics from nodes and pods and exposing them in the
Kubernetes API server through the [metrics API](https://github.com/kubernetes/metrics).

### Storage

Lokomotive provides the [openebs-operator](../configuration-reference/components/openebs-operator.md) and
[openebs-storage-class](../configuration-reference/components/openebs-storage-class.md) components to
use OpenEBS as the block storage solution for the cluster.

Lokomotive also provides the [rook](../configuration-reference/components/rook.md) and
[rook-ceph](../configuration-reference/components/rook-ceph.md) components for using Rook as the storage
solution for Lokomotive cluster.

### Load Balancing/Ingress (for Equinix Metal platform only)

Lokomotive provides the [MetalLB](../configuration-reference/components/metallb.md) component for load
balancing in platforms without network load-balancers and
[Contour](../configuration-reference/components/contour.md) component for Ingress control.

Lokomotive also provides the [external-dns](../configuration-reference/components/external-dns.md)
component for automatic management of DNS entries for Ingress resources.

### Node management

Lokomotive provides the
[cluster-autoscaler](../configuration-reference/components/cluster-autoscaler.md) component for adjusting
the size of Lokomotive cluster.

### Update

Lokomotive provides the
[flatcar-linux-update-operator](../configuration-reference/components/flatcar-linux-update-operator.md)
component for orchestrating updates of the Flatcar Container Linux OS on cluster nodes.

## Listing Available Components

To list available Lokomotive components, run the following command:

```
lokoctl component list
```

Sample output:

```
Available components:
	 aws-ebs-csi-driver
	 cert-manager
	 cluster-autoscaler
	 contour
	 dex
	 experimental-istio-operator
	 experimental-linkerd
	 external-dns
	 flatcar-linux-update-operator
	 gangway
	 web-ui
	 httpbin
	 inspektor-gadget
	 metallb
	 metrics-server
        node-problem-detector
	 openebs-operator
	 openebs-storage-class
	 prometheus-operator
	 rook
	 rook-ceph
	 velero
```

## Installing a Component

To install a Lokomotive component add it to a `.lokocfg` file:

```hcl
component "flatcar-linux-update-operator" {}

component "contour" {}
```

Then you can apply a particular component:

```console
lokoctl component apply <component_name>
```

> If this command is executed in the directory containing cluster configuration, `lokoctl` will try
> to install the component on configured cluster. If no configuration is found, the configuration
> from `KUBECONFIG` environment variable will be used. If the environment variable is empty,
> `~/.kube/config` file will be used.

>To use specific `kubeconfig` file, `--kubeconfig` flag can be used.

You can pass configuration parameters to components, check the [component reference
documentation](../configuration-reference/components) for details.

To install all the components listed in a `.lokocfg` file, omit the component name:

```console
lokoctl component apply
```

>NOTE: `lokoctl` automatically detects all `.lokocfg` files in the working directory. This can be
>used to organize component configuration in separate files.

Installing a Lokomotive component is the same operation as applying its latest configuration, so if
you change the configuration in a `.lokocfg` file you can run apply again to apply the new
configuration to the cluster.

## Rendering a Manifest

Sometimes it can be useful to render a component's manifests without actually applying them to a
cluster, for example in order to verify what is going to be applied to a cluster. To render a
component's manifests, run the following command:

```console
lokoctl component render-manifest <component_name>
```

Alternatively, omit the component name to render the manifests for all the components listed in
`.lokocfg` files (while templating them with any specified configuration parameters).
