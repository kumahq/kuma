package konvoy_cp

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	api_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/core/discovery"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/xds"

	"github.com/pkg/errors"
)

var _ config.Config = &Config{}

type EnvironmentType = string

const (
	KubernetesEnvironment EnvironmentType = "kubernetes"
	UniversalEnvironment  EnvironmentType = "universal"
)

// Konvoy default entities
type Defaults struct {
	Mesh v1alpha1.Mesh `yaml:"mesh"`
}

type Config struct {
	// Environment Type, can be either "kubernetes" or "universal"
	Environment EnvironmentType `yaml:"environment" envconfig:"konvoy_environment"`
	// Resource Store configuration
	Store *store.StoreConfig `yaml:"store"`
	// Discovery configuration
	Discovery *discovery.DiscoveryConfig `yaml:"discovery"`
	// Envoy XDS server configuration
	XdsServer *xds.XdsServerConfig `yaml:"xdsServer"`
	// Configuration of Bootstrap Server, which provides bootstrap config to Dataplanes
	BootstrapServer *xds.BootstrapServerConfig `yaml:"bootstrapServer"`
	// API Server configuration
	ApiServer *api_server.ApiServerConfig `yaml:"apiServer"`
	// Defaults that are applied on first run of the Control Plane
	Defaults *Defaults `yaml:"defaults"`
	// Snapshot configuration
	Snapshot *xds.SnapshotConfig
}

func DefaultConfig() Config {
	defaultMesh := v1alpha1.Mesh{
		Mtls: &v1alpha1.Mesh_Mtls{
			Ca: &v1alpha1.CertificateAuthority{
				Type: &v1alpha1.CertificateAuthority_Builtin_{
					Builtin: &v1alpha1.CertificateAuthority_Builtin{},
				},
			},
		},
	}

	return Config{
		Environment:     UniversalEnvironment,
		Store:           store.DefaultStoreConfig(),
		XdsServer:       xds.DefaultXdsServerConfig(),
		ApiServer:       api_server.DefaultApiServerConfig(),
		BootstrapServer: xds.DefaultBootstrapServerConfig(),
		Discovery:       discovery.DefaultDiscoveryConfig(),
		Defaults: &Defaults{
			Mesh: defaultMesh,
		},
		Snapshot: xds.DefaultSnapshotConfig(),
	}
}

func (c *Config) Validate() error {
	if err := c.XdsServer.Validate(); err != nil {
		return errors.Wrap(err, "Xds Server validation failed")
	}
	if err := c.BootstrapServer.Validate(); err != nil {
		return errors.Wrap(err, "Bootstrap Server validation failed")
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
	if err := c.Snapshot.Validate(); err != nil {
		return errors.Wrap(err, "Snapshot validation failed")
	}
	return nil
}
