package cmd

import (
	"github.com/spf13/cobra"
)

var healthCheckLog = testServerLog.WithName("health-check")

func newHealthCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health-check",
		Short: "Run Test Server for Health Check test",
		Long:  "Run Test Server for Health Check test.",
	}
	cmd.AddCommand(newHealthCheckTCP())
	cmd.AddCommand(newHealthCheckHTTP())
	return cmd
}
