package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd/apply"
	"github.com/kumahq/kuma/app/kumactl/cmd/completion"
	"github.com/kumahq/kuma/app/kumactl/cmd/config"
	"github.com/kumahq/kuma/app/kumactl/cmd/delete"
	"github.com/kumahq/kuma/app/kumactl/cmd/generate"
	"github.com/kumahq/kuma/app/kumactl/cmd/get"
	"github.com/kumahq/kuma/app/kumactl/cmd/inspect"
	"github.com/kumahq/kuma/app/kumactl/cmd/install"
	"github.com/kumahq/kuma/app/kumactl/cmd/uninstall"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	kumactl_config "github.com/kumahq/kuma/app/kumactl/pkg/config"
	kumactl_errors "github.com/kumahq/kuma/app/kumactl/pkg/errors"
	"github.com/kumahq/kuma/pkg/api-server/types"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/cmd/version"
	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"

	// Register gateway resources.
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
	kuma_version "github.com/kumahq/kuma/pkg/version"

	// import Envoy protobuf definitions so (un)marshaling Envoy protobuf works
	_ "github.com/kumahq/kuma/pkg/xds/envoy"
)

var (
	kumactlLog       = core.Log.WithName("kumactl")
	kumaBuildVersion *types.IndexResponse
)

// newRootCmd represents the base command when called without any subcommands.
func NewRootCmd(root *kumactl_cmd.RootContext) *cobra.Command {
	args := struct {
		logLevel string
		noConfig bool
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

			if args.noConfig {
				root.Runtime.Config = kumactl_config.DefaultConfiguration()
				return nil
			}

			if root.IsFirstTimeUsage() {
				root.Runtime.Config = kumactl_config.DefaultConfiguration()
				if err := root.SaveConfig(); err != nil {
					return err
				}
			}

			if err := root.LoadConfig(); err != nil {
				return err
			}

			client, err := root.CurrentApiClient()
			if err != nil {
				kumactlLog.Error(err, "Unable to get index client")
			} else {
				kumaBuildVersion, err = client.GetVersion(context.Background())
				if err != nil {
					kumactlLog.Error(err, "Unable to retrieve server version")
				}
			}

			if kumaBuildVersion == nil {
				cmd.PrintErr("WARNING: Unable to confirm the server supports this kumactl version\n")
			} else if kumaBuildVersion.Version != kuma_version.Build.Version || kumaBuildVersion.Tagline != kuma_version.Product {
				cmd.PrintErr("WARNING: You are using kumactl version " + kuma_version.Build.Version + " for " + kuma_version.Product + ", but the server returned version: " + kumaBuildVersion.Tagline + " " + kumaBuildVersion.Version + "\n")
			}
			return nil
		},
	}

	cmd.SetOut(os.Stdout)

	// root flags
	cmd.PersistentFlags().StringVar(&root.Args.ConfigFile, "config-file", "", "path to the configuration file to use")
	cmd.PersistentFlags().StringVarP(&root.Args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.PersistentFlags().StringVar(&args.logLevel, "log-level", kuma_log.OffLevel.String(), kuma_cmd.UsageOptions("log level", kuma_log.OffLevel, kuma_log.InfoLevel, kuma_log.DebugLevel))
	cmd.PersistentFlags().BoolVar(&args.noConfig, "no-config", false, "if set no config file and config directory will be created")

	// sub-commands
	cmd.AddCommand(apply.NewApplyCmd(root))
	cmd.AddCommand(completion.NewCompletionCommand(root))
	cmd.AddCommand(config.NewConfigCmd(root))
	cmd.AddCommand(delete.NewDeleteCmd(root))
	cmd.AddCommand(generate.NewGenerateCmd(root))
	cmd.AddCommand(get.NewGetCmd(root))
	cmd.AddCommand(inspect.NewInspectCmd(root))
	cmd.AddCommand(install.NewInstallCmd(root))
	cmd.AddCommand(uninstall.NewUninstallCmd())
	cmd.AddCommand(version.NewVersionCmd())

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
