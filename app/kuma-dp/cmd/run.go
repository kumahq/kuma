package cmd

import (
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/Kong/kuma/pkg/config"
	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	"github.com/Kong/kuma/pkg/core"
)

var (
	runLog = dataplaneLog.WithName("run")
	// overridable by tests
	bootstrapGenerator = envoy.NewRemoteBootstrapGenerator(&http.Client{Timeout: 10 * time.Second})
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

			runLog.Info("starting Dataplane (Envoy) ...")

			dataplane := envoy.New(envoy.Opts{
				Config:    cfg,
				Generator: bootstrapGenerator,
				Stdout:    cmd.OutOrStdout(),
				Stderr:    cmd.OutOrStderr(),
			})
			if err := dataplane.Run(core.SetupSignalHandler()); err != nil {
				runLog.Error(err, "problem running Dataplane (Envoy)")
				return err
			}

			runLog.Info("stopped Dataplane (Envoy)")
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Name, "name", cfg.Dataplane.Name, "Name of the Dataplane")
	cmd.PersistentFlags().Uint32Var(&cfg.Dataplane.AdminPort, "admin-port", cfg.Dataplane.AdminPort, "Port for Envoy Admin")
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Mesh, "mesh", cfg.Dataplane.Mesh, "Mesh that Dataplane belongs to")
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.BootstrapServer.URL, "cp-address", cfg.ControlPlane.BootstrapServer.URL, "Mesh that Dataplane belongs to")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.BinaryPath, "binary-path", cfg.DataplaneRuntime.BinaryPath, "Binary path of Envoy executable")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.ConfigDir, "config-dir", cfg.DataplaneRuntime.ConfigDir, "Directory in which Envoy config will be generated")
	return cmd
}
