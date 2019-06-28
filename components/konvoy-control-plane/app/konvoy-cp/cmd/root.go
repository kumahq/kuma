package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
)

var (
	controlPlaneLog = core.Log.WithName("konvoy-cp")
)

// newRootCmd represents the base command when called without any subcommands.
func newRootCmd() *cobra.Command {
	args := struct {
		debug bool
	}{}
	cmd := &cobra.Command{
		Use:   "konvoy-control-plane",
		Short: "Universal Control Plane for Envoy-based Service Mesh",
		Long:  `Universal Control Plane for Envoy-based Service Mesh.`,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			core.SetLogger(core.NewLogger(args.debug))
		},
	}
	// root flags
	cmd.PersistentFlags().BoolVar(&args.debug, "debug", true, "enable debug-level logging")
	// sub-commands
	cmd.AddCommand(newRunCmd())
	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
