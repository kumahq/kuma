package cmd

import (
	"github.com/spf13/cobra"
)

var (
	runLog = dataplaneLog.WithName("run")
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Dataplane (Envoy)",
		Long:  `Launch Dataplane (Envoy).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			runLog.Info("starting Dataplane (Envoy) ...")

			runLog.Info("stopped Dataplane (Envoy)")
			return nil
		},
	}
	return cmd
}
