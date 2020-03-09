# Lokomotive Components

`lokoctl` can be used to manage Lokomotive *components*. A component is a k8s workload which adds
functionality to a Lokomotive cluster. Without components, a Lokomotive cluster is just a barebone
k8s cluster. In some cases (e.g. bare metal environments), a Lokomotive cluster may even be
unusable without certain components.

Components can take care of tasks such as load balancing, monitoring, authentication, storage and
others.

## Listing Available Components

To list available Lokomotive components, run the following command:

```
lokoctl component list
```

Sample output:

```
Available components:
	 calico-hostendpoint-controller
	 cert-manager
	 cluster-autoscaler
	 contour
	 dex
	 external-dns
	 flatcar-linux-update-operator
	 gangway
	 httpbin
	 metallb
	 metrics-server
	 openebs-operator
	 openebs-storage-class
	 prometheus-operator
	 rook
	 rook-ceph
	 velero
```

## Installing a Component

To install a Lokomotive component, run the following command:

```
lokoctl component install <component_name>
```

This will take the `kubeconfig` from your cluster asset directory if you run it
in the folder of your cluster configuration files. Otherwise, the default kubeconfig
location will be used (`~/.kube/config`) unless you specify it through the `--kubeconfig`
flag or the `KUBECONFIG` environment variable.

A set of components to install may also be provided in a `.lokocfg` file:

```hcl
component "flatcar-linux-update-operator" {}

component "contour" {}
```

Specifying components in a `.lokocfg` file also allows passing configuration parameters to
components which support them. See the documentation for individual components for information about
the supported parameters.

To install all the components listed in a `.lokocfg` file, omit the component name:

```
lokoctl component install
```

>NOTE: `lokoctl` automatically detects all `.lokocfg` files in the working directory. This can be
>used to organize component configuration in separate files.

## Rendering a Manifest

Sometimes it can be useful to render a component's manifests without actually applying them to a
cluster, for example in order to verify what is going to be applied to a cluster. To render a
component's manifests, run the following command:

```
lokoctl component render-manifest <component_name>
```

Alternatively, omit the component name to render the manifests for all the components listed in
`.lokocfg` files (while templating them with any specified configuration parameters).
