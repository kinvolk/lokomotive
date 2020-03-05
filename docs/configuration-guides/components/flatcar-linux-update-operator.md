# Flatcar Container Linux update operator configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Argument reference](#argument-reference)
* [Installing](#installing)
* [Uninstalling](#uninstalling)

## Introduction

This component is a controller that manages node reboots for nodes running Flatcar Container Linux
images. When a reboot is needed after updating the system via
[update_engine](https://github.com/coreos/update_engine), the operator will drain the node before
rebooting it.


## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

## Configuration

This component does not require any specific configuration.

An empty configuration block is also accepted as valid configuration.

Flatcar Container Linux update operator component configuration example:

```tf
component "flatcar-linux-update-operator" {}
```

In some cases, you would want to prevent a certain node from rebooting by the operator. To do that:

```bash
kubectl label nodes NODENAME flatcar-linux-update.v1.flatcar-linux.net/reboot-pause=true
```

For more details visit the [Flatcar Container Linux update operator GitHub
repository](https://github.com/kinvolk/flatcar-linux-update-operator).

## Argument reference

This component does not accept any arguments in its configuration.

## Installing

To install the Flatcar Container Linux update operator component:

```bash
lokoctl component install flatcar-linux-update-operator
```

This component is installed in the `reboot-coordinator` namespace.

## Uninstalling

To uninstall the component:

```bash
lokoctl component render-manifest flatcar-linux-update-operator | kubectl delete -f -
```
