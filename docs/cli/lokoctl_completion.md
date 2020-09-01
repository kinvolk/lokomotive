## lokoctl completion

Generate the completion code for the specified shell

### Synopsis

  Generate the completion code for lokoctl for the specified shell (Bash or zsh).


### Examples

```
  # Load the lokoctl completion code for Bash into the current shell.
  source <(lokoctl completion bash)

  # Load the lokoctl completion code for zsh into the current shell.
  source <(lokoctl completion zsh)

  # Generate a Bash completion file and load it for every shell.
  lokoctl completion bash > ~/.bash_lokoctl_completion
  echo "source ~/.bash_lokoctl_completion" >> ~/.bashrc && source ~/.bashrc

  # Set the lokoctl completion code for zsh to autoload on startup.
  lokoctl completion zsh > "${fpath[1]}/_lokoctl" && exec $SHELL
```

### Options

```
  -h, --help   help for completion
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
* [lokoctl completion bash](lokoctl_completion_bash.md)	 - Generate the completion code for Bash
* [lokoctl completion zsh](lokoctl_completion_zsh.md)	 - Generate the completion code for zsh

