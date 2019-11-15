# Flatcar Container Linux Update Operator

This component is for a node reboot controller for Kubernetes running Flatcar
Container Linux images. When a reboot is needed after updating the system via
[update_engine](https://github.com/coreos/update_engine), the operator will
drain the node before rebooting it.

## Installation

```
lokoctl component install flatcar-linux-update-operator
```

This component runs in the `reboot-coordinator` namespace.

## Annotation to prevent reboots

In some cases, you would want to prevent a certain node from rebooting from the
operator. To do that:

```
kubectl label nodes NODENAME flatcar-linux-update.v1.flatcar-linux.net/reboot-pause=true
```

See the [Github repo](https://github.com/kinvolk/flatcar-linux-update-operator) for details.
