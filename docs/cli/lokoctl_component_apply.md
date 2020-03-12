## lokoctl component apply

Apply a component configuration. If not present it will install it.
If ran with no arguments it will apply all components mentioned in the
configuration.

### Synopsis

Apply a component configuration. If not present it will install it.
If ran with no arguments it will apply all components mentioned in the
configuration.

```
lokoctl component apply [flags]
```

### Options

```
  -h, --help   help for apply
```

### Options inherited from parent commands

```
      --kubeconfig string     Path to kubeconfig file, taken from the asset dir if not given, and finally falls back to ~/.kube/config
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl component](lokoctl_component.md)	 - Manage Lokomotive components

