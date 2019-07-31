package cmd

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoy-injector/pkg/server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	konvoy_injector "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-injector"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"

	kube_config "sigs.k8s.io/controller-runtime/pkg/client/config"
	kube_manager "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/spf13/cobra"
)

var (
	runLog = injectorLog.WithName("run")
)

var (
	// overridable by unit tests
	setupSignalHandler = core.SetupSignalHandler
	getKubeConfig      = kube_config.GetConfig
)

func newRunCmd() *cobra.Command {
	args := struct {
		configPath string
	}{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Konvoy Sidecar injector",
		Long:  `Launch Konvoy Sidecar injector.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := konvoy_injector.DefaultConfig()
			err := config.Load(args.configPath, &cfg)
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
			config, err := getKubeConfig()
			if err != nil {
				runLog.Error(err, "unable to load Kubernetes config")
				return err
			}
			mgr, err := kube_manager.New(config, kube_manager.Options{})
			if err != nil {
				runLog.Error(err, "unable to create Kubernetes client")
				return err
			}
			if err := server.Setup(mgr, &cfg); err != nil {
				runLog.Error(err, "unable to set up Admission Web Hook server")
				return err
			}
			if err := mgr.Start(setupSignalHandler()); err != nil {
				runLog.Error(err, "problem running Konvoy Injector")
				return err
			}
			return nil
		},
	}
	// flags
	cmd.PersistentFlags().StringVarP(&args.configPath, "config-file", "c", "", "configuration file")
	return cmd
}
