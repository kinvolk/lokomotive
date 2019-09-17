# lokoctl

lokoctl is a command line interface for Lokomotive, Kinvolk's open-source
Kubernetes distribution which includes installers for various platforms and
add-ons (Lokomotive Components) usable from any Kubernetes cluster.

### Supported Platforms

* [AWS](/docs/installer/aws.md)
* [Baremetal](/docs/installer/baremetal.md)
* [Packet](/docs/installer/packet.md)

### Components

Lokomotive Components are add-ons to the core Kubernetes installation that add
extra functionality.

A list of all available components can be get with `lokoctl component list`. Documentation for components can be found in [docs/components](docs/components/).

* [MetalLB](docs/components/metallb.md)
* [Contour](docs/components/contour.md)
* [Cluster Autoscaler](docs/components/cluster-autoscaler.md)

## Installation

### Using 'go get'

You can quickly get latest version of `lokoctl` by running following command:
```
go get github.com/kinvolk/lokoctl
```

Once finished, `lokoctl` binary should be available in `$GOPATH/bin`.

**NOTE: As `lokoctl` repository is currently private, following additional git configuration
is needed to make `go get` to work.**

Add following snippet to your `~/.gitconfig` file:
```
[url "ssh://git@github.com/kinvolk/lokoctl"]
  insteadOf = "https://github.com/kinvolk/lokoctl"
```

### Building from source

Clone this repository and build the lokoctl binary:

```bash
git clone https://github.com/kinvolk/lokoctl
cd $_
make
```

Run `lokoctl help` to get an overview of all available commands.

## Setting up a cluster

Detailed installation guides for all supported platforms can be found
in [docs/installer](docs/installer).

## Contributing

Please read the [contribution guidelines](./docs/CONTRIBUTING.md).
