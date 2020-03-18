# Calico HostEndpoint controller configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Destroying](#destroying)

## Introduction

A host endpoint resource (HostEndpoint) in Calico represents one or more real or virtual interfaces
attached to a host that is running Calico. It enforces Calico policy on the traffic that is entering
or leaving the host’s default network namespace through those interfaces.

This component makes sure new nodes get Calico HostEndpoint objects when they're created and those
objects get removed when nodes they refer to are deleted.

This is relevant for Lokomotive clusters in bare-metal or Packet because there are no external
security primitives and nodes must rely on HostEndpoint objects to be secured.


## Prerequisites

* A Lokomotive cluster accessible via `kubectl` deployed on Packet.

* Calico as the CNI plugin.

## Configuration

This component does not require any specific configuration.

An empty configuration block is also accepted as valid configuration.

Calico HostEndpoint controller component configuration example:

```tf
component "calico-hostendpoint-controller" {}
```

## Attribute reference

This component does not accept any arguments in its configuration.

## Applying

To install the Calico HostEndpoint controller component:

```bash
lokoctl component apply calico-hostendpoint-controller
```

This component is installed in the `kube-system` namespace.

## Destroying

To destroy the component:

```bash
lokoctl component render-manifest calico-hostendpoint-controller | kubectl delete -f -
```
