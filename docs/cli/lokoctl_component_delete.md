## lokoctl component delete

Delete an installed component

### Synopsis

Delete a component.
When run with no arguments, all components listed in the configuration are deleted.

```
lokoctl component delete [flags]
```

### Options

```
      --delete-namespace   Delete namespace with component.
  -h, --help               help for delete
```

### Options inherited from parent commands

```
      --kubeconfig string     Path to kubeconfig file, taken from the asset dir if not given, and finally falls back to ~/.kube/config
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl component](lokoctl_component.md)	 - Manage components

