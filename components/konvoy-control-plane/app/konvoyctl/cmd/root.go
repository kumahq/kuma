package cmd

import (
	"os"

	"github.com/spf13/cobra"

	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
)

var (
	konvoyCtlLog = core.Log.WithName("konvoyctl")
)

type rootArgs struct {
	configFile string
	debug      bool
}

type rootRuntime struct {
	config config_proto.Configuration
}

type rootContext struct {
	args    rootArgs
	runtime rootRuntime
}

// newRootCmd represents the base command when called without any subcommands.
func newRootCmd(root *rootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "konvoyctl",
		Short: "Management tool for Konvoy Service Mesh",
		Long:  `Management tool for Konvoy Service Mesh.`,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			core.SetLogger(core.NewLogger(root.args.debug))

			return root.LoadConfig()
		},
	}
	// root flags
	cmd.PersistentFlags().StringVar(&root.args.configFile, "config-file", "", "path to the configuration file to use")
	cmd.PersistentFlags().BoolVar(&root.args.debug, "debug", true, "enable debug-level logging")
	// sub-commands
	cmd.AddCommand(newConfigCmd(root))
	cmd.AddCommand(newGetCmd(root))
	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := newRootCmd(&rootContext{}).Execute(); err != nil {
		os.Exit(1)
	}
}
