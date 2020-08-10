---
title: lokoctl component delete
weight: 10
---

Delete an installed component

### Synopsis

Delete a component.
When run with no arguments, all components listed in the configuration are deleted.

```
lokoctl component delete [flags]
```

### Options

```
      --confirm                  Delete component without asking for confirmation
      --debug                    Print debug messages
      --delete-namespace         Delete namespace with component
  -h, --help                     help for delete
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

