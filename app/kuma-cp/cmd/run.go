package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/clusterid"

	admin_server "github.com/kumahq/kuma/pkg/admin-server"
	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/bootstrap"
	"github.com/kumahq/kuma/pkg/defaults"
	"github.com/kumahq/kuma/pkg/diagnostics"
	"github.com/kumahq/kuma/pkg/dns/components"
	dp_server "github.com/kumahq/kuma/pkg/dp-server"
	"github.com/kumahq/kuma/pkg/gc"
	kds_global "github.com/kumahq/kuma/pkg/kds/global"
	kds_remote "github.com/kumahq/kuma/pkg/kds/remote"
	mads_server "github.com/kumahq/kuma/pkg/mads/server"
	metrics "github.com/kumahq/kuma/pkg/metrics/components"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var (
	runLog = controlPlaneLog.WithName("run")
)

const gracefullyShutdownDuration = 3 * time.Second

func newRunCmd() *cobra.Command {
	return newRunCmdWithOpts(runCmdOpts{
		SetupSignalHandler: core.SetupSignalHandler,
	})
}

type runCmdOpts struct {
	SetupSignalHandler func() (stopCh <-chan struct{})
}

func newRunCmdWithOpts(opts runCmdOpts) *cobra.Command {
	args := struct {
		configPath string
	}{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Control Plane",
		Long:  `Launch Control Plane.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := kuma_cp.DefaultConfig()
			err := config.Load(args.configPath, &cfg)
			if err != nil {
				runLog.Error(err, "could not load the configuration")
				return err
			}
			rt, err := bootstrap.Bootstrap(cfg)
			if err != nil {
				runLog.Error(err, "unable to set up Control Plane runtime")
				return err
			}
			cfgForDisplay, err := config.ConfigForDisplay(&cfg)
			if err != nil {
				runLog.Error(err, "unable to prepare config for display")
				return err
			}
			cfgBytes, err := config.ToJson(cfgForDisplay)
			if err != nil {
				runLog.Error(err, "unable to convert config to json")
				return err
			}
			runLog.Info(fmt.Sprintf("Current config %s", cfgBytes))
			runLog.Info(fmt.Sprintf("Running in mode `%s`", cfg.Mode))
			switch cfg.Mode {
			case config_core.Standalone:
				if err := mads_server.SetupServer(rt); err != nil {
					runLog.Error(err, "unable to set up Monitoring Assignment server")
					return err
				}
				if err := components.SetupServer(rt); err != nil {
					runLog.Error(err, "unable to set up DNS server")
					return err
				}
				if err := gc.Setup(rt); err != nil {
					runLog.Error(err, "unable to set up GC")
					return err
				}
				if err := clusterid.Setup(rt); err != nil {
					runLog.Error(err, "unable to set up clusterID")
					return err
				}
				if err := dp_server.SetupServer(rt); err != nil {
					runLog.Error(err, "unable to set up DP Server")
					return err
				}
			case config_core.Remote:
				if err := mads_server.SetupServer(rt); err != nil {
					runLog.Error(err, "unable to set up Monitoring Assignment server")
					return err
				}
				if err := kds_remote.Setup(rt); err != nil {
					runLog.Error(err, "unable to set up KDS Remote")
					return err
				}
				if err := components.SetupServer(rt); err != nil {
					runLog.Error(err, "unable to set up DNS server")
					return err
				}
				if err := gc.Setup(rt); err != nil {
					runLog.Error(err, "unable to set up GC")
					return err
				}
				if err := dp_server.SetupServer(rt); err != nil {
					runLog.Error(err, "unable to set up DP Server")
					return err
				}
			case config_core.Global:
				if err := kds_global.Setup(rt); err != nil {
					runLog.Error(err, "unable to set up KDS Global")
					return err
				}
				if err := clusterid.Setup(rt); err != nil {
					runLog.Error(err, "unable to set up clusterID")
					return err
				}
			}

			if err := diagnostics.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up Diagnostics server")
				return err
			}
			if err := api_server.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up API server")
				return err
			}
			if err := admin_server.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up Admin server")
				return err
			}

			if err := metrics.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up Metrics")
				return err
			}

			if err := defaults.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up Defaults")
				return err
			}

			runLog.Info("starting Control Plane", "version", kuma_version.Build.Version)
			if err := rt.Start(opts.SetupSignalHandler()); err != nil {
				runLog.Error(err, "problem running Control Plane")
				return err
			}

			runLog.Info("Stop signal received. Waiting 3 seconds for components to stop gracefully...")
			time.Sleep(gracefullyShutdownDuration)
			runLog.Info("Stopping Control Plane")
			return nil
		},
	}
	// flags
	cmd.PersistentFlags().StringVarP(&args.configPath, "config-file", "c", "", "configuration file")
	return cmd
}
