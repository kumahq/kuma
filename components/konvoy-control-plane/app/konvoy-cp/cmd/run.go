package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/server"
)

var (
	runArgs = struct {
		grpcPort int
		httpPort int
	}{}

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launch Control Plane",
		Long:  `Launch Control Plane.`,
		RunE:  func(cmd *cobra.Command, args []string) error {
			return server.Run(server.RunArgs{
				GrpcPort: runArgs.grpcPort,
				HttpPort: runArgs.httpPort,
			})
		},
	}
)

func init() {
	runCmd.PersistentFlags().IntVar(&runArgs.grpcPort, "grpc-port", 5678, "gRPC port to run xDS API server on")
	runCmd.PersistentFlags().IntVar(&runArgs.httpPort, "http-port", 5679, "HTTP port to run xDS API server on")

	rootCmd.AddCommand(runCmd)
}
