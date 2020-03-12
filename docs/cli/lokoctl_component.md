## lokoctl component

Manage Lokomotive components

### Synopsis

Manage Lokomotive components

### Options

```
  -h, --help   help for component
```

### Options inherited from parent commands

```
      --kubeconfig string     Path to kubeconfig file, taken from the asset dir if not given, and finally falls back to ~/.kube/config
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl](lokoctl.md)	 - Manage Lokomotive clusters.
* [lokoctl component apply](lokoctl_component_apply.md)	 - Apply a component configuration. If not present it will install it.
If ran with no arguments it will apply all components mentioned in the
configuration.
* [lokoctl component list](lokoctl_component_list.md)	 - List all available components
* [lokoctl component render-manifest](lokoctl_component_render-manifest.md)	 - Render and print manifests for a component

