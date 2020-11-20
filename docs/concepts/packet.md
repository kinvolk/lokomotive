---
title: Packet
weight: 10
---

Packet is one of the cloud platforms supported by Lokomotive. This document explains various architecture decisions specific to this platform.

## Blocked access to metadata service

By default, access to Packet's [metadata service](https://www.packet.com/developers/docs/servers/key-features/metadata/) is blocked for all pods. This is to prevent possible exploitation of information provided by the metadata service such as user data, which may contain secrets.

To allow an application to access the metadata service, you can create a NetworkPolicy selecting the application.

Here's a simple NetworkPolicy that allows pods with the label `foo: foo` to send packets to any IP address including the metadata service.

For simplicity, this is a very open NetworkPolicy. You should consider creating a more restrictive one for your production clusters.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-metadata-access
spec:
  podSelector:
    matchLabels:
      foo: foo
  policyTypes:
  - Egress
  egress:
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
```

## Flatcar Linux Customization

Flatcar Container Linux deployments on Packet can be customized with Container Linux Configs.
For more information, see [Flatcar Container Linux Customization](/docs/concepts/flatcar-container-linux.md#Customization).