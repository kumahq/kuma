package postgres

import (
	"errors"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
)

var _ config.Config = &PostgresStoreConfig{}

// Postgres store configuration
type PostgresStoreConfig struct {
	// Host of the Postgres DB
	Host string `yaml:"host" envconfig:"konvoy_store_postgres_host"`
	// Port of the Postgres DB
	Port int `yaml:"port" envconfig:"konvoy_store_postgres_port"`
	// User of the Postgres DB
	User string `yaml:"user" envconfig:"konvoy_store_postgres_user"`
	// Password of the Postgres DB
	Password string `yaml:"password" envconfig:"konvoy_store_postgres_password"`
	// Database name of the Postgres DB
	DbName string `yaml:"dbName" envconfig:"konvoy_store_postgres_db_name"`
	// Connection Timeout to the DB in seconds
	ConnectionTimeout int `yaml:"connectionTimeout" envconfig:"konvoy_store_postgres_connection_timeout"`
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

func DefaultPostgresStoreConfig() *PostgresStoreConfig {
	return &PostgresStoreConfig{
		ConnectionTimeout: 5,
	}
}
