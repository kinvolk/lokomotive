# lokoctl

lokoctl is a command line interface for Lokomotive, Kinvolk's open-source Kubernetes distribution
which includes installers for various platforms and Lokomotive components.

## Installation

### Download official releases

To install, find the [appropriate
package](https://github.com/kinvolk/lokomotive/releases) for your system and
download it. lokoctl is packaged as a `tar.gz` archive.

```console
wget <URL-TO-DOWNLOAD-LOKOCTL>
```
After downloading, untar the package. Lokoctl runs as a single binary named `lokoctl`.

```console
tar xvzf <LOKOCTL_TAR_GZ_ARCHIVE>
```

The final step is to make sure that the `lokoctl` binary is available on the `PATH`.

```console
export PATH=$PATH:/path/to/lokoctl/binary
```

### Using 'go get'

You can quickly get latest version of `lokoctl` by running following command:
```
go get github.com/kinvolk/lokomotive/cmd/lokoctl
```

Once finished, `lokoctl` binary should be available in `$GOPATH/bin`.

### Building from source

Clone this repository and build the lokoctl binary:

```bash
git clone https://github.com/kinvolk/lokomotive
cd lokomotive
make
```

The binary will be created in the project main directory.

Run `lokoctl help` to get an overview of all available commands.
