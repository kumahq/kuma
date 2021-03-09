package cmd

import (
	"os"

	"github.com/spf13/cobra"

	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/cmd/version"
	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	_ "github.com/kumahq/kuma/pkg/xds/envoy" // import Envoy protobuf definitions so (un)marshalling Envoy protobuf works
)

var (
	controlPlaneLog = core.Log.WithName("kuma-cp")
)

// newRootCmd represents the base command when called without any subcommands.
func newRootCmd() *cobra.Command {
	args := struct {
		logLevel string
	}{}
	cmd := &cobra.Command{
		Use:   "kuma-cp",
		Short: "Universal Control Plane for Envoy-based Service Mesh",
		Long:  `Universal Control Plane for Envoy-based Service Mesh.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			level, err := kuma_log.ParseLogLevel(args.logLevel)
			if err != nil {
				return err
			}
			core.SetLogger(core.NewLogger(level))

			// once command line flags have been parsed,
			// avoid printing usage instructions
			cmd.SilenceUsage = true

			return nil
		},
	}
	// root flags
	cmd.PersistentFlags().StringVar(&args.logLevel, "log-level", kuma_log.InfoLevel.String(), kuma_cmd.UsageOptions("log level", kuma_log.OffLevel, kuma_log.InfoLevel, kuma_log.DebugLevel))
	// sub-commands
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newMigrateCmd())
	cmd.AddCommand(version.NewVersionCmd())
	return cmd
}

func DefaultRootCmd() *cobra.Command {
	return newRootCmd()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := DefaultRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
