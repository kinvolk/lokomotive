## lokoctl cluster

Manage a cluster

### Synopsis

Manage a cluster

### Options

```
  -h, --help   help for cluster
```

### Options inherited from parent commands

```
      --kubeconfig-file string   Path to a kubeconfig file. If empty, the following precedence order is used:
                                   1. Cluster asset dir when a lokocfg file is present in the current directory.
                                   2. KUBECONFIG environment variable.
                                   3. ~/.kube/config file.
      --lokocfg string           Path to lokocfg directory or file (default "./")
      --lokocfg-vars string      Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl](lokoctl.md)	 - Manage Lokomotive clusters
* [lokoctl cluster apply](lokoctl_cluster_apply.md)	 - Deploy or update a cluster
* [lokoctl cluster destroy](lokoctl_cluster_destroy.md)	 - Destroy a cluster

