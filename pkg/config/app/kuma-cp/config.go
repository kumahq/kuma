package kuma_cp

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/config/mode"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/config"
	admin_server "github.com/Kong/kuma/pkg/config/admin-server"
	api_server "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/config/core"
	"github.com/Kong/kuma/pkg/config/core/resources/store"
	dns_server "github.com/Kong/kuma/pkg/config/dns-server"
	gui_server "github.com/Kong/kuma/pkg/config/gui-server"
	"github.com/Kong/kuma/pkg/config/kds"
	"github.com/Kong/kuma/pkg/config/mads"
	"github.com/Kong/kuma/pkg/config/plugins/runtime"
	"github.com/Kong/kuma/pkg/config/sds"
	token_server "github.com/Kong/kuma/pkg/config/token-server"
	"github.com/Kong/kuma/pkg/config/xds"
	"github.com/Kong/kuma/pkg/config/xds/bootstrap"
	util_error "github.com/Kong/kuma/pkg/util/error"
	"github.com/Kong/kuma/pkg/util/proto"
)

var _ config.Config = &Config{}

var _ config.Config = &Defaults{}

type Defaults struct {
	// Default Mesh configuration in YAML that will be applied on first usage of Kuma CP
	Mesh string `yaml:"mesh"`
}

func (d *Defaults) Sanitize() {
}

func (d *Defaults) MeshProto() v1alpha1.Mesh {
	mesh, err := d.parseMesh()
	util_error.MustNot(err)
	return mesh
}

func (d *Defaults) parseMesh() (v1alpha1.Mesh, error) {
	mesh := v1alpha1.Mesh{}
	if err := proto.FromYAML([]byte(d.Mesh), &mesh); err != nil {
		return mesh, errors.Wrap(err, "Mesh is not valid")
	}
	return mesh, nil
}

func (d *Defaults) Validate() error {
	_, err := d.parseMesh()
	return err
}

type Metrics struct {
	Dataplane *DataplaneMetrics `yaml:"dataplane"`
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

type Reports struct {
	// If true then usage stats will be reported
	Enabled bool `yaml:"enabled" envconfig:"kuma_reports_enabled"`
}

type Config struct {
	// General configuration
	General *GeneralConfig `yaml:"general,omitempty"`
	// Environment Type, can be either "kubernetes" or "universal"
	Environment core.EnvironmentType `yaml:"environment,omitempty" envconfig:"kuma_environment"`
	// Resource Store configuration
	Store *store.StoreConfig `yaml:"store,omitempty"`
	// Configuration of Bootstrap Server, which provides bootstrap config to Dataplanes
	BootstrapServer *bootstrap.BootstrapServerConfig `yaml:"bootstrapServer,omitempty"`
	// Envoy XDS server configuration
	XdsServer *xds.XdsServerConfig `yaml:"xdsServer,omitempty"`
	// Envoy SDS server configuration
	SdsServer *sds.SdsServerConfig `yaml:"sdsServer,omitempty"`
	// Dataplane Token server configuration (DEPRECATED: use adminServer)
	DataplaneTokenServer *token_server.DataplaneTokenServerConfig `yaml:"dataplaneTokenServer,omitempty"`
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
	// Kuma CP Mode
	Mode *mode.ModeConfig `yaml:"mode,omitempty"`
	// DNS Server Config
	DNSServer *dns_server.DNSServerConfig `yaml:"dnsServer,omitempty"`
	// KDS configuration
	KDS *kds.KdsConfig `yaml:"kds,omitempty"`
}

func (c *Config) Sanitize() {
	c.General.Sanitize()
	c.Store.Sanitize()
	c.BootstrapServer.Sanitize()
	c.XdsServer.Sanitize()
	c.SdsServer.Sanitize()
	c.DataplaneTokenServer.Sanitize()
	c.MonitoringAssignmentServer.Sanitize()
	c.AdminServer.Sanitize()
	c.ApiServer.Sanitize()
	c.Runtime.Sanitize()
	c.Metrics.Sanitize()
	c.Defaults.Sanitize()
	c.GuiServer.Sanitize()
	c.DNSServer.Sanitize()
	c.Mode.Sanitize()
	c.KDS.Sanitize()
}

func DefaultConfig() Config {
	return Config{
		Environment:                core.UniversalEnvironment,
		Store:                      store.DefaultStoreConfig(),
		XdsServer:                  xds.DefaultXdsServerConfig(),
		SdsServer:                  sds.DefaultSdsServerConfig(),
		DataplaneTokenServer:       token_server.DefaultDataplaneTokenServerConfig(),
		MonitoringAssignmentServer: mads.DefaultMonitoringAssignmentServerConfig(),
		AdminServer:                admin_server.DefaultAdminServerConfig(),
		ApiServer:                  api_server.DefaultApiServerConfig(),
		BootstrapServer:            bootstrap.DefaultBootstrapServerConfig(),
		Runtime:                    runtime.DefaultRuntimeConfig(),
		Defaults: &Defaults{
			Mesh: `type: Mesh
name: default
`,
		},
		Metrics: &Metrics{
			Dataplane: &DataplaneMetrics{
				Enabled:           true,
				SubscriptionLimit: 10,
			},
		},
		Reports: &Reports{
			Enabled: true,
		},
		General:   DefaultGeneralConfig(),
		GuiServer: gui_server.DefaultGuiServerConfig(),
		DNSServer: dns_server.DefaultDNSServerConfig(),
		Mode:      mode.DefaultModeConfig(),
		KDS:       kds.DefaultKdsConfig(),
	}
}

func (c *Config) Validate() error {
	if err := c.Mode.Validate(); err != nil {
		return errors.Wrap(err, "Mode validation failed")
	}
	switch c.Mode.Mode {
	case mode.Global:
		if err := c.GuiServer.Validate(); err != nil {
			return errors.Wrap(err, "GuiServer validation failed")
		}
	case mode.Standalone:
		if err := c.GuiServer.Validate(); err != nil {
			return errors.Wrap(err, "GuiServer validation failed")
		}
		fallthrough
	case mode.Remote:
		if err := c.XdsServer.Validate(); err != nil {
			return errors.Wrap(err, "Xds Server validation failed")
		}
		if err := c.BootstrapServer.Validate(); err != nil {
			return errors.Wrap(err, "Bootstrap Server validation failed")
		}
		if err := c.SdsServer.Validate(); err != nil {
			return errors.Wrap(err, "SDS Server validation failed")
		}
		if err := c.DataplaneTokenServer.Validate(); err != nil {
			return errors.Wrap(err, "Dataplane Token Server validation failed")
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
	if err := c.KDS.Validate(); err != nil {
		return errors.Wrap(err, "KDS validation failed")
	}
	return nil
}

type GeneralConfig struct {
	// Hostname that other components should use in order to connect to the Control Plane.
	// Control Plane will use this value in configuration generated for dataplanes, in responses to `kumactl`, etc.
	AdvertisedHostname string `yaml:"advertisedHostname" envconfig:"kuma_general_advertised_hostname"`
}

var _ config.Config = &GeneralConfig{}

func (g *GeneralConfig) Sanitize() {
}

func (g *GeneralConfig) Validate() error {
	return nil
}

func DefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		AdvertisedHostname: "localhost",
	}
}
