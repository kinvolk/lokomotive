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

## Update the lokomotive-kubernetes to current master

To update the local git submodule `lokomotive-kubernetes` dir to the latest master run following commands:

```bash
cd $GOPATH/src/github.com/kinvolk/lokoctl/lokomotive-kubernetes/
git pull --ff-only origin master
cd ..
```

Now commit those changes and send a PR.
