package cmd

import (
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/documentation/examples/custom-sd/adapter"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kuma-prometheus-sd/pkg/discovery/xds"
	"github.com/kumahq/kuma/app/kuma-prometheus-sd/pkg/discovery/xds/common"
	util_log "github.com/kumahq/kuma/app/kuma-prometheus-sd/pkg/util/go-kit/log"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/config"
	kuma_promsd "github.com/kumahq/kuma/pkg/config/app/kuma-prometheus-sd"
	util_os "github.com/kumahq/kuma/pkg/util/os"
)

var (
	runLog = prometheusSdLog.WithName("run")
)

func newRunCmdWithOpts(opts kuma_cmd.RunCmdOpts) *cobra.Command {
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

			outputDir, _ := filepath.Split(cfg.Prometheus.OutputFile)
			if err := util_os.TryWriteToDir(outputDir); err != nil {
				return errors.Wrapf(err, "unable to write to directory %q", outputDir)
			}

			ctx := opts.SetupSignalHandler()

			discoverer, err := xds.NewDiscoverer(
				common.DiscoveryConfig{
					ServerURL:  cfg.MonitoringAssignment.Client.URL,
					ClientName: cfg.MonitoringAssignment.Client.Name,
					ApiVersion: common.ApiVersion(cfg.MonitoringAssignment.Client.ApiVersion),
				},
				runLog.WithName("xds_sd").WithName("discoverer"),
			)
			if err != nil {
				runLog.Error(err, "unable to set up xDS discoverer")
				return err
			}
			discovery := adapter.NewAdapter(ctx, cfg.Prometheus.OutputFile, "xds_sd", discoverer, util_log.NewLogger(runLog.WithName("xds_sd"), "adapter"))
			discovery.Run()

			<-ctx.Done()
			return nil
		},
	}
	// flags
	cmd.PersistentFlags().StringVar(&cfg.MonitoringAssignment.Client.URL, "cp-address", cfg.MonitoringAssignment.Client.URL, "URL of the Control Plane Monitoring Assignment Discovery Server. Example: grpc://localhost:5676")
	cmd.PersistentFlags().StringVar(&cfg.MonitoringAssignment.Client.Name, "name", cfg.MonitoringAssignment.Client.Name, "Name to use to identify itself to the Monitoring Assignment server.")
	cmd.PersistentFlags().StringVar(&cfg.MonitoringAssignment.Client.ApiVersion, "api-version", cfg.MonitoringAssignment.Client.ApiVersion, "MADS API version to request from the Monitoring Assignment server.")
	cmd.PersistentFlags().StringVar(&cfg.Prometheus.OutputFile, "output-file", cfg.Prometheus.OutputFile, "Path to an output file with a list of scrape targets. The same file path must be used on Prometheus side in a configuration of `file_sd` discovery mechanism.")
	return cmd
}
