package cmd

import (
	"github.com/spf13/cobra"
)

func newInstallCmd(pctx *rootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Konvoy on Kubernetes",
		Long:  `Install Konvoy on Kubernetes.`,
	}
	// sub-commands
	cmd.AddCommand(newInstallControlPlaneCmd(pctx))
	return cmd
}
