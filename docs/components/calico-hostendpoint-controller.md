# Calico HostEndpoint controller

This component makes sure new nodes get Calico HostEndpoint objects when
they're created and those objects get removed when nodes they refer to are
deleted.

This is relevant for bare-metal or Packet clusters because there are no
external security primitives and nodes must rely on HostEndpoint objects to be
secured.

## Installation

```
lokoctl component install calico-hostendpoint-controller
```

This component runs in the `kube-system` namespace.

## Next steps

Once the Calico HostEndpoint controller is installed, adding or removing nodes
manually or via the Cluster Autoscaler results in HostEndpoint objects being
added or removed.
