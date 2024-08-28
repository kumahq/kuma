package cmd

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/accesslogs"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/dnsserver"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/envoy"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/metrics"
	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/readiness"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/config"
	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/net"
	"github.com/kumahq/kuma/pkg/util/proto"
	kuma_version "github.com/kumahq/kuma/pkg/version"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var runLog = dataplaneLog.WithName("run")

// PersistentPreRunE in root command sets the logger and initial config
// PreRunE loads the Kuma DP config
// PostRunE actually runs all the components with loaded config
// To extend Kuma DP, plug your code in RunE. Use RootContext.Config and add components to RootContext.ComponentManager
func newRunCmd(opts kuma_cmd.RunCmdOpts, rootCtx *RootContext) *cobra.Command {
	cfg := rootCtx.Config
	var tmpDir string
	var proxyResource model.Resource
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launch Dataplane (Envoy)",
		Long:  `Launch Dataplane (Envoy).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			// only support configuration via environment variables and args
			if err := config.Load("", cfg); err != nil {
				runLog.Error(err, "unable to load configuration")
				return err
			}

			kumadp.PrintDeprecations(cfg, cmd.OutOrStdout())

			if conf, err := config.ToJson(cfg); err == nil {
				runLog.Info("effective configuration", "config", string(conf))
			} else {
				runLog.Error(err, "unable to format effective configuration", "config", cfg)
				return err
			}

			// Map the resource types that are acceptable depending on the value of the `--proxy-type` flag.
			proxyTypeMap := map[string]model.ResourceType{
				string(mesh_proto.DataplaneProxyType): mesh.DataplaneType,
				string(mesh_proto.IngressProxyType):   mesh.ZoneIngressType,
				string(mesh_proto.EgressProxyType):    mesh.ZoneEgressType,
			}

			if _, ok := proxyTypeMap[cfg.Dataplane.ProxyType]; !ok {
				return errors.Errorf("invalid proxy type %q", cfg.Dataplane.ProxyType)
			}

			if cfg.DataplaneRuntime.EnvoyLogLevel == "" {
				cfg.DataplaneRuntime.EnvoyLogLevel = rootCtx.LogLevel.String()
			}

			proxyResource, err = readResource(cmd, &cfg.DataplaneRuntime)
			if err != nil {
				runLog.Error(err, "failed to read policy", "proxyType", cfg.Dataplane.ProxyType)

				return err
			}

			if proxyResource != nil {
				if resType := proxyTypeMap[cfg.Dataplane.ProxyType]; resType != proxyResource.Descriptor().Name {
					return errors.Errorf("invalid proxy resource type %q, expected %s",
						proxyResource.Descriptor().Name, resType)
				}

				if cfg.Dataplane.Name != "" || cfg.Dataplane.Mesh != "" {
					return errors.New("--name and --mesh cannot be specified when a dataplane definition is provided, mesh and name will be read from the dataplane definition")
				}

				cfg.Dataplane.Mesh = proxyResource.GetMeta().GetMesh()
				cfg.Dataplane.Name = proxyResource.GetMeta().GetName()
			}

			if cfg.DataplaneRuntime.ConfigDir == "" || cfg.DNS.ConfigDir == "" {
				tmpDir, err = os.MkdirTemp("", "kuma-dp-")
				if err != nil {
					runLog.Error(err, "unable to create a temporary directory to store generated configuration")
					return err
				}

				if cfg.DataplaneRuntime.ConfigDir == "" {
					cfg.DataplaneRuntime.ConfigDir = tmpDir
				}

				if cfg.DataplaneRuntime.SocketDir == "" {
					cfg.DataplaneRuntime.SocketDir = tmpDir
				}

				if cfg.DNS.ConfigDir == "" {
					cfg.DNS.ConfigDir = tmpDir
				}

				runLog.Info("generated configurations will be stored in a temporary directory", "dir", tmpDir)
			}

			if cfg.ControlPlane.CaCert == "" && cfg.ControlPlane.CaCertFile != "" {
				cert, err := os.ReadFile(cfg.ControlPlane.CaCertFile)
				if err != nil {
					return errors.Wrapf(err, "could not read certificate file %s", cfg.ControlPlane.CaCertFile)
				}
				cfg.ControlPlane.CaCert = string(cert)
			}
			return nil
		},
		PostRunE: func(cmd *cobra.Command, _ []string) error {
			tokenComp, err := rootCtx.DataplaneTokenGenerator(cfg)
			if err != nil {
				runLog.Error(err, "unable to get or generate dataplane token")
				return err
			}

			if tmpDir != "" { // clean up temp dir if it was created
				defer func() {
					if err := os.RemoveAll(tmpDir); err != nil {
						runLog.Error(err, "unable to remove a temporary directory with a generated Envoy config")
					}
				}()
			}

			// gracefulCtx indicate that the process received a signal to shutdown
			gracefulCtx, ctx := opts.SetupSignalHandler()
			// componentCtx indicates that components should shutdown (you can use cancel to trigger the shutdown of all components)
			componentCtx, cancelComponents := context.WithCancel(gracefulCtx)
			accessLogSocketPath := core_xds.AccessLogSocketName(cfg.DataplaneRuntime.SocketDir, cfg.Dataplane.Name, cfg.Dataplane.Mesh)
			components := []component.Component{
				tokenComp,
				component.NewResilientComponent(
					runLog.WithName("access-log-streamer"),
					accesslogs.NewAccessLogStreamer(accessLogSocketPath),
				),
			}

			opts := envoy.Opts{
				Config:    *cfg,
				Dataplane: rest.From.Resource(proxyResource),
				Stdout:    cmd.OutOrStdout(),
				Stderr:    cmd.OutOrStderr(),
				OnFinish:  cancelComponents,
			}

			if cfg.DNS.Enabled && !cfg.Dataplane.IsZoneProxy() {
				dnsOpts := &dnsserver.Opts{
					Config:   *cfg,
					Stdout:   cmd.OutOrStdout(),
					Stderr:   cmd.OutOrStderr(),
					OnFinish: cancelComponents,
				}

				dnsServer, err := dnsserver.New(dnsOpts)
				if err != nil {
					return err
				}

				version, err := dnsServer.GetVersion()
				if err != nil {
					return err
				}

				rootCtx.BootstrapDynamicMetadata[core_xds.FieldPrefixDependenciesVersion+".coredns"] = version

				components = append(components, dnsServer)
			}

			envoyVersion, err := envoy.GetEnvoyVersion(opts.Config.DataplaneRuntime.BinaryPath)
			if err != nil {
				return errors.Wrap(err, "failed to get Envoy version")
			}

			if envoyVersion.KumaDpCompatible, err = envoy.VersionCompatible("~"+kuma_version.Envoy, envoyVersion.Version); err != nil {
				runLog.Error(err, "cannot determine envoy version compatibility")
			} else if !envoyVersion.KumaDpCompatible {
				runLog.Info("Envoy version incompatible", "expected", kuma_version.Envoy, "current", envoyVersion.Version)
			}

			runLog.Info("fetched Envoy version", "version", envoyVersion)

			runLog.Info("generating bootstrap configuration")
			metricsSocketPath := core_xds.MetricsHijackerSocketName(cfg.DataplaneRuntime.SocketDir, cfg.Dataplane.Name, cfg.Dataplane.Mesh)
			bootstrap, kumaSidecarConfiguration, err := rootCtx.BootstrapGenerator(gracefulCtx, opts.Config.ControlPlane.URL, opts.Config, envoy.BootstrapParams{
				Dataplane:           opts.Dataplane,
				DNSPort:             cfg.DNS.EnvoyDNSPort,
				ReadinessPort:       cfg.Dataplane.ReadinessPort,
				EmptyDNSPort:        cfg.DNS.CoreDNSEmptyPort,
				EnvoyVersion:        *envoyVersion,
				AccessLogSocketPath: accessLogSocketPath,
				MetricsSocketPath:   metricsSocketPath,
				DynamicMetadata:     rootCtx.BootstrapDynamicMetadata,
				MetricsCertPath:     cfg.DataplaneRuntime.Metrics.CertPath,
				MetricsKeyPath:      cfg.DataplaneRuntime.Metrics.KeyPath,
			})
			if err != nil {
				return errors.Errorf("Failed to generate Envoy bootstrap config. %v", err)
			}
			runLog.Info("received bootstrap configuration", "adminPort", bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetPortValue())

			opts.BootstrapConfig, err = proto.ToYAML(bootstrap)
			if err != nil {
				return errors.Errorf("could not convert to yaml. %v", err)
			}
			opts.AdminPort = bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetPortValue()

			envoyComponent, err := envoy.New(opts)
			if err != nil {
				return err
			}
			components = append(components, envoyComponent)

			metricsServer := metrics.New(
				metricsSocketPath,
				getApplicationsToScrape(kumaSidecarConfiguration, bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetPortValue()),
				kumaSidecarConfiguration.Networking.IsUsingTransparentProxy,
			)
			components = append(components, metricsServer)

			var readinessReporter *readiness.Reporter
			if cfg.Dataplane.ReadinessPort > 0 {
				readinessReporter = readiness.NewReporter(
					bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetAddress(),
					cfg.Dataplane.ReadinessPort)
				components = append(components, readinessReporter)
			}

			if err := rootCtx.ComponentManager.Add(components...); err != nil {
				return err
			}

			stopComponents := make(chan struct{})
			go func() {
				select {
				case <-gracefulCtx.Done():
					runLog.Info("Kuma DP caught an exit signal. Draining Envoy connections")
					if err := envoyComponent.DrainConnections(); err != nil {
						runLog.Error(err, "could not drain connections")
					} else {
						if readinessReporter != nil {
							readinessReporter.Terminating()
						}
						runLog.Info("waiting for connections to be drained", "waitTime", cfg.Dataplane.DrainTime)
						select {
						case <-time.After(cfg.Dataplane.DrainTime.Duration):
						case <-ctx.Done():
						}
					}
				case <-componentCtx.Done():
				}
				runLog.Info("stopping all Kuma DP components")
				close(stopComponents)
			}()

			runLog.Info("starting Kuma DP", "version", kuma_version.Build.Version)
			if err := rootCtx.ComponentManager.Start(stopComponents); err != nil {
				runLog.Error(err, "error while running Kuma DP")
				return err
			}
			runLog.Info("stopping Kuma DP")
			return nil
		},
	}
	var bootstrapVersion string
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Name, "name", cfg.Dataplane.Name, "Name of the Dataplane")
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.Mesh, "mesh", cfg.Dataplane.Mesh, "Mesh that Dataplane belongs to")
	cmd.PersistentFlags().StringVar(&cfg.Dataplane.ProxyType, "proxy-type", "dataplane", `type of the Dataplane ("dataplane", "ingress")`)
	cmd.PersistentFlags().DurationVar(&cfg.Dataplane.DrainTime.Duration, "drain-time", cfg.Dataplane.DrainTime.Duration, `drain time for Envoy connections on Kuma DP shutdown`)
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.URL, "cp-address", cfg.ControlPlane.URL, "URL of the Control Plane Dataplane Server. Example: https://localhost:5678")
	cmd.PersistentFlags().StringVar(&cfg.ControlPlane.CaCertFile, "ca-cert-file", cfg.ControlPlane.CaCertFile, "Path to CA cert by which connection to the Control Plane will be verified if HTTPS is used")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.BinaryPath, "binary-path", cfg.DataplaneRuntime.BinaryPath, "Binary path of Envoy executable")
	cmd.PersistentFlags().Uint32Var(&cfg.DataplaneRuntime.Concurrency, "concurrency", cfg.DataplaneRuntime.Concurrency, "Number of Envoy worker threads")
	// todo(lobkovilya): delete deprecated bootstrap-version flag. Issue https://github.com/kumahq/kuma/issues/2986
	cmd.PersistentFlags().StringVar(&bootstrapVersion, "bootstrap-version", "", "Bootstrap version (and API version) of xDS config. If empty, default version defined in Kuma CP will be used. (ex. '2', '3')")
	_ = cmd.PersistentFlags().MarkDeprecated("bootstrap-version", "Envoy API v3 is used and can not be changed")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.ConfigDir, "config-dir", cfg.DataplaneRuntime.ConfigDir, "Directory in which Envoy config will be generated")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.TokenPath, "dataplane-token-file", cfg.DataplaneRuntime.TokenPath, "Path to a file with dataplane token (use 'kumactl generate dataplane-token' to get one)")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.Token, "dataplane-token", cfg.DataplaneRuntime.Token, "Dataplane Token")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.Resource, "dataplane", "", "Dataplane template to apply (YAML or JSON)")
	cmd.PersistentFlags().StringVarP(&cfg.DataplaneRuntime.ResourcePath, "dataplane-file", "d", "", "Path to Dataplane template to apply (YAML or JSON)")
	cmd.PersistentFlags().StringToStringVarP(&cfg.DataplaneRuntime.ResourceVars, "dataplane-var", "v", map[string]string{}, "Variables to replace Dataplane template")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.EnvoyLogLevel, "envoy-log-level", "", "Envoy log level. Available values are: [trace][debug][info][warning|warn][error][critical][off]. By default it inherits Kuma DP logging level.")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.EnvoyComponentLogLevel, "envoy-component-log-level", "", "Configures Envoy's --component-log-level")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.Metrics.CertPath, "metrics-cert-path", cfg.DataplaneRuntime.Metrics.CertPath, "A path to the certificate for metrics listener")
	cmd.PersistentFlags().StringVar(&cfg.DataplaneRuntime.Metrics.KeyPath, "metrics-key-path", cfg.DataplaneRuntime.Metrics.KeyPath, "A path to the certificate key for metrics listener")
	cmd.PersistentFlags().BoolVar(&cfg.DNS.Enabled, "dns-enabled", cfg.DNS.Enabled, "If true then builtin DNS functionality is enabled and CoreDNS server is started")
	cmd.PersistentFlags().Uint32Var(&cfg.DNS.EnvoyDNSPort, "dns-envoy-port", cfg.DNS.EnvoyDNSPort, "A port that handles Virtual IP resolving by Envoy. CoreDNS should be configured that it first tries to use this DNS resolver and then the real one")
	cmd.PersistentFlags().Uint32Var(&cfg.DNS.CoreDNSPort, "dns-coredns-port", cfg.DNS.CoreDNSPort, "A port that handles DNS requests. When transparent proxy is enabled then iptables will redirect DNS traffic to this port.")
	cmd.PersistentFlags().Uint32Var(&cfg.DNS.CoreDNSEmptyPort, "dns-coredns-empty-port", cfg.DNS.CoreDNSEmptyPort, "A port that always responds with empty NXDOMAIN respond. It is required to implement a fallback to a real DNS.")
	cmd.PersistentFlags().StringVar(&cfg.DNS.CoreDNSBinaryPath, "dns-coredns-path", cfg.DNS.CoreDNSBinaryPath, "A path to CoreDNS binary.")
	cmd.PersistentFlags().StringVar(&cfg.DNS.CoreDNSConfigTemplatePath, "dns-coredns-config-template-path", cfg.DNS.CoreDNSConfigTemplatePath, "A path to a CoreDNS config template.")
	cmd.PersistentFlags().StringVar(&cfg.DNS.ConfigDir, "dns-server-config-dir", cfg.DNS.ConfigDir, "Directory in which DNS Server config will be generated")
	cmd.PersistentFlags().Uint32Var(&cfg.DNS.PrometheusPort, "dns-prometheus-port", cfg.DNS.PrometheusPort, "A port for exposing Prometheus stats")
	return cmd
}

func getApplicationsToScrape(kumaSidecarConfiguration *types.KumaSidecarConfiguration, envoyAdminPort uint32) []metrics.ApplicationToScrape {
	applicationsToScrape := []metrics.ApplicationToScrape{}
	if kumaSidecarConfiguration != nil {
		for _, item := range kumaSidecarConfiguration.Metrics.Aggregate {
			applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
				Address:       item.Address,
				Name:          item.Name,
				Path:          item.Path,
				Port:          item.Port,
				IsIPv6:        net.IsAddressIPv6(item.Address),
				QueryModifier: metrics.RemoveQueryParameters,
			})
		}
	}
	// by default add envoy configuration
	applicationsToScrape = append(applicationsToScrape, metrics.ApplicationToScrape{
		Name:          "envoy",
		Path:          "/stats",
		Address:       "127.0.0.1",
		Port:          envoyAdminPort,
		IsIPv6:        false,
		QueryModifier: metrics.AddPrometheusFormat,
		Mutator:       metrics.MergeClusters,
	})
	return applicationsToScrape
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filename), perm); err != nil {
		return err
	}
	return os.WriteFile(filename, data, perm)
}
