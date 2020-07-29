# Lokomotive <img align="right" width=384 src="docs/images/lokomotive-logo.svg">

Lokomotive is an open source Kubernetes distribution that ships pure upstream
Kubernetes.
It focuses on being minimal, easy to use, and secure by default.

Lokomotive is fully self-hosted and includes [Lokomotive
Components](docs/concepts/components.md) adding essential functionality for production
not shipped with upstream Kubernetes.

<img src="docs/images/lokomotive-example.gif" alt="Example gif showing `lokoctl cluster apply --confirm`" width="700"/>

## Features

<a href="https://landscape.cncf.io/selected=lokomotive"><img src="https://raw.githubusercontent.com/cncf/artwork/1c1a10d9cc7de24235e07c8831923874331ef233/projects/kubernetes/certified-kubernetes/versionless/color/certified-kubernetes-color.svg" align="right" width="100px"></a>

* Kubernetes 1.18 (upstream, via
  [kubernetes-incubator/bootkube](https://github.com/kubernetes-incubator/bootkube))
* Fully self-hosted, including the kubelet
* Single or multi-master
* [Calico](https://www.projectcalico.org/) networking
* On-cluster etcd with TLS,
  [RBAC](https://kubernetes.io/docs/admin/authorization/rbac/)-enabled,
  [network policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

## Installation

Lokomotive provides the lokoctl CLI tool to manage clusters.
Check the [installation guide](docs/installer/lokoctl.md) to install it.

## Getting started

Follow one of the quickstart guides for the supported platforms:

* [Packet quickstart](docs/quickstarts/packet.md)
* [AWS quickstart](docs/quickstarts/aws.md)
* [Bare metal quickstart](docs/quickstarts/baremetal.md)

## Documentation

### Reference guides

* [Platform configuration references](docs/configuration-reference/platforms)
* [Component configuration references](docs/configuration-reference/components)
* [CLI reference](docs/cli/lokoctl.md)

### How to guides

* [Setting up cluster authentication on Lokomotive with GitHub, Dex and Gangway](docs/how-to-guides/authentication-with-dex-gangway.md)
* [Setting up an HTTP ingress controller on Lokomotive with MetalLB and Contour on Packet](docs/how-to-guides/ingress-with-contour-metallb.md))

## Issues

Please file [issues](https://github.com/kinvolk/lokomotive/issues) on this
repository.

Before filing an issue, please ensure you have searched through / reviewed
existing issues.

If an issue or PR you’d like to contribute to is already assigned to someone,
please reach out to them to coordinate your work.

If you would like to start contributing to an issue or PR, please request to
have it assigned to yourself.

## Contributing

Check out our [contributing guidelines](docs/CONTRIBUTING.md).

## License

Unless otherwise noted, all code in the Lokomotive repository is licensed under
the [Apache 2.0 license](LICENSE).
Some portions of the codebase are derived from other projects under different
licenses; the appropriate information can be found in the header of those
source files, as applicable.
