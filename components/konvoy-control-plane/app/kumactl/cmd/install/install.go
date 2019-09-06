package install

import (
	kumactl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewInstallCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Kuma on Kubernetes",
		Long:  `Install Kuma on Kubernetes.`,
	}
	// sub-commands
	cmd.AddCommand(newInstallControlPlaneCmd(pctx))
	return cmd
}
