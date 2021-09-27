package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/cmd/version"
	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
)

var (
	dataplaneLog = core.Log.WithName("kuma-dp")
)

// NewRootCmd represents the base command when called without any subcommands.
func NewRootCmd(opts kuma_cmd.RunCmdOpts, rootCtx *RootContext) *cobra.Command {
	args := struct {
		logLevel   string
		outputPath string
		maxSize    int
		maxBackups int
		maxAge     int
	}{}
	cmd := &cobra.Command{
		Use:   "kuma-dp",
		Short: "Dataplane manager for Envoy-based Service Mesh",
		Long:  `Dataplane manager for Envoy-based Service Mesh.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			level, err := kuma_log.ParseLogLevel(args.logLevel)
			if err != nil {
				return err
			}

			if args.outputPath != "" {
				output, err := filepath.Abs(args.outputPath)
				if err != nil {
					return err
				}

				fmt.Printf("%s: logs will be stored in %q\n", "kuma-dp", output)
				core.SetLogger(core.NewLoggerWithRotation(level, output, args.maxSize, args.maxBackups, args.maxAge))
			} else {
				core.SetLogger(core.NewLogger(level))
			}
			rootCtx.LogLevel = level

			// once command line flags have been parsed,
			// avoid printing usage instructions
			cmd.SilenceUsage = true

			return nil
		},
	}

	cmd.SetOut(os.Stdout)

	// root flags
	cmd.PersistentFlags().StringVar(&args.logLevel, "log-level", kuma_log.InfoLevel.String(), kuma_cmd.UsageOptions("log level", kuma_log.OffLevel, kuma_log.InfoLevel, kuma_log.DebugLevel))
	cmd.PersistentFlags().StringVar(&args.outputPath, "log-output-path", args.outputPath, "path to the file that will be filled with logs. Example: if we set it to /tmp/kuma.log then after the file is rotated we will have /tmp/kuma-2021-06-07T09-15-18.265.log")
	cmd.PersistentFlags().IntVar(&args.maxBackups, "log-max-retained-files", 1000, "maximum number of the old log files to retain")
	cmd.PersistentFlags().IntVar(&args.maxSize, "log-max-size", 100, "maximum size in megabytes of a log file before it gets rotated")
	cmd.PersistentFlags().IntVar(&args.maxAge, "log-max-age", 30, "maximum number of days to retain old log files based on the timestamp encoded in their filename")

	// sub-commands
	cmd.AddCommand(newRunCmd(opts, rootCtx))
	cmd.AddCommand(version.NewVersionCmd())

	return cmd
}

func DefaultRootCmd() *cobra.Command {
	return NewRootCmd(kuma_cmd.DefaultRunCmdOpts, DefaultRootContext())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := DefaultRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
