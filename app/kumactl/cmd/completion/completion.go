package completion

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	kuma_version "github.com/kumahq/kuma/pkg/version"
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

func NewCompletionCommand(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Output shell completion code for bash, fish or zsh",
		Long:  completionLong,
	}
	cmd.AddCommand(newBashCommand(pctx))
	cmd.AddCommand(newFishCommand(pctx))
	cmd.AddCommand(newZshCommand(pctx))
	return cmd
}

func newBashCommand(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Output shell completions for bash",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Parent().Parent().GenBashCompletion(cmd.OutOrStdout())
		},
	}
	return cmd
}

func newFishCommand(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fish",
		Short: "Output shell completions for fish",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Parent().Parent().GenFishCompletion(cmd.OutOrStdout(), true)
		},
	}
	return cmd
}

func newZshCommand(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "Output shell completions for zsh",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Parent().Parent().GenZshCompletion(cmd.OutOrStdout())
		},
	}
	return cmd
}

func NewGenManCommand() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "genman",
		Short: "Generate man pages for the kuma CLI",
		Long:  `This command automatically generates up-to-date man pages of kuma's command-line interface.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Root().DisableAutoGenTag = true
			header := &doc.GenManHeader{
				Section: "1",
				Manual:  "Kuma CLI Manual",
				Source:  fmt.Sprintf("kumactl %s", kuma_version.Build.Version),
			}
			err := doc.GenManTree(cmd.Root(), header, path)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "output-dir", "o", "/tmp/", "directory to populate with documentation")
	return cmd
}
