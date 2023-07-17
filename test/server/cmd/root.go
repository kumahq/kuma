package cmd

import (
	"os"

	"github.com/spf13/cobra"

	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
)

var testServerLog = core.Log.WithName("test-server")

func NewRootCmd() *cobra.Command {
	args := struct {
		logLevel string
		logFormat string
	}{}
	cmd := &cobra.Command{
		Use:   "test-server",
		Short: "Test Server for Kuma e2e testing",
		Long:  `Test Server for Kuma e2e testing.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			format, errFormat := kuma_log.ParseLogFormat(args.logFormat)
			level, errLevel := kuma_log.ParseLogLevel(args.logLevel)
			if errLevel != nil {
				return errLevel
			}
			
			if errFormat != nil {
				return errFormat
			}
			
			core.SetLogger(core.NewLogger(level, format))

			cmd.SilenceUsage = true
			return nil
		},
	}
	
	cmd.PersistentFlags().StringVar(&args.logLevel, "log-level", kuma_log.InfoLevel.String(), kuma_cmd.UsageOptions("log level", kuma_log.OffLevel, kuma_log.InfoLevel, kuma_log.DebugLevel))
	cmd.PersistentFlags().StringVar(&args.logFormat, "log-format", args.logFormat, "specify logformat, json | logfmt (default)")

	cmd.AddCommand(newHealthCheckCmd())
	cmd.AddCommand(newEchoHTTPCmd())
	cmd.AddCommand(newGRPCCmd())
	return cmd
}

func DefaultRootCmd() *cobra.Command {
	return NewRootCmd()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := DefaultRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
