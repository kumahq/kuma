package cmd

import (
	"github.com/spf13/cobra"
	"net/http"
	"time"

	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	kuma_dp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/kuma-dp"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
)

var (
	runLog = dataplaneLog.WithName("run")
	// overridable by tests
	bootstrapGenerator = envoy.NewRemoteBootstrapGenerator(&http.Client{Timeout: 10 * time.Second})
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Dataplane (Envoy)",
		Long:  `Launch Dataplane (Envoy).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := kuma_dp.DefaultConfig()
			err := config.Load("", &cfg) // only support configuration via environment variables
			if err != nil {
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
	return cmd
}
