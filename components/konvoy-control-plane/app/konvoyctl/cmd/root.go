package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	konvoyctl_resources "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/resources"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
)

var (
	konvoyCtlLog = core.Log.WithName("konvoyctl")
)

type rootArgs struct {
	configFile string
	mesh       string
	debug      bool
}

type rootRuntime struct {
	config           config_proto.Configuration
	now              func() time.Time
	newResourceStore func(*config_proto.ControlPlane) (core_store.ResourceStore, error)
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
	cmd.PersistentFlags().StringVar(&root.args.mesh, "mesh", "", "mesh to use")
	cmd.PersistentFlags().BoolVar(&root.args.debug, "debug", true, "enable debug-level logging")
	// sub-commands
	cmd.AddCommand(newInstallCmd(root))
	cmd.AddCommand(newConfigCmd(root))
	cmd.AddCommand(newGetCmd(root))
	cmd.AddCommand(newApplyCmd(root))
	return cmd
}

func defaultRootCmd() *cobra.Command {
	return newRootCmd(defaultRootContext())
}

func defaultRootContext() *rootContext {
	return &rootContext{
		runtime: rootRuntime{
			now:              time.Now,
			newResourceStore: konvoyctl_resources.NewResourceStore,
		},
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := defaultRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
