package install

import (
	konvoyctl_ctx "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/context"
	"github.com/spf13/cobra"
)

func NewInstallCmd(pctx *konvoyctl_ctx.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Konvoy on Kubernetes",
		Long:  `Install Konvoy on Kubernetes.`,
	}
	// sub-commands
	cmd.AddCommand(newInstallControlPlaneCmd(pctx))
	return cmd
}
