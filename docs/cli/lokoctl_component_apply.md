## lokoctl component apply

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
  -h, --help   help for apply
```

### Options inherited from parent commands

```
      --kubeconfig-file string   Path to kubeconfig file, taken from the asset dir if not given and finally falls back to ~/.kube/config
      --lokocfg string           Path to lokocfg directory or file (default "./")
      --lokocfg-vars string      Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl component](lokoctl_component.md)	 - Manage components

