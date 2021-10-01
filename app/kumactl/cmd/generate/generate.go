package generate

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/cli"
)

func NewGenerateCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate resources, tokens, etc",
		Long:  `Generate resources, tokens, etc.`,
	}
	generateCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := kumactl_cmd.RunParentPreRunE(generateCmd, args); err != nil {
			return err
		}
		if err := pctx.CheckServerVersionCompatibility(); err != nil {
			cmd.PrintErrln(err)
		}
		return nil
	}
	// sub-commands
	generateCmd.AddCommand(NewGenerateDataplaneTokenCmd(pctx))
	generateCmd.AddCommand(NewGenerateZoneIngressTokenCmd(pctx))
	generateCmd.AddCommand(NewGenerateCertificateCmd(pctx))
	generateCmd.AddCommand(NewGenerateSigningKeyCmd(pctx))
	generateCmd.AddCommand(cli.NewGenerateUserTokenCmd(pctx))
	return generateCmd
}
