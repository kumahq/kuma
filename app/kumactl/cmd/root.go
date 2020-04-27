package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/cmd/apply"
	"github.com/Kong/kuma/app/kumactl/cmd/config"
	"github.com/Kong/kuma/app/kumactl/cmd/delete"
	"github.com/Kong/kuma/app/kumactl/cmd/generate"
	"github.com/Kong/kuma/app/kumactl/cmd/get"
	"github.com/Kong/kuma/app/kumactl/cmd/inspect"
	"github.com/Kong/kuma/app/kumactl/cmd/install"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	kumactl_config "github.com/Kong/kuma/app/kumactl/pkg/config"
	kumactl_errors "github.com/Kong/kuma/app/kumactl/pkg/errors"
	kuma_cmd "github.com/Kong/kuma/pkg/cmd"
	"github.com/Kong/kuma/pkg/cmd/version"
	"github.com/Kong/kuma/pkg/core"
	kuma_log "github.com/Kong/kuma/pkg/log"
)

// newRootCmd represents the base command when called without any subcommands.
func NewRootCmd(root *kumactl_cmd.RootContext) *cobra.Command {
	args := struct {
		logLevel string
	}{}
	cmd := &cobra.Command{
		Use:   "kumactl",
		Short: "Management tool for Kuma",
		Long:  `Management tool for Kuma.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			level, err := kuma_log.ParseLogLevel(args.logLevel)
			if err != nil {
				return err
			}
			core.SetLogger(core.NewLogger(level))

			// once command line flags have been parsed,
			// avoid printing usage instructions
			cmd.SilenceUsage = true

			if root.IsFirstTimeUsage() {
				root.Runtime.Config = kumactl_config.DefaultConfiguration()
				if err := root.SaveConfig(); err != nil {
					return err
				}
			}
			return root.LoadConfig()
		},
	}
	// root flags
	cmd.PersistentFlags().StringVar(&root.Args.ConfigFile, "config-file", "", "path to the configuration file to use")
	cmd.PersistentFlags().StringVarP(&root.Args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.PersistentFlags().StringVar(&args.logLevel, "log-level", kuma_log.OffLevel.String(), kuma_cmd.UsageOptions("log level", kuma_log.OffLevel, kuma_log.InfoLevel, kuma_log.DebugLevel))
	// sub-commands
	cmd.AddCommand(install.NewInstallCmd(root))
	cmd.AddCommand(config.NewConfigCmd(root))
	cmd.AddCommand(get.NewGetCmd(root))
	cmd.AddCommand(delete.NewDeleteCmd(root))
	cmd.AddCommand(inspect.NewInspectCmd(root))
	cmd.AddCommand(apply.NewApplyCmd(root))
	cmd.AddCommand(version.NewVersionCmd())
	cmd.AddCommand(generate.NewGenerateCmd(root))
	kumactl_cmd.WrapRunnables(cmd, kumactl_errors.FormatErrorWrapper)
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
