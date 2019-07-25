package konvoy_cp

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/xds"
	"github.com/pkg/errors"
)

var _ config.Config = &Config{}

type EnvironmentType = string

const (
	KubernetesEnvironmentType EnvironmentType = "kubernetes"
	StandaloneEnvironmentType EnvironmentType = "standalone"
)

type Config struct {
	// Environment Type, can be either "kubernetes" or "standalone"
	Environment EnvironmentType `yaml:"environment" envconfig:"konvoy_environment"`
	// Resource Store configuration
	Store *store.StoreConfig `yaml:"store"`
	// Envoy XDS server configuration
	XdsServer *xds.XdsServerConfig `yaml:"xdsServer"`
	// API Server configuration
	ApiServer *api_server.ApiServerConfig `yaml:"apiServer"`
}

func DefaultConfig() Config {
	return Config{
		Environment: StandaloneEnvironmentType,
		XdsServer:   xds.DefaultXdsServerConfig(),
		ApiServer:   api_server.DefaultApiServerConfig(),
		Store:       store.DefaultStoreConfig(),
	}
}

func (c *Config) Validate() error {
	if err := c.XdsServer.Validate(); err != nil {
		return errors.Wrap(err, "Xds Server validation failed")
	}
	if c.Environment != KubernetesEnvironmentType && c.Environment != StandaloneEnvironmentType {
		return errors.Errorf("Environment should be either %s or %s", KubernetesEnvironmentType, StandaloneEnvironmentType)
	}
	if err := c.Store.Validate(); err != nil {
		return errors.Wrap(err, "Store validation failed")
	}
	if err := c.ApiServer.Validate(); err != nil {
		return errors.Wrap(err, "ApiServer validation failed")
	}
	return nil
}
