// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

const completionDesc = `  Generate the completion code for lokoctl for the specified shell (Bash or zsh).
`

const completionExample = `  # Load the lokoctl completion code for Bash into the current shell.
  source <(lokoctl completion bash)

  # Load the lokoctl completion code for zsh into the current shell.
  source <(lokoctl completion zsh)

  # Generate a Bash completion file and load it for every shell.
  lokoctl completion bash > ~/.bash_lokoctl_completion
  echo "source ~/.bash_lokoctl_completion" >> ~/.bashrc && source ~/.bashrc

  # Set the lokoctl completion code for zsh to autoload on startup.
  lokoctl completion zsh > "${fpath[1]}/_lokoctl" && exec $SHELL`

const bashCompDesc = `  Generate the completion code for lokoctl for the Bash shell.
`

const bashExample = `  # If running Bash 3.2 that is included with macOS, install Bash completion using Homebrew.
  brew install bash-completion
	
  # If running Bash 4.1+ on macOS, install Bash completion using homebrew.
  brew install bash-completion@2

  # Load the lokoctl completion code for Bash into the current shell.
  source <(lokoctl completion bash)

  # Generate a Bash completion file and load it for every shell.
  lokoctl completion bash > ~/.bash_lokoctl_completion
  echo "source ~/.bash_lokoctl_completion" >> ~/.bashrc && source ~/.bashrc
`

const zshCompDesc = `  Generate the completion code for lokoctl for the zsh shell.
`

const zshExample = `  # Load the lokoctl completion code for zsh into the current shell.
  source <(lokoctl completion zsh)

  # Set the lokoctl completion code for zsh to autoload on startup.
  lokoctl completion zsh > "${fpath[1]}/_lokoctl" && exec $SHELL
`

func newCompletionCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "completion",
		Short:             "Generate the completion code for the specified shell",
		Long:              completionDesc,
		Example:           completionExample,
		Args:              noArgs,
		ValidArgsFunction: noCompletions, // Disable file completion.
	}

	bash := &cobra.Command{
		Use:                   "bash",
		Short:                 "Generate the completion code for Bash",
		Long:                  bashCompDesc,
		Example:               bashExample,
		Args:                  noArgs,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletionBash(out, cmd)
		},
	}

	zsh := &cobra.Command{
		Use:                   "zsh",
		Short:                 "Generate the completion code for zsh",
		Long:                  zshCompDesc,
		Example:               zshExample,
		Args:                  noArgs,
		DisableFlagsInUseLine: true,
		ValidArgsFunction:     noCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletionZsh(out, cmd)
		},
	}

	cmd.AddCommand(bash, zsh)

	return cmd
}

func runCompletionBash(out io.Writer, cmd *cobra.Command) error {
	return cmd.Root().GenBashCompletion(out)
}

func runCompletionZsh(out io.Writer, cmd *cobra.Command) error {
	return cmd.Root().GenZshCompletion(out)
}

// noCompletions is used to disable file completion.
func noCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// noArgs returns an error if any args are included.
func noArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf(
			"%q accepts no arguments\n\nUsage:  %s",
			cmd.CommandPath(),
			cmd.UseLine(),
		)
	}

	return nil
}

func init() { //nolint:gochecknoinits
	RootCmd.AddCommand(newCompletionCmd(os.Stdout))
}
