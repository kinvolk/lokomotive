## lokoctl

Manage Lokomotive clusters

### Synopsis

Manage Lokomotive clusters

### Options

```
  -h, --help                     help for lokoctl
      --kubeconfig-file string   Path to a kubeconfig file. If empty, the following precedence order is used:
                                   1. Cluster asset dir when a lokocfg file is present in the current directory.
                                   2. KUBECONFIG environment variable.
                                   3. ~/.kube/config file.
      --lokocfg string           Path to lokocfg directory or file (default "./")
      --lokocfg-vars string      Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl cluster](lokoctl_cluster.md)	 - Manage a cluster
* [lokoctl component](lokoctl_component.md)	 - Manage components
* [lokoctl health](lokoctl_health.md)	 - Get the health of a cluster
* [lokoctl version](lokoctl_version.md)	 - Print version information

