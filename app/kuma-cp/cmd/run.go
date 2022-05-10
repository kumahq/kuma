package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/clusterid"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/bootstrap"
	"github.com/kumahq/kuma/pkg/defaults"
	"github.com/kumahq/kuma/pkg/diagnostics"
	dp_server "github.com/kumahq/kuma/pkg/dp-server"
	"github.com/kumahq/kuma/pkg/gc"
	"github.com/kumahq/kuma/pkg/hds"
	"github.com/kumahq/kuma/pkg/insights"
	kds_global "github.com/kumahq/kuma/pkg/kds/global"
	kds_zone "github.com/kumahq/kuma/pkg/kds/zone"
	mads_server "github.com/kumahq/kuma/pkg/mads/server"
	metrics "github.com/kumahq/kuma/pkg/metrics/components"
	"github.com/kumahq/kuma/pkg/util/os"
	kuma_version "github.com/kumahq/kuma/pkg/version"
	"github.com/kumahq/kuma/pkg/xds"
)

var (
	runLog = controlPlaneLog.WithName("run")
)

const gracefullyShutdownDuration = 3 * time.Second

// This is the open file limit below which the control plane may not
// reasonably have enough descriptors to accept all its clients.
const minOpenFileLimit = 4096

func newRunCmdWithOpts(opts kuma_cmd.RunCmdOpts) *cobra.Command {
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

			kuma_cp.PrintDeprecations(&cfg, cmd.OutOrStdout())

			gracefulCtx, ctx := opts.SetupSignalHandler()
			rt, err := bootstrap.Bootstrap(gracefulCtx, cfg)
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

			if err := os.RaiseFileLimit(); err != nil {
				runLog.Error(err, "unable to raise the open file limit")
			}

			if limit, _ := os.CurrentFileLimit(); limit < minOpenFileLimit {
				runLog.Info("for better performance, raise the open file limit",
					"minimim-open-files", minOpenFileLimit)
			}

			if err := mads_server.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up Monitoring Assignment server")
				return err
			}
			if err := xds.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up XDS")
				return err
			}
			if err := hds.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up HDS")
				return err
			}
			if err := dp_server.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up DP Server")
				return err
			}
			if err := insights.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up Insights resyncer")
				return err
			}
			if err := defaults.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up Defaults")
				return err
			}
			if err := kds_zone.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up Zone KDS")
				return err
			}
			if err := kds_global.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up Global KDS")
				return err
			}
			if err := clusterid.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up clusterID")
				return err
			}
			if err := diagnostics.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up Diagnostics server")
				return err
			}
			if err := api_server.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up API server")
				return err
			}
			if err := metrics.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up Metrics")
				return err
			}
			if err := gc.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up GC")
				return err
			}

			runLog.Info("starting Control Plane", "version", kuma_version.Build.Version)
			if err := rt.Start(gracefulCtx.Done()); err != nil {
				runLog.Error(err, "problem running Control Plane")
				return err
			}

			runLog.Info("Stop signal received. Waiting 3 seconds for components to stop gracefully...")
			select {
			case <-ctx.Done():
			case <-time.After(gracefullyShutdownDuration):
			}
			runLog.Info("Stopping Control Plane")
			return nil
		},
	}
	// flags
	cmd.PersistentFlags().StringVarP(&args.configPath, "config-file", "c", "", "configuration file")
	return cmd
}
