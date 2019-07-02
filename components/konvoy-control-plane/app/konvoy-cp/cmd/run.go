package cmd

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/bootstrap"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/server"
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
		grpcPort        int
		httpPort        int
		diagnosticsPort int
		metricsPort     int
	}{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Control Plane",
		Long:  `Launch Control Plane.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := bootstrap.Bootstrap(config.TODO())
			if err != nil {
				runLog.Error(err, "unable to set up Control Plane runtime")
				return err
			}
			server := &server.Server{
				Args: server.RunArgs{
					GrpcPort:        args.grpcPort,
					HttpPort:        args.httpPort,
					DiagnosticsPort: args.diagnosticsPort,
				}}
			if err := server.Setup(rt); err != nil {
				runLog.Error(err, "unable to set up xDS API server")
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
	cmd.PersistentFlags().IntVar(&args.grpcPort, "grpc-port", 5678, "port to run gRPC xDS API server on")
	cmd.PersistentFlags().IntVar(&args.httpPort, "http-port", 5679, "port to run HTTP xDS API server on")
	cmd.PersistentFlags().IntVar(&args.diagnosticsPort, "diagnostics-port", 5680, "port to run diagnostics server on")
	cmd.PersistentFlags().IntVar(&args.metricsPort, "metrics-port", 5681, "port to run metrics server on")
	return cmd
}
