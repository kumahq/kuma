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
	cmd.AddCommand(newInstallMetrics(pctx))
	cmd.AddCommand(newInstallTracing())
	cmd.AddCommand(newInstallDNS())
	cmd.AddCommand(newInstallLogging())
	cmd.AddCommand(newInstallTransparentProxy())
	return cmd
}
