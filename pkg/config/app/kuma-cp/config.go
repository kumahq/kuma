package kuma_cp

import (
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/access"
	api_server "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/config/diagnostics"
	dns_server "github.com/kumahq/kuma/pkg/config/dns-server"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/config/intercp"
	"github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/config/plugins/runtime"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/config/xds"
	"github.com/kumahq/kuma/pkg/config/xds/bootstrap"
)

var _ config.Config = &Config{}

var _ config.Config = &Defaults{}

type Defaults struct {
	// If true, it skips creating the default Mesh
	SkipMeshCreation bool `json:"skipMeshCreation" envconfig:"kuma_defaults_skip_mesh_creation"`
}

func (d *Defaults) Sanitize() {
}

func (d *Defaults) Validate() error {
	return nil
}

type Metrics struct {
	Dataplane *DataplaneMetrics `json:"dataplane"`
	Zone      *ZoneMetrics      `json:"zone"`
	Mesh      *MeshMetrics      `json:"mesh"`
}

func (m *Metrics) Sanitize() {
}

func (m *Metrics) Validate() error {
	if err := m.Dataplane.Validate(); err != nil {
		return errors.Wrap(err, "Dataplane validation failed")
	}
	return nil
}

type DataplaneMetrics struct {
	SubscriptionLimit int                   `json:"subscriptionLimit" envconfig:"kuma_metrics_dataplane_subscription_limit"`
	IdleTimeout       config_types.Duration `json:"idleTimeout" envconfig:"kuma_metrics_dataplane_idle_timeout"`
}

func (d *DataplaneMetrics) Sanitize() {
}

func (d *DataplaneMetrics) Validate() error {
	if d.SubscriptionLimit < 0 {
		return errors.New("SubscriptionLimit should be positive or equal 0")
	}
	return nil
}

type ZoneMetrics struct {
	SubscriptionLimit int                   `json:"subscriptionLimit" envconfig:"kuma_metrics_zone_subscription_limit"`
	IdleTimeout       config_types.Duration `json:"idleTimeout" envconfig:"kuma_metrics_zone_idle_timeout"`
}

func (d *ZoneMetrics) Sanitize() {
}

func (d *ZoneMetrics) Validate() error {
	if d.SubscriptionLimit < 0 {
		return errors.New("SubscriptionLimit should be positive or equal 0")
	}
	return nil
}

type MeshMetrics struct {
	// MinResyncTimeout is a minimal time that should pass between MeshInsight resync
	MinResyncTimeout config_types.Duration `json:"minResyncTimeout" envconfig:"kuma_metrics_mesh_min_resync_timeout"`
	// MaxResyncTimeout is a maximum time that MeshInsight could spend without resync
	MaxResyncTimeout config_types.Duration `json:"maxResyncTimeout" envconfig:"kuma_metrics_mesh_max_resync_timeout"`
}

func (d *MeshMetrics) Sanitize() {
}

func (d *MeshMetrics) Validate() error {
	if d.MaxResyncTimeout.Duration <= d.MinResyncTimeout.Duration {
		return errors.New("MaxResyncTimeout should be greater than MinResyncTimeout")
	}
	return nil
}

type Reports struct {
	// If true then usage stats will be reported
	Enabled bool `json:"enabled" envconfig:"kuma_reports_enabled"`
}

type Config struct {
	// General configuration
	General *GeneralConfig `json:"general,omitempty"`
	// Environment Type, can be either "kubernetes" or "universal"
	Environment core.EnvironmentType `json:"environment,omitempty" envconfig:"kuma_environment"`
	// Mode in which Kuma CP is running. Available values are: "standalone", "global", "zone"
	Mode core.CpMode `json:"mode" envconfig:"kuma_mode"`
	// Resource Store configuration
	Store *store.StoreConfig `json:"store,omitempty"`
	// Configuration of Bootstrap Server, which provides bootstrap config to Dataplanes
	BootstrapServer *bootstrap.BootstrapServerConfig `json:"bootstrapServer,omitempty"`
	// Envoy XDS server configuration
	XdsServer *xds.XdsServerConfig `json:"xdsServer,omitempty"`
	// Monitoring Assignment Discovery Service (MADS) server configuration
	MonitoringAssignmentServer *mads.MonitoringAssignmentServerConfig `json:"monitoringAssignmentServer,omitempty"`
	// API Server configuration
	ApiServer *api_server.ApiServerConfig `json:"apiServer,omitempty"`
	// Environment-specific configuration
	Runtime *runtime.RuntimeConfig
	// Default Kuma entities configuration
	Defaults *Defaults `json:"defaults,omitempty"`
	// Metrics configuration
	Metrics *Metrics `json:"metrics,omitempty"`
	// Reports configuration
	Reports *Reports `json:"reports,omitempty"`
	// Multizone Config
	Multizone *multizone.MultizoneConfig `json:"multizone,omitempty"`
	// DNS Server Config
	DNSServer *dns_server.Config `json:"dnsServer,omitempty"`
	// Diagnostics configuration
	Diagnostics *diagnostics.DiagnosticsConfig `json:"diagnostics,omitempty"`
	// Dataplane Server configuration
	DpServer *dp_server.DpServerConfig `json:"dpServer"`
	// Access Control configuration
	Access access.AccessConfig `json:"access"`
	// Configuration of experimental features
	Experimental ExperimentalConfig `json:"experimental"`
	// Proxy holds configuration for proxies
	Proxy xds.Proxy `json:"proxy"`
	// Intercommunication CP configuration
	InterCp intercp.InterCpConfig `json:"interCp"`
}

