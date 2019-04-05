# Contribution Guidelines

## Setup developer environment

```bash
mkdir -p $GOPATH/src/github.com/kinvolk
cd $GOPATH/src/github.com/kinvolk
git clone git@github.com:kinvolk/lokoctl.git
cd lokoctl
```

## Build the code

```bash
cd $GOPATH/src/github.com/kinvolk/lokoctl
make
```

To build a "dev" version of lokoctl, use

```
make update-lk-submodule
make build-slim
```

The resulting binary won't include the Lokomotive Kubernetes assets and
requires the lokomotive-kubernetes code in the submodule directory.
