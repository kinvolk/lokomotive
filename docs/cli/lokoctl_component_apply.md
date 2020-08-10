---
title: lokoctl component apply
weight: 10
---

Deploy or update a component

### Synopsis

Deploy or update a component.
Deploys a component if not yet present, otherwise updates it.
When run with no arguments, all components listed in the configuration are applied.

```
lokoctl component apply [flags]
```

### Options

```
      --debug                    Print debug messages
  -h, --help                     help for apply
      --kubeconfig-file string   Path to a kubeconfig file. If empty, the following precedence order is used:
                                   1. Cluster asset dir when a lokocfg file is present in the current directory.
                                   2. KUBECONFIG environment variable.
                                   3. ~/.kube/config file.
```

### Options inherited from parent commands

```
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl component](lokoctl_component.md)	 - Manage components

