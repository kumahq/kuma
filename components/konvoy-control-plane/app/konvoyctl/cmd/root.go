package cmd

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/apply"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/get"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/cmd/install"
	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/spf13/cobra"
	"os"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
)

var (
	konvoyCtlLog = core.Log.WithName("konvoyctl")
)

// newRootCmd represents the base command when called without any subcommands.
func NewRootCmd(root *konvoyctl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "konvoyctl",
		Short: "Management tool for Konvoy Service Mesh",
		Long:  `Management tool for Konvoy Service Mesh.`,
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
	cmd.AddCommand(apply.NewApplyCmd(root))
	return cmd
}

func DefaultRootCmd() *cobra.Command {
	return NewRootCmd(konvoyctl_cmd.DefaultRootContext())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := DefaultRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
