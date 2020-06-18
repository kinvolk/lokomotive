## lokoctl component

Manage components

### Synopsis

Manage components

### Options

```
  -h, --help   help for component
```

### Options inherited from parent commands

```
      --kubeconfig-file string   Path to a kubeconfig file. If empty, the following precedence order is used: 1. cluster asset dir when a lokocfg file is present in the current directory 2. KUBECONFIG environment variable 3. "~/.kube/config"
      --lokocfg string           Path to lokocfg directory or file (default "./")
      --lokocfg-vars string      Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl](lokoctl.md)	 - Manage Lokomotive clusters
* [lokoctl component apply](lokoctl_component_apply.md)	 - Deploy or update a component
* [lokoctl component delete](lokoctl_component_delete.md)	 - Delete an installed component
* [lokoctl component list](lokoctl_component_list.md)	 - List all available components
* [lokoctl component render-manifest](lokoctl_component_render-manifest.md)	 - Print the manifests for a component

