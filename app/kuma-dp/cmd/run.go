package cmd

import (
	"github.com/Kong/kuma/pkg/catalogue/client"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumadp_config "github.com/Kong/kuma/app/kuma-dp/pkg/config"
	"github.com/Kong/kuma/app/kuma-dp/pkg/dataplane/accesslogs"
	"github.com/Kong/kuma/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/Kong/kuma/pkg/config"
	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	"github.com/Kong/kuma/pkg/core"
)

type CatalogueClientFactory func(string) (client.CatalogueClient, error)

var (
	runLog = dataplaneLog.WithName("run")
	// overridable by tests
	bootstrapGenerator     = envoy.NewRemoteBootstrapGenerator(&http.Client{Timeout: 10 * time.Second})
	catalogueClientFactory = client.NewCatalogueClient
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

			catalogueClient, err := catalogueClientFactory(cfg.ControlPlane.ApiServer.URL)
			if err != nil {
				return errors.Wrap(err, "could not create catalogue client")
			}
			catalogue, err := catalogueClient.Catalogue()
			if err != nil {
				return errors.Wrap(err, "could retrieve catalogue")
			}
			if err := kumadp_config.ValidateTokenPath(cfg.DataplaneRuntime.TokenPath); err != nil {
				return err
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
				Catalogue: catalogue,
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
	cmd.PersistentFlags().Uint32Var(&cfg.Dataplane.AdminPort, "admin-port", cfg.Dataplane.AdminPort, "Port for Envoy Admin")
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Mesh, "mesh", cfg.Dataplane.Mesh, "Mesh that Dataplane belongs to")
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.ApiServer.URL, "cp-address", cfg.ControlPlane.ApiServer.URL, "URL of the Control Plane API Server")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.BinaryPath, "binary-path", cfg.DataplaneRuntime.BinaryPath, "Binary path of Envoy executable")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.ConfigDir, "config-dir", cfg.DataplaneRuntime.ConfigDir, "Directory in which Envoy config will be generated")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.TokenPath, "dataplane-token", cfg.DataplaneRuntime.TokenPath, "Path to a file with dataplane token (use 'kumactl generate dataplane-token' to get one)")
	return cmd
}
