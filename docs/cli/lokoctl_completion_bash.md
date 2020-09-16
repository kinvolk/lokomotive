---
title: lokoctl completion bash
weight: 10
---

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
      --lokocfg string        Path to lokocfg directory or file (default "./")
      --lokocfg-vars string   Path to lokocfg.vars file (default "./lokocfg.vars")
```

### SEE ALSO

* [lokoctl completion](lokoctl_completion.md)	 - Generate the completion code for the specified shell

