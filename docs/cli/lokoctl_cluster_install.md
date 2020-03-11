## lokoctl cluster install

Install Lokomotive cluster with components

### Synopsis

Install Lokomotive cluster with components

```
lokoctl cluster install [flags]
```

### Options

```
      --confirm           Upgrade cluster without asking for confirmation
  -h, --help              help for install
      --skip-components   Skip component installation
  -v, --verbose           Show output from Terraform
```

### Options inherited from parent commands

```
      --kubeconfig string     Path to kubeconfig file, taken from the asset dir if not given, and finally falls back to ~/.kube/config
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl cluster](lokoctl_cluster.md)	 - Manage a Lokomotive cluster

