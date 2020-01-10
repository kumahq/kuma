package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/pkg/core"
)

var (
	runLog = prometheusSdLog.WithName("run")
)

var (
	// overridable by unit tests
	setupSignalHandler = core.SetupSignalHandler
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Kuma Prometheus SD adapter",
		Long:  `Launch Kuma Prometheus SD adapter.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			runLog.Info("not implemented yet")
			<-setupSignalHandler()
			return nil
		},
	}
	return cmd
}
