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

const zshInitialization = `#compdef lokoctl
__lokoctl_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__lokoctl_type() {
	# -t is not supported by zsh.
	if [ "$1" == "-t" ]; then
		shift
		# Fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time.
		if [ "$1" = "__lokoctl_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__lokoctl_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# Filter by given word as prefix.
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			# Use printf instead of echo because it is possible that
			# the value to print is -n, which would be interpreted
			# as a flag to echo.
			printf "%s\n" "${w}"
		fi
	done
}
__lokoctl_compopt() {
	true # Don't do anything. Not supported by bashcompinit in zsh.
}
__lokoctl_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items.
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__lokoctl_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__lokoctl_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# Somehow does not work. Maybe, zsh does not call this at all.
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__lokoctl_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__lokoctl_quote() {
	if [[ $1 == \'* || $1 == \"* ]]; then
		# Leave out first character.
		printf %q "${1:1}"
	else
		printf %q "$1"
	fi
}
autoload -U +X bashcompinit && bashcompinit
# Use word boundary patterns for BSD or GNU sed.
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q 'GNU\|BusyBox'; then
	LWORD='\<'
	RWORD='\>'
fi
__lokoctl_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__lokoctl_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__lokoctl_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__lokoctl_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__lokoctl_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__lokoctl_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__lokoctl_type/g" \
	-e 's/aliashash\["\(.\{1,\}\)"\]/aliashash[\1]/g' \
	-e 's/FUNCNAME/funcstack/g' \
	<<'BASH_COMPLETION_EOF'
`

const zshTail = `
BASH_COMPLETION_EOF
}
__lokoctl_bash_source <(__lokoctl_convert_bash_to_zsh)
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
	// TODO: Change the whole process to GenZshCompletion once
	// cobra releases their new updates.
	if _, err := out.Write([]byte(zshInitialization)); err != nil {
		return fmt.Errorf("writing zsh initialization: %w", err)
	}

	if err := runCompletionBash(out, cmd); err != nil {
		return fmt.Errorf("running Bash completion: %w", err)
	}

	if _, err := out.Write([]byte(zshTail)); err != nil {
		return fmt.Errorf("writing zsh tail: %w", err)
	}

	return nil
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
