# lokoctl

lokoctl is a command line interface for Lokomotive, Kinvolk's open-source Kubernetes distribution which includes installers for various platforms and add-ons (Lokomotive Components) usable from any Kubernetes cluster.

### Supported Platforms

* [AWS](/docs/installer/aws.md)
* [Baremetal](/docs/installer/baremetal.md)

### Components
Lokomotive Components are add-ons to the core Kubernetes installation that add extra functionality.

* [Cert Manager](/manifests/cert-manager/README.md)
* [Network Policy](/manifests/default-network-policies/README.md)
* [Ingress Nginx](/manifests/ingress-nginx/README.md)

If you want to add a new component that does not already exist or modify one, you can find instructions [here](/manifests/README.md).

## Quick Start

Clone this repository and build.
```bash
git clone https://github.com/kinvolk/lokoctl $GOPATH/src/github.com/kinvolk/lokoctl
cd $_
make
```
This will create `lokoctl`. This binary can then be invoked to get help and view other commands, like this:
```
./lokoctl --help
```

### Install a cluster
You can install a Kubernetes cluster on a [given platform](#Supported-Platforms) like this:
```
$ ./lokoctl install <platform> <flags>
...
...
Apply complete! Resources: 59 added, 0 changed, 0 destroyed.

Node                 Ready    Reason          Message                      
    
node1.example.com    True     KubeletReady    kubelet is posting ready status
node2.example.com    True     KubeletReady    kubelet is posting ready status   

Your configurations are stored in /path/to/storage/directory
```

### Install a component
You can install a [component](#Components) on a Kubernetes cluster like this:
```
$ ./lokoctl component install <component> <flags>

Waiting for api-server...
Creating assets...
Created ExamplePolicy namespace/component
```

### Cleanup
TBD

## Contributing
Please read the [contribution guidelines](/docs/dev.md).