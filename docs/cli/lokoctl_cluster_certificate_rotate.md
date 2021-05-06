---
title: lokoctl cluster certificate rotate
weight: 10
---

Rotate certificates of a cluster

### Synopsis

Rotate certificates of a cluster.
Rotate will replace all certificates inside a cluster with new ones.
This can be used to renew all certificates with a longer validity.

```
lokoctl cluster certificate rotate [flags]
```

### Options

```
  -h, --help                           help for rotate
      --skip-pre-update-health-check   Skip ensuring that cluster is healthy before updating (not recommended)
  -v, --verbose                        Show output from Terraform
```

### Options inherited from parent commands

```
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl cluster certificate](lokoctl_cluster_certificate.md)	 - Manage cluster certificates

