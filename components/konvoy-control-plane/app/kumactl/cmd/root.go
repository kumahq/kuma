package cmd

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/cmd/apply"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/cmd/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/cmd/get"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/cmd/inspect"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/cmd/install"
	kumactl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/cmd/version"
	"github.com/spf13/cobra"
	"os"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
)

var (
	kumactlLog = core.Log.WithName("kumactl")
)

// newRootCmd represents the base command when called without any subcommands.
func NewRootCmd(root *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kumactl",
		Short: "Management tool for Kuma",
		Long:  `Management tool for Kuma.`,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			core.SetLogger(core.NewLogger(root.Args.Debug))

			return root.LoadConfig()
		},
	}
	// root flags
	cmd.PersistentFlags().StringVar(&root.Args.ConfigFile, "config-file", "", "path to the configuration file to use")
	cmd.PersistentFlags().StringVar(&root.Args.Mesh, "mesh", "", "mesh to use")
	cmd.PersistentFlags().BoolVar(&root.Args.Debug, "debug", true, "enable debug-level logging")
	// sub-commands
	cmd.AddCommand(install.NewInstallCmd(root))
	cmd.AddCommand(config.NewConfigCmd(root))
	cmd.AddCommand(get.NewGetCmd(root))
	cmd.AddCommand(inspect.NewInspectCmd(root))
	cmd.AddCommand(apply.NewApplyCmd(root))
	cmd.AddCommand(version.NewVersionCmd())
	return cmd
}

func DefaultRootCmd() *cobra.Command {
	return NewRootCmd(kumactl_cmd.DefaultRootContext())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := DefaultRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
