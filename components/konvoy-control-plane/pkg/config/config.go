package config

import "github.com/pkg/errors"

type EnvironmentType = string

const (
	KubernetesEnvironmentType EnvironmentType = "kubernetes"
	StandaloneEnvironmentType EnvironmentType = "standalone"
)

type Config struct {
	GrpcPort        int              `yaml:"grpcPort" envconfig:"konvoy_grpc_port"`
	HttpPort        int              `yaml:"httpPort" envconfig:"konvoy_http_port"`
	DiagnosticsPort int              `yaml:"diagnosticsPort" envconfig:"konvoy_diagnostics_port"`
	Environment     EnvironmentType  `yaml:"environment" envconfig:"konvoy_environment"`
	Store           *StoreConfig     `yaml:"store"`
	ApiServer       *ApiServerConfig `yaml:"apiServer"`
}

func DefaultConfig() Config {
	return Config{
		GrpcPort:        5678,
		HttpPort:        5679,
		DiagnosticsPort: 5680,
		ApiServer:       DefaultApiServerConfig(),
	}
}

func (c *Config) Validate() error {
	if c.GrpcPort < 0 {
		return errors.New("GrpcPort cannot be negative")
	}
	if c.HttpPort < 0 {
		return errors.New("HttpPort cannot be negative")
	}
	if c.DiagnosticsPort < 0 {
		return errors.New("DiagnosticPort cannot be negative")
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

type StoreType = string

const (
	KubernetesStoreType StoreType = "kubernetes"
	PostgresStoreType   StoreType = "postgres"
	MemoryStoreType     StoreType = "memory"
)

type StoreConfig struct {
	Type     StoreType            `yaml:"type" envconfig:"konvoy_store_type"`
	Postgres *PostgresStoreConfig `yaml:"postgres"`
}

func (s *StoreConfig) Validate() error {
	switch s.Type {
	case PostgresStoreType:
		if err := s.Postgres.Validate(); err != nil {
			return errors.Wrap(err, "Postgres validation failed")
		}
	case KubernetesEnvironmentType:
		return nil
	case MemoryStoreType:
		return nil
	default:
		return errors.Errorf("Type should be either %s, %s or %s", PostgresStoreType, KubernetesEnvironmentType, MemoryStoreType)
	}
	return nil
}

type PostgresStoreConfig struct {
	Host     string `yaml:"host" envconfig:"konvoy_store_postgres_host"`
	Port     int    `yaml:"port" envconfig:"konvoy_store_postgres_port"`
	User     string `yaml:"user" envconfig:"konvoy_store_postgres_user"`
	Password string `yaml:"password" envconfig:"konvoy_store_postgres_password"`
	DbName   string `yaml:"dbName" envconfig:"konvoy_store_postgres_db_name"`
}

func (p *PostgresStoreConfig) Validate() error {
	if len(p.Host) < 1 {
		return errors.New("Host should not be empty")
	}
	if p.Port < 0 {
		return errors.New("Port cannot be negative")
	}
	if len(p.User) < 1 {
		return errors.New("User should not be empty")
	}
	if len(p.Password) < 1 {
		return errors.New("Password should not be empty")
	}
	if len(p.DbName) < 1 {
		return errors.New("DbName should not be empty")
	}
	return nil
}

type ApiServerConfig struct {
	BindAddress string `yaml:"bindAddress" envconfig:"konvoy_api_server_bind_address"`
	ReadOnly    bool   `yaml:"readOnly" envconfig:"konvoy_api_server_read_only"`
	ApiDocsPath string `yaml:"apiDocsPath" envconfig:"konvoy_api_server_api_docs_path"`
}

func (a *ApiServerConfig) Validate() error {
	if len(a.BindAddress) < 1 {
		return errors.New("BindAddress should not be empty")
	}
	if len(a.ApiDocsPath) < 1 {
		return errors.New("ApiDocsPath should not be empty")
	}
	return nil
}

func DefaultApiServerConfig() *ApiServerConfig {
	return &ApiServerConfig{
		BindAddress: "0.0.00.0:5681",
		ReadOnly:    false,
		ApiDocsPath: "/apidocs.json",
	}
}
