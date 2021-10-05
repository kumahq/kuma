package install

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

func newInstallGatewayCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Install ingress gateway on Kubernetes",
		Long:  "Install ingress gateway on Kubernetes in its own namespace.",
	}
	// sub-commands
	cmd.AddCommand(newInstallGatewayKongCmd(&pctx.InstallGatewayKongContext))
	cmd.AddCommand(newInstallGatewayKongEnterpriseCmd(&pctx.InstallGatewayKongEnterpriseContext))
	return cmd
}
