package kuma_cp

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/diagnostics"
	"github.com/kumahq/kuma/pkg/config/multicluster"

	"github.com/kumahq/kuma/pkg/config"
	admin_server "github.com/kumahq/kuma/pkg/config/admin-server"
	api_server "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	dns_server "github.com/kumahq/kuma/pkg/config/dns-server"
	gui_server "github.com/kumahq/kuma/pkg/config/gui-server"
	"github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/config/plugins/runtime"
	"github.com/kumahq/kuma/pkg/config/sds"
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
	Enabled           bool `yaml:"enabled" envconfig:"kuma_metrics_dataplane_enabled"`
	SubscriptionLimit int  `yaml:"subscriptionLimit" envconfig:"kuma_metrics_dataplane_subscription_limit"`
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
	Enabled           bool `yaml:"enabled" envconfig:"kuma_metrics_zone_enabled"`
	SubscriptionLimit int  `yaml:"subscriptionLimit" envconfig:"kuma_metrics_zone_subscription_limit"`
}

func (d *ZoneMetrics) Sanitize() {
}

func (d *ZoneMetrics) Validate() error {
	if d.SubscriptionLimit < 0 {
		return errors.New("SubscriptionLimit should be positive or equal 0")
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
	// Envoy SDS server configuration
	SdsServer *sds.SdsServerConfig `yaml:"sdsServer,omitempty"`
	// Monitoring Assignment Discovery Service (MADS) server configuration
	MonitoringAssignmentServer *mads.MonitoringAssignmentServerConfig `yaml:"monitoringAssignmentServer,omitempty"`
	// Admin server configuration
	AdminServer *admin_server.AdminServerConfig `yaml:"adminServer,omitempty"`
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
	// Multicluster Config
	Multicluster *multicluster.MulticlusterConfig `yaml:"multicluster,omitempty"`
	// DNS Server Config
	DNSServer *dns_server.DNSServerConfig `yaml:"dnsServer,omitempty"`
	// Diagnostics configuration
	Diagnostics *diagnostics.DiagnosticsConfig `yaml:"diagnostics,omitempty"`
}

func (c *Config) Sanitize() {
	c.General.Sanitize()
	c.Store.Sanitize()
	c.BootstrapServer.Sanitize()
	c.XdsServer.Sanitize()
	c.SdsServer.Sanitize()
	c.MonitoringAssignmentServer.Sanitize()
	c.AdminServer.Sanitize()
	c.ApiServer.Sanitize()
	c.Runtime.Sanitize()
	c.Metrics.Sanitize()
	c.Defaults.Sanitize()
	c.GuiServer.Sanitize()
	c.DNSServer.Sanitize()
	c.Multicluster.Sanitize()
	c.Diagnostics.Sanitize()
}

func DefaultConfig() Config {
	return Config{
		Environment:                core.UniversalEnvironment,
		Mode:                       core.Standalone,
		Store:                      store.DefaultStoreConfig(),
		XdsServer:                  xds.DefaultXdsServerConfig(),
		SdsServer:                  sds.DefaultSdsServerConfig(),
		MonitoringAssignmentServer: mads.DefaultMonitoringAssignmentServerConfig(),
		AdminServer:                admin_server.DefaultAdminServerConfig(),
		ApiServer:                  api_server.DefaultApiServerConfig(),
		BootstrapServer:            bootstrap.DefaultBootstrapServerConfig(),
		Runtime:                    runtime.DefaultRuntimeConfig(),
		Defaults: &Defaults{
			SkipMeshCreation: false,
		},
		Metrics: &Metrics{
			Dataplane: &DataplaneMetrics{
				Enabled:           true,
				SubscriptionLimit: 10,
			},
			Zone: &ZoneMetrics{
				Enabled:           true,
				SubscriptionLimit: 10,
			},
		},
		Reports: &Reports{
			Enabled: true,
		},
		General:      DefaultGeneralConfig(),
		GuiServer:    gui_server.DefaultGuiServerConfig(),
		DNSServer:    dns_server.DefaultDNSServerConfig(),
		Multicluster: multicluster.DefaultMulticlusterConfig(),
		Diagnostics:  diagnostics.DefaultDiagnosticsConfig(),
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
		if err := c.Multicluster.Global.Validate(); err != nil {
			return errors.Wrap(err, "Multicluster Global validation failed")
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
		if err := c.SdsServer.Validate(); err != nil {
			return errors.Wrap(err, "SDS Server validation failed")
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
	case core.Remote:
		if err := c.Multicluster.Remote.Validate(); err != nil {
			return errors.Wrap(err, "Multicluster Remote validation failed")
		}
		if err := c.XdsServer.Validate(); err != nil {
			return errors.Wrap(err, "Xds Server validation failed")
		}
		if err := c.BootstrapServer.Validate(); err != nil {
			return errors.Wrap(err, "Bootstrap Server validation failed")
		}
		if err := c.SdsServer.Validate(); err != nil {
			return errors.Wrap(err, "SDS Server validation failed")
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
	if err := c.AdminServer.Validate(); err != nil {
		return errors.Wrap(err, "Admin Server validation failed")
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
	// Hostname that other components should use in order to connect to the Control Plane.
	// Control Plane will use this value in configuration generated for dataplanes, in responses to `kumactl`, etc.
	AdvertisedHostname string `yaml:"advertisedHostname" envconfig:"kuma_general_advertised_hostname"`
	// DNSCacheTTL represents duration for how long Kuma CP will cache result of resolving dataplane's domain name
	DNSCacheTTL time.Duration `yaml:"dnsCacheTTL" envconfig:"kuma_general_dns_cache_ttl"`
	// TlsCertFile defines a path to a file with PEM-encoded TLS cert that will be used across all the Kuma Servers.
	TlsCertFile string `yaml:"tlsCertFile" envconfig:"kuma_general_tls_cert_file"`
	// TlsKeyFile defines a path to a file with PEM-encoded TLS key that will be used across all the Kuma Servers.
	TlsKeyFile string `yaml:"tlsKeyFile" envconfig:"kuma_general_tls_key_file"`
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
		AdvertisedHostname: "localhost",
		DNSCacheTTL:        10 * time.Second,
	}
}
