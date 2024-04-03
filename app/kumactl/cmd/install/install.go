package install

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

func NewInstallCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install various Kuma components.",
		Long:  `Install various Kuma components.`,
	}

	// sub-commands
	cmd.AddCommand(newInstallControlPlaneCmd(&pctx.InstallCpContext))
	cmd.AddCommand(newInstallCrdsCmd(&pctx.InstallCRDContext))
	cmd.AddCommand(newInstallObservability(pctx))
	cmd.AddCommand(newInstallDemoCmd(&pctx.InstallDemoContext))
	cmd.AddCommand(newInstallTransparentProxy())

	return cmd
}
