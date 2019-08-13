package cmd

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/apply"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/get"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/install"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/spf13/cobra"
	"os"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
)

var (
	konvoyCtlLog = core.Log.WithName("konvoyctl")
)

// newRootCmd represents the base command when called without any subcommands.
func NewRootCmd(root *cmd.RootContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "konvoyctl",
		Short: "Management tool for Konvoy Service Mesh",
		Long:  `Management tool for Konvoy Service Mesh.`,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			core.SetLogger(core.NewLogger(root.Args.Debug))

			return root.LoadConfig()
		},
	}
	// root flags
	command.PersistentFlags().StringVar(&root.Args.ConfigFile, "config-file", "", "path to the configuration file to use")
	command.PersistentFlags().StringVar(&root.Args.Mesh, "mesh", "", "mesh to use")
	command.PersistentFlags().BoolVar(&root.Args.Debug, "debug", true, "enable debug-level logging")
	// sub-commands
	command.AddCommand(install.NewInstallCmd(root))
	command.AddCommand(config.NewConfigCmd(root))
	command.AddCommand(get.NewGetCmd(root))
	command.AddCommand(apply.NewApplyCmd(root))
	return command
}

func DefaultRootCmd() *cobra.Command {
	return NewRootCmd(cmd.DefaultRootContext())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := DefaultRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