func (c *Config) Sanitize() {
	c.General.Sanitize()
	c.Store.Sanitize()
	c.BootstrapServer.Sanitize()
	c.XdsServer.Sanitize()
	c.MonitoringAssignmentServer.Sanitize()
	c.ApiServer.Sanitize()
	c.Runtime.Sanitize()
	c.Metrics.Sanitize()
	c.Defaults.Sanitize()
	c.DNSServer.Sanitize()
	c.Multizone.Sanitize()
	c.Diagnostics.Sanitize()
}

var DefaultConfig = func() Config {
	return Config{
		Environment:                core.UniversalEnvironment,
		Mode:                       core.Standalone,
		Store:                      store.DefaultStoreConfig(),
		XdsServer:                  xds.DefaultXdsServerConfig(),
		MonitoringAssignmentServer: mads.DefaultMonitoringAssignmentServerConfig(),
		ApiServer:                  api_server.DefaultApiServerConfig(),
		BootstrapServer:            bootstrap.DefaultBootstrapServerConfig(),
		Runtime:                    runtime.DefaultRuntimeConfig(),
		Defaults: &Defaults{
			SkipMeshCreation: false,
		},
		Metrics: &Metrics{
			Dataplane: &DataplaneMetrics{
				SubscriptionLimit: 2,
				IdleTimeout:       config_types.Duration{Duration: 5 * time.Minute},
			},
			Zone: &ZoneMetrics{
				SubscriptionLimit: 10,
				IdleTimeout:       config_types.Duration{Duration: 5 * time.Minute},
			},
			Mesh: &MeshMetrics{
				MinResyncTimeout: config_types.Duration{Duration: 1 * time.Second},
				MaxResyncTimeout: config_types.Duration{Duration: 20 * time.Second},
			},
		},
		Reports: &Reports{
			Enabled: false,
		},
		General:     DefaultGeneralConfig(),
		DNSServer:   dns_server.DefaultDNSServerConfig(),
		Multizone:   multizone.DefaultMultizoneConfig(),
		Diagnostics: diagnostics.DefaultDiagnosticsConfig(),
		DpServer:    dp_server.DefaultDpServerConfig(),
		Access:      access.DefaultAccessConfig(),
		Experimental: ExperimentalConfig{
			GatewayAPI:          false,
			KubeOutboundsAsVIPs: true,
			KDSDeltaEnabled:     false,
		},
		Proxy:   xds.DefaultProxyConfig(),
		InterCp: intercp.DefaultInterCpConfig(),
	}
}

func (c *Config) Validate() error {
	if err := core.ValidateCpMode(c.Mode); err != nil {
		return errors.Wrap(err, "Mode validation failed")
	}
	switch c.Mode {
	case core.Global:
		if err := c.Multizone.Global.Validate(); err != nil {
			return errors.Wrap(err, "Multizone Global validation failed")
		}
	case core.Standalone:
		if err := c.XdsServer.Validate(); err != nil {
			return errors.Wrap(err, "Xds Server validation failed")
		}
		if err := c.BootstrapServer.Validate(); err != nil {
			return errors.Wrap(err, "Bootstrap Server validation failed")
		}
		if err := c.MonitoringAssignmentServer.Validate(); err != nil {
			return errors.Wrap(err, "Monitoring Assignment Server validation failed")
		}
		if c.Environment != core.KubernetesEnvironment && c.Environment != core.UniversalEnvironment {
			return errors.Errorf("Environment should be either %s or %s", core.KubernetesEnvironment, core.UniversalEnvironment)
		}
		if err := c.Runtime.Validate(c.Environment); err != nil {
			return errors.Wrap(err, "Runtime validation failed")
		}
		if err := c.Metrics.Validate(); err != nil {
			return errors.Wrap(err, "Metrics validation failed")
		}
	case core.Zone:
		if err := c.Multizone.Zone.Validate(); err != nil {
			return errors.Wrap(err, "Multizone Zone validation failed")
		}
		if err := c.XdsServer.Validate(); err != nil {
			return errors.Wrap(err, "Xds Server validation failed")
		}
		if err := c.BootstrapServer.Validate(); err != nil {
			return errors.Wrap(err, "Bootstrap Server validation failed")
		}
		if err := c.MonitoringAssignmentServer.Validate(); err != nil {
			return errors.Wrap(err, "Monitoring Assignment Server validation failed")
		}
		if c.Environment != core.KubernetesEnvironment && c.Environment != core.UniversalEnvironment {
			return errors.Errorf("Environment should be either %s or %s", core.KubernetesEnvironment, core.UniversalEnvironment)
		}
		if err := c.Runtime.Validate(c.Environment); err != nil {
			return errors.Wrap(err, "Runtime validation failed")
		}
		if err := c.Metrics.Validate(); err != nil {
			return errors.Wrap(err, "Metrics validation failed")
		}
	}
	if err := c.Store.Validate(); err != nil {
		return errors.Wrap(err, "Store validation failed")
	}
	if err := c.ApiServer.Validate(); err != nil {
		return errors.Wrap(err, "ApiServer validation failed")
	}
	if err := c.Defaults.Validate(); err != nil {
		return errors.Wrap(err, "Defaults validation failed")
	}
	if err := c.DNSServer.Validate(); err != nil {
		return errors.Wrap(err, "DNSServer validation failed")
	}
	if err := c.Diagnostics.Validate(); err != nil {
		return errors.Wrap(err, "Diagnostics validation failed")
	}
	if err := c.Experimental.Validate(); err != nil {
		return errors.Wrap(err, "Experimental validation failed")
	}
	if err := c.InterCp.Validate(); err != nil {
		return errors.Wrap(err, "InterCp validation failed")
	}
	return nil
}

