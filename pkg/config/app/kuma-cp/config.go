package kuma_cp

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/access"
	api_server "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/config/diagnostics"
	dns_server "github.com/kumahq/kuma/pkg/config/dns-server"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	gui_server "github.com/kumahq/kuma/pkg/config/gui-server"
	"github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/config/plugins/runtime"
	"github.com/kumahq/kuma/pkg/config/xds"
	"github.com/kumahq/kuma/pkg/config/xds/bootstrap"
)

var _ config.Config = &Config{}

var _ config.Config = &Defaults{}

type Defaults struct {
	SkipMeshCreation bool `yaml:"skipMeshCreation" envconfig:"kuma_defaults_skip_mesh_creation"`
}

func (d *Defaults) Sanitize() {
}

func (d *Defaults) Validate() error {
	return nil
}

type Metrics struct {
	Dataplane *DataplaneMetrics `yaml:"dataplane"`
	Zone      *ZoneMetrics      `yaml:"zone"`
	Mesh      *MeshMetrics      `yaml:"mesh"`
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
	SubscriptionLimit int           `yaml:"subscriptionLimit" envconfig:"kuma_metrics_dataplane_subscription_limit"`
	IdleTimeout       time.Duration `yaml:"idleTimeout" envconfig:"kuma_metrics_dataplane_idle_timeout"`
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
	SubscriptionLimit int           `yaml:"subscriptionLimit" envconfig:"kuma_metrics_zone_subscription_limit"`
	IdleTimeout       time.Duration `yaml:"idleTimeout" envconfig:"kuma_metrics_zone_idle_timeout"`
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
	MinResyncTimeout time.Duration `yaml:"minResyncTimeout" envconfig:"kuma_metrics_mesh_min_resync_timeout"`
	// MaxResyncTimeout is a maximum time that MeshInsight could spend without resync
	MaxResyncTimeout time.Duration `yaml:"maxResyncTimeout" envconfig:"kuma_metrics_mesh_max_resync_timeout"`
}

func (d *MeshMetrics) Sanitize() {
}

func (d *MeshMetrics) Validate() error {
	if d.MaxResyncTimeout <= d.MinResyncTimeout {
		return errors.New("MaxResyncTimeout should be greater than MinResyncTimeout")
	}
	return nil
}

type Reports struct {
	// If true then usage stats will be reported
	Enabled bool `yaml:"enabled" envconfig:"kuma_reports_enabled"`
}

type Config struct {
	// General configuration
	General *GeneralConfig `yaml:"general,omitempty"`
	// Environment Type, can be either "kubernetes" or "universal"
	Environment core.EnvironmentType `yaml:"environment,omitempty" envconfig:"kuma_environment"`
	// Mode
	Mode core.CpMode `yaml:"mode" envconfig:"kuma_mode"`
	// Resource Store configuration
	Store *store.StoreConfig `yaml:"store,omitempty"`
	// Configuration of Bootstrap Server, which provides bootstrap config to Dataplanes
	BootstrapServer *bootstrap.BootstrapServerConfig `yaml:"bootstrapServer,omitempty"`
	// Envoy XDS server configuration
	XdsServer *xds.XdsServerConfig `yaml:"xdsServer,omitempty"`
	// Monitoring Assignment Discovery Service (MADS) server configuration
	MonitoringAssignmentServer *mads.MonitoringAssignmentServerConfig `yaml:"monitoringAssignmentServer,omitempty"`
	// API Server configuration
	ApiServer *api_server.ApiServerConfig `yaml:"apiServer,omitempty"`
	// Environment-specific configuration
	Runtime *runtime.RuntimeConfig
	// Default Kuma entities configuration
	Defaults *Defaults `yaml:"defaults,omitempty"`
	// Metrics configuration
	Metrics *Metrics `yaml:"metrics,omitempty"`
	// Reports configuration
	Reports *Reports `yaml:"reports,omitempty"`
	// GUI Server Config
	GuiServer *gui_server.GuiServerConfig `yaml:"guiServer,omitempty"`
	// Multizone Config
	Multizone *multizone.MultizoneConfig `yaml:"multizone,omitempty"`
	// DNS Server Config
	DNSServer *dns_server.DNSServerConfig `yaml:"dnsServer,omitempty"`
	// Diagnostics configuration
	Diagnostics *diagnostics.DiagnosticsConfig `yaml:"diagnostics,omitempty"`
	// Dataplane Server configuration
	DpServer *dp_server.DpServerConfig `yaml:"dpServer"`
	// Access Control configuration
	Access access.AccessConfig `yaml:"access"`
	// Configuration of experimental features
	Experimental ExperimentalConfig `yaml:"experimental"`
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
	c.GuiServer.Sanitize()
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
				IdleTimeout:       5 * time.Minute,
			},
			Zone: &ZoneMetrics{
				SubscriptionLimit: 10,
				IdleTimeout:       5 * time.Minute,
			},
			Mesh: &MeshMetrics{
				MinResyncTimeout: 1 * time.Second,
				MaxResyncTimeout: 20 * time.Second,
			},
		},
		Reports: &Reports{
			Enabled: false,
		},
		General:     DefaultGeneralConfig(),
		GuiServer:   gui_server.DefaultGuiServerConfig(),
		DNSServer:   dns_server.DefaultDNSServerConfig(),
		Multizone:   multizone.DefaultMultizoneConfig(),
		Diagnostics: diagnostics.DefaultDiagnosticsConfig(),
		DpServer:    dp_server.DefaultDpServerConfig(),
		Access:      access.DefaultAccessConfig(),
		Experimental: ExperimentalConfig{
			MeshGateway: false,
		},
	}
}

func (c *Config) Validate() error {
	if err := core.ValidateCpMode(c.Mode); err != nil {
		return errors.Wrap(err, "Mode validation failed")
	}
	switch c.Mode {
	case core.Global:
		if err := c.GuiServer.Validate(); err != nil {
			return errors.Wrap(err, "GuiServer validation failed")
		}
		if err := c.Multizone.Global.Validate(); err != nil {
			return errors.Wrap(err, "Multizone Global validation failed")
		}
	case core.Standalone:
		if err := c.GuiServer.Validate(); err != nil {
			return errors.Wrap(err, "GuiServer validation failed")
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
	return nil
}

type GeneralConfig struct {
	// DNSCacheTTL represents duration for how long Kuma CP will cache result of resolving dataplane's domain name
	DNSCacheTTL time.Duration `yaml:"dnsCacheTTL" envconfig:"kuma_general_dns_cache_ttl"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert that will be used across all the Kuma Servers.
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_general_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key that will be used across all the Kuma Servers.
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_general_tls_key_file"`
	// WorkDir defines a path to the working directory
	WorkDir string `yaml:"workDir" envconfig:"kuma_general_work_dir"`
}

var _ config.Config = &GeneralConfig{}

func (g *GeneralConfig) Sanitize() {
}

func (g *GeneralConfig) Validate() error {
	if g.TlsCertFile == "" && g.TlsKeyFile != "" {
		return errors.New("TlsCertFile cannot be empty if TlsKeyFile has been set")
	}
	if g.TlsKeyFile == "" && g.TlsCertFile != "" {
		return errors.New("TlsKeyFile cannot be empty if TlsCertFile has been set")
	}
	return nil
}

func DefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		DNSCacheTTL: 10 * time.Second,
		WorkDir:     "",
	}
}

type ExperimentalConfig struct {
	// If true, experimental built-in gateway is enabled.
	MeshGateway bool `yaml:"meshGateway" envconfig:"KUMA_EXPERIMENTAL_MESHGATEWAY"`
}
