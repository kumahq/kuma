package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	kumadp_config "github.com/kumahq/kuma/app/kuma-dp/pkg/config"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/accesslogs"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/kumahq/kuma/pkg/config"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	util_net "github.com/kumahq/kuma/pkg/util/net"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var runLog = dataplaneLog.WithName("run")

func newRunCmd(rootCtx *RootContext) *cobra.Command {
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
			if conf, err := config.ToJson(&cfg); err == nil {
				runLog.Info("effective configuration", "config", string(conf))
			} else {
				runLog.Error(err, "unable to format effective configuration", "config", cfg)
				return err
			}

			dp, err := readDataplaneResource(cmd, &cfg)
			if err != nil {
				runLog.Error(err, "unable to read provided dataplane")
				return err
			}
			if dp != nil {
				cfg.Dataplane.Mesh = dp.Meta.GetMesh()
				cfg.Dataplane.Name = dp.Meta.GetName()
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

			if cfg.DataplaneRuntime.Token != "" {
				path := filepath.Join(cfg.DataplaneRuntime.ConfigDir, cfg.Dataplane.Name)
				if err := writeFile(path, []byte(cfg.DataplaneRuntime.Token), 0600); err != nil {
					runLog.Error(err, "unable to create file with dataplane token")
					return err
				}
				cfg.DataplaneRuntime.TokenPath = path
			}

			if cfg.DataplaneRuntime.TokenPath != "" {
				if err := kumadp_config.ValidateTokenPath(cfg.DataplaneRuntime.TokenPath); err != nil {
					return err
				}
			}

			if cfg.ControlPlane.CaCert == "" && cfg.ControlPlane.CaCertFile != "" {
				cert, err := ioutil.ReadFile(cfg.ControlPlane.CaCertFile)
				if err != nil {
					return errors.Wrapf(err, "could not read certificate file %s", cfg.ControlPlane.CaCertFile)
				}
				cfg.ControlPlane.CaCert = string(cert)
			}

			shouldQuit := setupQuitChannel()

			dataplane, err := envoy.New(envoy.Opts{
				Config:    cfg,
				Generator: rootCtx.BootstrapGenerator,
				Dataplane: dp,
				Stdout:    cmd.OutOrStdout(),
				Stderr:    cmd.OutOrStderr(),
				Quit:      shouldQuit,
			})
			if err != nil {
				return err
			}
			server := accesslogs.NewAccessLogServer(cfg.Dataplane)

			if err := rootCtx.ComponentManager.Add(server, dataplane); err != nil {
				return err
			}

			runLog.Info("starting Kuma DP", "version", kuma_version.Build.Version)
			if err := rootCtx.ComponentManager.Start(shouldQuit); err != nil {
				runLog.Error(err, "error while running Kuma DP")
				return err
			}
			runLog.Info("stopping Kuma DP")
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Name, "name", cfg.Dataplane.Name, "Name of the Dataplane")
	cmd.PersistentFlags().Var(&cfg.Dataplane.AdminPort, "admin-port", `Port (or range of ports to choose from) for Envoy Admin API to listen on. Empty value indicates that Envoy Admin API should not be exposed over TCP. Format: "9901 | 9901-9999 | 9901- | -9901"`)
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Mesh, "mesh", cfg.Dataplane.Mesh, "Mesh that Dataplane belongs to")
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.URL, "cp-address", cfg.ControlPlane.URL, "URL of the Control Plane Dataplane Server. Example: https://localhost:5678")
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.CaCertFile, "ca-cert-file", cfg.ControlPlane.CaCert, "Path to CA cert by which connection to the Control Plane will be verified if HTTPS is used")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.BinaryPath, "binary-path", cfg.DataplaneRuntime.BinaryPath, "Binary path of Envoy executable")
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.BootstrapVersion, "bootstrap-version", cfg.Dataplane.BootstrapVersion, "Bootstrap version (and API version) of xDS config. If empty, default version defined in Kuma CP will be used. (ex. '2', '3')")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.ConfigDir, "config-dir", cfg.DataplaneRuntime.ConfigDir, "Directory in which Envoy config will be generated")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.TokenPath, "dataplane-token-file", cfg.DataplaneRuntime.TokenPath, "Path to a file with dataplane token (use 'kumactl generate dataplane-token' to get one)")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.Token, "dataplane-token", cfg.DataplaneRuntime.Token, "Dataplane Token")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.Resource, "dataplane", "", "Dataplane template to apply (YAML or JSON)")
	cmd.PersistentFlags().StringVarP(&cfg.DataplaneRuntime.ResourcePath, "dataplane-file", "d", "", "Path to Dataplane template to apply (YAML or JSON)")
	cmd.PersistentFlags().StringToStringVarP(&cfg.DataplaneRuntime.ResourceVars, "dataplane-var", "v", map[string]string{}, "Variables to replace Dataplane template")
	return cmd
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filename), perm); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, perm)
}

func setupQuitChannel() chan struct{} {
	quit := make(chan struct{})
	quitOnSignal := core.SetupSignalHandler()
	go func() {
		<-quitOnSignal
		runLog.Info("Kuma DP caught an exit signal")
		if quit != nil {
			close(quit)
		}
	}()

	return quit
}