type GeneralConfig struct {
	// DNSCacheTTL represents duration for how long Kuma CP will cache result of resolving dataplane's domain name
	DNSCacheTTL config_types.Duration `json:"dnsCacheTTL" envconfig:"kuma_general_dns_cache_ttl"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert that will be used across all the Kuma Servers.
	TlsCertFile string `json:"tlsCertFile" envconfig:"kuma_general_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key that will be used across all the Kuma Servers.
	TlsKeyFile string `json:"tlsKeyFile" envconfig:"kuma_general_tls_key_file"`
	// TlsMinVersion defines the minimum TLS version to be used
	TlsMinVersion string `json:"tlsMinVersion" envconfig:"kuma_general_tls_min_version"`
	// TlsMaxVersion defines the maximum TLS version to be used
	TlsMaxVersion string `json:"tlsMaxVersion" envconfig:"kuma_general_tls_max_version"`
	// TlsCipherSuites defines the list of ciphers to use
	TlsCipherSuites []string `json:"tlsCipherSuites" envconfig:"kuma_general_tls_cipher_suites"`
	// WorkDir defines a path to the working directory
	// Kuma stores in this directory autogenerated entities like certificates.
	// If empty then the working directory is $HOME/.kuma
	WorkDir string `json:"workDir" envconfig:"kuma_general_work_dir"`
}

var _ config.Config = &GeneralConfig{}

func (g *GeneralConfig) Sanitize() {
}

func (g *GeneralConfig) Validate() error {
	var errs error
	if g.TlsCertFile == "" && g.TlsKeyFile != "" {
		errs = multierr.Append(errs, errors.New(".TlsCertFile cannot be empty if TlsKeyFile has been set"))
	}
	if g.TlsKeyFile == "" && g.TlsCertFile != "" {
		errs = multierr.Append(errs, errors.New(".TlsKeyFile cannot be empty if TlsCertFile has been set"))
	}
	if _, err := config_types.TLSVersion(g.TlsMinVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMinVersion"+err.Error()))
	}
	if _, err := config_types.TLSVersion(g.TlsMaxVersion); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsMaxVersion"+err.Error()))
	}
	if _, err := config_types.TLSCiphers(g.TlsCipherSuites); err != nil {
		errs = multierr.Append(errs, errors.New(".TlsCipherSuites"+err.Error()))
	}
	return errs
}

func DefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		DNSCacheTTL:     config_types.Duration{Duration: 10 * time.Second},
		WorkDir:         "",
		TlsCipherSuites: []string{},
		TlsMinVersion:   "TLSv1_2",
	}
}

type ExperimentalConfig struct {
	// If true, experimental Gateway API is enabled
	GatewayAPI bool `json:"gatewayAPI" envconfig:"KUMA_EXPERIMENTAL_GATEWAY_API"`
	// If true, instead of embedding kubernetes outbounds into Dataplane object, they are persisted next to VIPs in ConfigMap
	// This can improve performance, but it should be enabled only after all instances are migrated to version that supports this config
	KubeOutboundsAsVIPs bool `json:"kubeOutboundsAsVIPs" envconfig:"KUMA_EXPERIMENTAL_KUBE_OUTBOUNDS_AS_VIPS"`
	// KDSDeltaEnabled defines if using KDS Sync with incremental xDS
	KDSDeltaEnabled bool `json:"kdsDeltaEnabled" envconfig:"KUMA_EXPERIMENTAL_KDS_DELTA_ENABLED"`
}

func (e ExperimentalConfig) Validate() error {
	return nil
}

func (c Config) GetEnvoyAdminPort() uint32 {
	if c.BootstrapServer == nil || c.BootstrapServer.Params == nil {
		return 0
	}
	return c.BootstrapServer.Params.AdminPort
}
