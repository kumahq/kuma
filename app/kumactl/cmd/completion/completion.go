package completion

import (
	"github.com/spf13/cobra"
)

const completionLong = `
Outputs kumactl shell completion for the given shell (bash, fish or zsh).
This depends on the bash-completion package.  Example installation instructions:
# for bash users
	$ kumactl completion bash > ~/.kumactl-completion
	$ source ~/.kumactl-completion

# for zsh users
	% kumactl completion zsh > /usr/local/share/zsh/site-functions/_kumactl
	% autoload -U compinit && compinit
# or if zsh-completion is installed via homebrew
    % kumactl completion zsh > "${fpath[1]}/_kumactl"

# for fish users
	% kumactl completion fish > ~/.config/fish/completions/kumactl.fish

Additionally, you may want to output the completion to a file and source in your .bashrc
Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2
`

func NewCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Output shell completion code for bash, fish or zsh",
		Long:  completionLong,
	}
	cmd.AddCommand(newBashCommand())
	cmd.AddCommand(newFishCommand())
	cmd.AddCommand(newZshCommand())
	return cmd
}

func newBashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Output shell completions for bash",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Parent().Parent().GenBashCompletion(cmd.OutOrStdout())
		},
	}
	return cmd
}

func newFishCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fish",
		Short: "Output shell completions for fish",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Parent().Parent().GenFishCompletion(cmd.OutOrStdout(), true)
		},
	}
	return cmd
}

func newZshCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "Output shell completions for zsh",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Parent().Parent().GenZshCompletion(cmd.OutOrStdout())
		},
	}
	return cmd
}
