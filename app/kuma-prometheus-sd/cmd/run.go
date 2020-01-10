package cmd

import (
	"context"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/prometheus/prometheus/documentation/examples/custom-sd/adapter"

	catalog_client "github.com/Kong/kuma/pkg/catalog/client"
	"github.com/Kong/kuma/pkg/config"
	kuma_promsd "github.com/Kong/kuma/pkg/config/app/kuma-prometheus-sd"
	"github.com/Kong/kuma/pkg/core"

	"github.com/Kong/kuma/app/kuma-prometheus-sd/pkg/discovery/xds"
	util_log "github.com/Kong/kuma/app/kuma-prometheus-sd/pkg/util/go-kit/log"
	util_os "github.com/Kong/kuma/pkg/util/os"
)

var (
	runLog = prometheusSdLog.WithName("run")
)

type CatalogClientFactory func(string) (catalog_client.CatalogClient, error)

var (
	// overridable by unit tests
	setupSignalHandler   = core.SetupSignalHandler
	catalogClientFactory = catalog_client.NewCatalogClient
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

			outputDir, _ := filepath.Split(cfg.Prometheus.OutputFile)
			if err := util_os.TryWriteToDir(outputDir); err != nil {
				return errors.Wrapf(err, "unable to write to directory %q", outputDir)
			}

			catalogClient, err := catalogClientFactory(cfg.ControlPlane.ApiServer.URL)
			if err != nil {
				return errors.Wrap(err, "unable to create a client for Kuma API Catalog")
			}
			catalog, err := catalogClient.Catalog()
			if err != nil {
				return errors.Wrap(err, "unable to retrieve Catalog of Kuma APIs")
			}
			runLog.Info("fetched Catalog of Kuma APIs", "catalog", catalog)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			discoverer, err := xds.NewDiscoverer(
				xds.DiscoveryConfig{
					ServerURL:  catalog.Apis.MonitoringAssignment.Url,
					ClientName: cfg.MonitoringAssignment.Client.Name,
				},
				runLog.WithName("xds_sd").WithName("discoverer"),
			)
			if err != nil {
				runLog.Error(err, "unable to set up xDS discoverer")
				return err
			}
			discovery := adapter.NewAdapter(ctx, cfg.Prometheus.OutputFile, "xds_sd", discoverer, util_log.NewLogger(runLog.WithName("xds_sd"), "adapter"))
			discovery.Run()

			<-setupSignalHandler()
			return nil
		},
	}
	// flags
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.ApiServer.URL, "cp-address", cfg.ControlPlane.ApiServer.URL, "URL of the Control Plane API Server")
	cmd.PersistentFlags().StringVar(&cfg.MonitoringAssignment.Client.Name, "name", cfg.MonitoringAssignment.Client.Name, "Name to use to identify itself to the Monitoring Assignment server.")
	cmd.PersistentFlags().StringVar(&cfg.Prometheus.OutputFile, "output-file", cfg.Prometheus.OutputFile, "Path to an output file with a list of scrape targets. The same file path must be used on Prometheus side in a configuration of `file_sd` discovery mechanism.")
	return cmd
}
