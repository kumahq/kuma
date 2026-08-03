package cmd

import "github.com/spf13/cobra"

var grpcLog = testServerLog.WithName("grpc")

func newGRPCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grpc",
		Short: "GRPC testing",
		Long:  "GRPC testing.",
	}
	cmd.AddCommand(newGRPCServerCmd())
	cmd.AddCommand(newGRPCClientCmd())
	return cmd
}
