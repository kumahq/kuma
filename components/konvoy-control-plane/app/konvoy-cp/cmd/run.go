package cmd

import (
	api_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	konvoy_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/bootstrap"
	sds_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/sds/server"
	xds_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/server"
	"github.com/spf13/cobra"
)

var (
	runLog = controlPlaneLog.WithName("run")
)

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
			cfg := konvoy_cp.DefaultConfig()
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
			if err := sds_server.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up SDS server")
				return err
			}
			if err := xds_server.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up xDS server")
				return err
			}
			if err := api_server.SetupServer(rt); err != nil {
				runLog.Error(err, "unable to set up API server")
				return err
			}

			runLog.Info("starting Control Plane")
			if err := rt.Start(opts.SetupSignalHandler()); err != nil {
				runLog.Error(err, "problem running Control Plane")
				return err
			}

			runLog.Info("stopping Control Plane")
			return nil
		},
	}
	// flags
	cmd.PersistentFlags().StringVarP(&args.configPath, "config-file", "c", "", "configuration file")
	return cmd
}
