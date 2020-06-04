# Prometheus Operator CRDs configuration reference for Lokomotive

## Contents

* [Introduction](#introduction)
* [Prerequisites](#prerequisites)
* [Configuration](#configuration)
* [Attribute reference](#attribute-reference)
* [Applying](#applying)
* [Deleting](#deleting)

## Introduction

This component provides CRDs for [prometheus-operator](prometheus-operator.md) component, so they can be used
before actual prometheus-operator itself is installed. That allows enabling monitoring for storage solution, which
might be a dependency for prometheus-operator component, so avoiding circular dependency.

## Prerequisites

* A Lokomotive cluster.

## Configuration

Prometheus Operator CRDs component configuration example:

```tf
component "prometheus-operator-crds" {}
```

## Attribute reference

Table of all the arguments accepted by the component.

Example:

| Argument | Description | Default | Required |
|--------	|--------------|:-------:|:--------:|

## Applying

To apply the Prometheus Operator component:

```bash
lokoctl component apply prometheus-operator-crds
```

### Post-installation

After installation, you can proceed with installing [prometheus-operator](prometheus-operator.md) itself.

## Deleting

To destroy the component:

```bash
lokoctl component delete prometheus-operator-crds --delete-namespace
```
