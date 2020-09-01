## lokoctl completion bash

Generate the completion code for Bash

### Synopsis

  Generate the completion code for lokoctl for the Bash shell.


```
lokoctl completion bash
```

### Examples

```
  # If running Bash 3.2 that is included with macOS, install Bash completion using Homebrew.
  brew install bash-completion
	
  # If running Bash 4.1+ on macOS, install Bash completion using homebrew.
  brew install bash-completion@2

  # Load the lokoctl completion code for Bash into the current shell.
  source <(lokoctl completion bash)

  # Generate a Bash completion file and load it for every shell.
  lokoctl completion bash > ~/.bash_lokoctl_completion
  echo "source ~/.bash_lokoctl_completion" >> ~/.bashrc && source ~/.bashrc

```

### Options

```
  -h, --help   help for bash
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

* [lokoctl completion](lokoctl_completion.md)	 - Generate the completion code for the specified shell

