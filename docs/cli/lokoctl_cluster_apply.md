## lokoctl cluster apply

Apply configuration changes to a Lokomotive cluster with components

### Synopsis

Apply configuration changes to a Lokomotive cluster with components

```
lokoctl cluster apply [flags]
```

### Options

```
      --confirm            Upgrade cluster without asking for confirmation
  -h, --help               help for apply
      --skip-components    Skip applying component configuration
      --upgrade-kubelets   Experimentally upgrade self-hosted kubelets
  -v, --verbose            Show output from Terraform
```

### Options inherited from parent commands

```
      --kubeconfig string     Path to kubeconfig file, taken from the asset dir if not given, and finally falls back to ~/.kube/config
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl cluster](lokoctl_cluster.md)	 - Manage a Lokomotive cluster

