package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/pkg/core"

	"github.com/Kong/kuma/pkg/config"
	kuma_promsd "github.com/Kong/kuma/pkg/config/app/kuma-prometheus-sd"
)

var (
	runLog = prometheusSdLog.WithName("run")
)

var (
	// overridable by unit tests
	setupSignalHandler = core.SetupSignalHandler
)

func newRunCmd() *cobra.Command {
	cfg := kuma_promsd.DefaultConfig()
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Kuma Prometheus SD adapter",
		Long:  `Launch Kuma Prometheus SD adapter.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// only support configuration via environment variables and args
			if err := config.Load("", &cfg); err != nil {
				runLog.Error(err, "unable to load configuration")
				return err
			}
			if conf, err := config.ToJson(&cfg); err == nil {
				runLog.Info("effective configuration", "config", string(conf))
			} else {
				runLog.Error(err, "unable to format effective configuration", "config", cfg)
				return err
			}

			runLog.Info("not implemented yet")
			<-setupSignalHandler()
			return nil
		},
	}
	// flags
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.ApiServer.URL, "cp-address", cfg.ControlPlane.ApiServer.URL, "URL of the Control Plane API Server")
	cmd.PersistentFlags().StringVar(&cfg.MonitoringAssignment.Client.Name, "name", cfg.MonitoringAssignment.Client.Name, "Name this adapter should use when connecting to Monitoring Assignment server.")
	cmd.PersistentFlags().StringVar(&cfg.Prometheus.OutputFile, "output-file", cfg.Prometheus.OutputFile, "Path to an output file with a list of scrape targets. The same file path must be used on Prometheus side in a configuration of `file_sd` discovery mechanism.")
	return cmd
}
