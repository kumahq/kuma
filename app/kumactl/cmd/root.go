package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/kumahq/kuma/app/kumactl/cmd/apply"
	"github.com/kumahq/kuma/app/kumactl/cmd/completion"
	"github.com/kumahq/kuma/app/kumactl/cmd/config"
	"github.com/kumahq/kuma/app/kumactl/cmd/delete"
	"github.com/kumahq/kuma/app/kumactl/cmd/export"
	"github.com/kumahq/kuma/app/kumactl/cmd/generate"
	"github.com/kumahq/kuma/app/kumactl/cmd/get"
	"github.com/kumahq/kuma/app/kumactl/cmd/inspect"
	"github.com/kumahq/kuma/app/kumactl/cmd/install"
	"github.com/kumahq/kuma/app/kumactl/cmd/uninstall"
	"github.com/kumahq/kuma/app/kumactl/cmd/version"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	kumactl_config "github.com/kumahq/kuma/app/kumactl/pkg/config"
	kumactl_errors "github.com/kumahq/kuma/app/kumactl/pkg/errors"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/plugins"
	core_apis "github.com/kumahq/kuma/pkg/core/resources/apis"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/plugins/policies"
	// Register gateway resources.
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
	// import Envoy protobuf definitions so (un)marshaling Envoy protobuf works
	_ "github.com/kumahq/kuma/pkg/xds/envoy"
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
			// The transparent proxy does not rely on any of the features below.
			// Since it handles its own flag parsing, processing the parent flags
			// like `--config-file` is unnecessary and would add unnecessary
			// complexity
			if cmd.Name() == "transparent-proxy" {
				return nil
			}

			level, err := kuma_log.ParseLogLevel(args.logLevel)
			if err != nil {
				return err
			}
			l := core.NewLogger(level)
			core.SetLogger(l)
			// Required for any k8s stuff that may log.
			klog.SetLogger(l)

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

			if root.Args.ConfigType == kumactl_cmd.InMemory {
				root.LoadInMemoryConfig()
				return nil
			}

			return root.LoadConfig()
		},
	}

	cmd.SetOut(os.Stdout)

	// root flags
	cmd.PersistentFlags().StringVar(&root.Args.ConfigFile, "config-file", "", "path to the configuration file to use")
	cmd.PersistentFlags().StringVar(&args.logLevel, "log-level", kuma_log.OffLevel.String(), kuma_cmd.UsageOptions("log level", kuma_log.OffLevel, kuma_log.InfoLevel, kuma_log.DebugLevel))
	cmd.PersistentFlags().BoolVar(&args.noConfig, "no-config", false, "if set no config file and config directory will be created")
	cmd.PersistentFlags().DurationVar(&root.Args.ApiTimeout, "api-timeout", time.Minute, "the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout")

	// sub-commands
	cmd.AddCommand(apply.NewApplyCmd(root))
	cmd.AddCommand(export.NewExportCmd(root))
	cmd.AddCommand(completion.NewCompletionCommand())
	cmd.AddCommand(config.NewConfigCmd(root))
	cmd.AddCommand(delete.NewDeleteCmd(root))
	cmd.AddCommand(generate.NewGenerateCmd(root))
	cmd.AddCommand(get.NewGetCmd(root))
	cmd.AddCommand(inspect.NewInspectCmd(root))
	cmd.AddCommand(install.NewInstallCmd(root))
	cmd.AddCommand(uninstall.NewUninstallCmd(root))
	cmd.AddCommand(version.NewCmd(root))

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

func init() {
	plugins.InitAll(core_apis.NameToModule)
	plugins.InitAll(policies.NameToModule)
}
