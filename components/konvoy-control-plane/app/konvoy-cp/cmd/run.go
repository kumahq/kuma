package cmd

import (
	"fmt"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/server"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	runLog = controlPlaneLog.WithName("run")

	scheme = runtime.NewScheme()
)

var (
	runArgs = struct {
		grpcPort        int
		httpPort        int
		diagnosticsPort int
		metricsPort     int
	}{}

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launch Control Plane",
		Long:  `Launch Control Plane.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			mgr, err := ctrl.NewManager(
				ctrl.GetConfigOrDie(),
				ctrl.Options{Scheme: scheme, MetricsBindAddress: fmt.Sprintf(":%d", runArgs.metricsPort)},
			)
			if err != nil {
				runLog.Error(err, "unable to set up Control Plane")
				return err
			}

			server := &server.Server{
				Args: server.RunArgs{
					GrpcPort:        runArgs.grpcPort,
					HttpPort:        runArgs.httpPort,
					DiagnosticsPort: runArgs.diagnosticsPort,
				}}

			if err := server.SetupWithManager(mgr); err != nil {
				runLog.Error(err, "unable to set up xDS API server")
				return err
			}

			runLog.Info("starting Control Plane")
			if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
				runLog.Error(err, "problem running Control Plane")
				return err
			}

			return nil
		},
	}
)

func init() {
	runCmd.PersistentFlags().IntVar(&runArgs.grpcPort, "grpc-port", 5678, "port to run gRPC xDS API server on")
	runCmd.PersistentFlags().IntVar(&runArgs.httpPort, "http-port", 5679, "port to run HTTP xDS API server on")
	runCmd.PersistentFlags().IntVar(&runArgs.diagnosticsPort, "diagnostics-port", 5680, "port to run diagnostics server on")
	runCmd.PersistentFlags().IntVar(&runArgs.metricsPort, "metrics-port", 5681, "port to run metrics server on")

	rootCmd.AddCommand(runCmd)
}
