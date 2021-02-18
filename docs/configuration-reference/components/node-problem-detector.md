# node-problem-detector configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

[node-problem-detector](https://github.com/kubernetes/node-problem-detector) aims to make various node problems visible to the upstream layers in the cluster management stack.

It is a daemon that runs on each node, detects node problems and reports them to apiserver.

## Prerequisites

* A Lokomotive cluster accessible via `kubectl`.

## Configuration

```tf
# node-problem-detector.lokocfg
component "node-problem-detector" {
  custom_monitors = [file("system-stats-monitor.json")]
}
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument          | Description                                                                                                                                                      | Default      | Type   | Required |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|--------|----------|
| `custom_monitors` | List of paths to system log monitor configuration files. [See](https://github.com/kubernetes/node-problem-detector/tree/master/config) for more custom monitors. | list(string) | string | false    |
| `service_monitor` | Specifies how metrics can be retrieved from a set of services.                                                                                                   | false        | bool   | false    |


## Applying

To apply the node-problem-detector component:

```bash
lokoctl component apply node-problem-detector
```

## Deleting

To destroy the component:

```bash
lokoctl component delete node-problem-detector
```
