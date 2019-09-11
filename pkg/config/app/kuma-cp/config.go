package kuma_cp

import (
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/config"
	api_server "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/config/core/discovery"
	"github.com/Kong/kuma/pkg/config/core/resources/store"
	"github.com/Kong/kuma/pkg/config/sds"
	"github.com/Kong/kuma/pkg/config/xds"
	"github.com/Kong/kuma/pkg/util/proto"

	"github.com/pkg/errors"
)

var _ config.Config = &Config{}

type EnvironmentType = string

const (
	KubernetesEnvironment EnvironmentType = "kubernetes"
	UniversalEnvironment  EnvironmentType = "universal"
)

var _ config.Config = &Defaults{}

type Defaults struct {
	// Default Mesh configuration in YAML that will be applied on first usage of Kuma CP
	Mesh string `yaml:"mesh"`
}

func (d *Defaults) MeshProto() (v1alpha1.Mesh, error) {
	mesh := v1alpha1.Mesh{}
	if err := proto.FromYAML([]byte(d.Mesh), &mesh); err != nil {
		return mesh, errors.Wrap(err, "Mesh is not valid")
	}
	return mesh, nil
}

func (d *Defaults) Validate() error {
	_, err := d.MeshProto()
	return err
}

type Reports struct {
	// If true then usage stats will be reported
	Enabled bool `yaml:"enabled" envconfig:"kuma_reports_enabled"`
}

type Config struct {
	// Environment Type, can be either "kubernetes" or "universal"
	Environment EnvironmentType `yaml:"environment" envconfig:"kuma_environment"`
	// Resource Store configuration
	Store *store.StoreConfig `yaml:"store"`
	// Discovery configuration
	Discovery *discovery.DiscoveryConfig `yaml:"discovery"`
	// Configuration of Bootstrap Server, which provides bootstrap config to Dataplanes
	BootstrapServer *xds.BootstrapServerConfig `yaml:"bootstrapServer"`
	// Envoy XDS server configuration
	XdsServer *xds.XdsServerConfig `yaml:"xdsServer"`
	// Envoy SDS server configuration
	SdsServer *sds.SdsServerConfig `yaml:"sdsServer"`
	// API Server configuration
	ApiServer *api_server.ApiServerConfig `yaml:"apiServer"`
	// Default Kuma entities configuration
	Defaults *Defaults `yaml:"defaults"`
	// Reports configuration
	Reports *Reports `yaml:"reports"`
}

func DefaultConfig() Config {
	return Config{
		Environment:     UniversalEnvironment,
		Store:           store.DefaultStoreConfig(),
		XdsServer:       xds.DefaultXdsServerConfig(),
		SdsServer:       sds.DefaultSdsServerConfig(),
		ApiServer:       api_server.DefaultApiServerConfig(),
		BootstrapServer: xds.DefaultBootstrapServerConfig(),
		Discovery:       discovery.DefaultDiscoveryConfig(),
		Defaults: &Defaults{
			Mesh: `type: Mesh
name: default
mtls:
  ca: {}
  enabled: false
`,
		},
		Reports: &Reports{
			Enabled: true,
		},
	}
}

func (c *Config) Validate() error {
	if err := c.XdsServer.Validate(); err != nil {
		return errors.Wrap(err, "Xds Server validation failed")
	}
	if err := c.BootstrapServer.Validate(); err != nil {
		return errors.Wrap(err, "Bootstrap Server validation failed")
	}
	if err := c.SdsServer.Validate(); err != nil {
		return errors.Wrap(err, "SDS Server validation failed")
	}
	if c.Environment != KubernetesEnvironment && c.Environment != UniversalEnvironment {
		return errors.Errorf("Environment should be either %s or %s", KubernetesEnvironment, UniversalEnvironment)
	}
	if err := c.Store.Validate(); err != nil {
		return errors.Wrap(err, "Store validation failed")
	}
	if err := c.ApiServer.Validate(); err != nil {
		return errors.Wrap(err, "ApiServer validation failed")
	}
	if err := c.Discovery.Validate(); err != nil {
		return errors.Wrap(err, "Discovery validation failed")
	}
	if err := c.Defaults.Validate(); err != nil {
		return errors.Wrap(err, "Defaults validation failed")
	}
	return nil
}
