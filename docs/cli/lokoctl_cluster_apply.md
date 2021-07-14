---
title: lokoctl cluster apply
weight: 10
---

Deploy or update a cluster

### Synopsis

Deploy or update a cluster.
Deploys a cluster if it isn't deployed, otherwise updates it.
Unless explicitly skipped, components listed in the configuration are applied as well.

```
lokoctl cluster apply [flags]
```

### Options

```
      --confirm                        Upgrade cluster without asking for confirmation
  -h, --help                           help for apply
      --skip-components                Skip applying component configuration
      --skip-control-plane-update      Skip updating the control plane (not recommended)
      --skip-pre-update-health-check   Skip ensuring that cluster is healthy before updating (not recommended)
      --upgrade-kubelets               Upgrade self-hosted kubelets (default true)
  -v, --verbose                        Show output from Terraform
```

### Options inherited from parent commands

```
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl cluster](lokoctl_cluster.md)	 - Manage a cluster

