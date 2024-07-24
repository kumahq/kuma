package generate

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
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
		_ = kumactl_cmd.CheckCompatibility(pctx.FetchServerVersion, cmd.ErrOrStderr())
		return nil
	}
	// sub-commands
	generateCmd.AddCommand(NewGenerateDataplaneTokenCmd(pctx))
	generateCmd.AddCommand(NewGenerateZoneTokenCmd(pctx))
	generateCmd.AddCommand(NewGenerateCertificateCmd(pctx))
	generateCmd.AddCommand(NewGenerateSigningKeyCmd(pctx))
	generateCmd.AddCommand(NewGeneratePublicKeyCmd(pctx))
	generateCmd.AddCommand(NewGenerateUserTokenCmd(pctx))
	return generateCmd
}
