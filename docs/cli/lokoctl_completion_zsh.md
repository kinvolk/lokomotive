---
title: lokoctl completion zsh
weight: 10
---

Generate the completion code for zsh

### Synopsis

  Generate the completion code for lokoctl for the zsh shell.


```
lokoctl completion zsh
```

### Examples

```
  # Load the lokoctl completion code for zsh into the current shell.
  source <(lokoctl completion zsh)

  # Set the lokoctl completion code for zsh to autoload on startup.
  lokoctl completion zsh > "${fpath[1]}/_lokoctl" && exec $SHELL

```

### Options

```
  -h, --help   help for zsh
```

### Options inherited from parent commands

```
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl completion](lokoctl_completion.md)	 - Generate the completion code for the specified shell

