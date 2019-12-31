package cmd

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Kong/kuma/pkg/catalog/client"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumadp_config "github.com/Kong/kuma/app/kuma-dp/pkg/config"
	"github.com/Kong/kuma/app/kuma-dp/pkg/dataplane/accesslogs"
	"github.com/Kong/kuma/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/Kong/kuma/pkg/config"
	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/Kong/kuma/pkg/config/types"
	"github.com/Kong/kuma/pkg/core"
	util_net "github.com/Kong/kuma/pkg/util/net"
)

type CatalogClientFactory func(string) (client.CatalogClient, error)

var (
	runLog = dataplaneLog.WithName("run")
	// overridable by tests
	bootstrapGenerator   = envoy.NewRemoteBootstrapGenerator(&http.Client{Timeout: 10 * time.Second})
	catalogClientFactory = client.NewCatalogClient
)

func newRunCmd() *cobra.Command {
	cfg := kuma_dp.DefaultConfig()
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Dataplane (Envoy)",
		Long:  `Launch Dataplane (Envoy).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// only support configuration via environment variables and args
			if err := config.Load("", &cfg); err != nil {
				runLog.Error(err, "unable to load configuration")
				return err
			}
			if conf, err := config.ToYAML(&cfg); err == nil {
				runLog.Info("effective configuration", "config", string(conf))
			} else {
				runLog.Error(err, "unable to format effective configuration", "config", cfg)
				return err
			}

			catalogClient, err := catalogClientFactory(cfg.ControlPlane.ApiServer.URL)
			if err != nil {
				return errors.Wrap(err, "could not create catalog client")
			}
			catalog, err := catalogClient.Catalog()
			if err != nil {
				return errors.Wrap(err, "could retrieve catalog")
			}
			if catalog.Apis.DataplaneToken.Enabled() {
				if cfg.DataplaneRuntime.TokenPath == "" {
					return errors.New("Kuma CP is configured with Dataplane Token Server therefore the Dataplane Token is required. " +
						"Generate token using 'kumactl generate dataplane-token > /path/file' and provide it via --dataplane-token-file=/path/file argument to Kuma DP")
				}
				if err := kumadp_config.ValidateTokenPath(cfg.DataplaneRuntime.TokenPath); err != nil {
					return err
				}
			}

			if !cfg.Dataplane.AdminPort.Empty() {
				// unless a user has explicitly opted out of Envoy Admin API, pick a free port from the range
				adminPort, err := util_net.PickTCPPort("127.0.0.1", cfg.Dataplane.AdminPort.Lowest(), cfg.Dataplane.AdminPort.Highest())
				if err != nil {
					return errors.Wrapf(err, "unable to find a free port in the range %q for Envoy Admin API to listen on", cfg.Dataplane.AdminPort)
				}
				cfg.Dataplane.AdminPort = config_types.MustExactPort(adminPort)
				runLog.Info("picked a free port for Envoy Admin API to listen on", "port", cfg.Dataplane.AdminPort)
			}

			if cfg.DataplaneRuntime.ConfigDir == "" {
				tmpDir, err := ioutil.TempDir("", "kuma-dp-")
				if err != nil {
					runLog.Error(err, "unable to create a temporary directory to store generated Envoy config at")
					return err
				}
				defer func() {
					if err := os.RemoveAll(tmpDir); err != nil {
						runLog.Error(err, "unable to remove a temporary directory with a generated Envoy config")
					}
				}()
				cfg.DataplaneRuntime.ConfigDir = tmpDir
				runLog.Info("generated Envoy configuration will be stored in a temporary directory", "dir", tmpDir)
			}

			runLog.Info("starting Dataplane (Envoy) ...")

			dataplane := envoy.New(envoy.Opts{
				Catalog:   catalog,
				Config:    cfg,
				Generator: bootstrapGenerator,
				Stdout:    cmd.OutOrStdout(),
				Stderr:    cmd.OutOrStderr(),
			})

			server := accesslogs.NewAccessLogServer()
			defer server.Close()

			logServerErr := make(chan error)
			go func() {
				defer close(logServerErr)
				if err := server.Start(cfg.Dataplane); err != nil {
					runLog.Error(err, "problem running Access Log server")
					logServerErr <- err
				}
				runLog.Info("stopped Access Log server")
			}()

			dataplaneErr := make(chan error)
			go func() {
				defer close(dataplaneErr)
				if err := dataplane.Run(core.SetupSignalHandler()); err != nil {
					runLog.Error(err, "problem running Dataplane (Envoy)")
					dataplaneErr <- err
				}
				runLog.Info("stopped Dataplane (Envoy)")
			}()

			select {
			case err := <-logServerErr:
				if err == nil {
					return errors.New("Access Log server terminated unexpectedly")
				}
				return err
			case err := <-dataplaneErr:
				return err
			}
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Name, "name", cfg.Dataplane.Name, "Name of the Dataplane")
	cmd.PersistentFlags().Var(&cfg.Dataplane.AdminPort, "admin-port", `Port (or range of ports to choose from) for Envoy Admin API to listen on. Empty value indicates that Envoy Admin API should not be exposed over TCP. Format: "9901 | 9901-9999 | 9901- | -9901"`)
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Mesh, "mesh", cfg.Dataplane.Mesh, "Mesh that Dataplane belongs to")
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.ApiServer.URL, "cp-address", cfg.ControlPlane.ApiServer.URL, "URL of the Control Plane API Server")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.BinaryPath, "binary-path", cfg.DataplaneRuntime.BinaryPath, "Binary path of Envoy executable")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.ConfigDir, "config-dir", cfg.DataplaneRuntime.ConfigDir, "Directory in which Envoy config will be generated")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.TokenPath, "dataplane-token-file", cfg.DataplaneRuntime.TokenPath, "Path to a file with dataplane token (use 'kumactl generate dataplane-token' to get one)")
	return cmd
}
