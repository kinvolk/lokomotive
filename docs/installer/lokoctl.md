---
title: lokoctl
weight: 10
description: >
   lokoctl is a command line interface for Lokomotive, Kinvolk's open-source Kubernetes distribution which includes installers for various platforms
   and Lokomotive components.
---

## Installation

### Download official releases

Every release of Lokomotive provides the `lokoctl` binary for several operating systems.
These binaries can be manually downloaded and installed.

1. Download your [desired version](https://github.com/kinvolk/lokomotive/releases), including the GPG
   signature.

2. Verify the signature. It should match one of the [Trusted
   keys](https://github.com/kinvolk/lokomotive/blob/master/docs/KEYS.md).

```console
gpg --verify lokoctl_v0.9.0_linux_amd64.tar.gz.sig
```

3. Unpack it

```console
tar xvf lokoctl_v0.9.0_linux_amd64.tar.gz
```

4. Find the lokoctl binary in the unpacked directory and move it to its desired location

```console
mv lokoctl_v0.9.0_linux_amd64/lokoctl ~/.local/bin/lokoctl
```

5. Verify the version of `lokoctl`

```console
lokoctl version
v0.9.0
```

### Using 'go get'

You can quickly get latest version of `lokoctl` by running following command:
```console
go get github.com/kinvolk/lokomotive/cmd/lokoctl
```

Once finished, `lokoctl` binary should be available in `$GOPATH/bin`.

### Building from source

Clone this repository and build the lokoctl binary:

```console
git clone https://github.com/kinvolk/lokomotive
cd lokomotive
make
```

The binary will be created in the project main directory.

Run `lokoctl help` to get an overview of all available commands.
