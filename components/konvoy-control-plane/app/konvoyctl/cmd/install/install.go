package install

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/spf13/cobra"
)

func NewInstallCmd(pctx *cmd.RootContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "install",
		Short: "Install Konvoy on Kubernetes",
		Long:  `Install Konvoy on Kubernetes.`,
	}
	// sub-commands
	command.AddCommand(newInstallControlPlaneCmd(pctx))
	return command
}
