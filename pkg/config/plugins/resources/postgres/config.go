package postgres

import (
	"errors"

	"github.com/Kong/kuma/pkg/config"
)

var _ config.Config = &PostgresStoreConfig{}

// Postgres store configuration
type PostgresStoreConfig struct {
	// Host of the Postgres DB
	Host string `yaml:"host" envconfig:"kuma_store_postgres_host"`
	// Port of the Postgres DB
	Port int `yaml:"port" envconfig:"kuma_store_postgres_port"`
	// User of the Postgres DB
	User string `yaml:"user" envconfig:"kuma_store_postgres_user"`
	// Password of the Postgres DB
	Password string `yaml:"password" envconfig:"kuma_store_postgres_password"`
	// Database name of the Postgres DB
	DbName string `yaml:"dbName" envconfig:"kuma_store_postgres_db_name"`
	// Connection Timeout to the DB in seconds
	ConnectionTimeout int `yaml:"connectionTimeout" envconfig:"kuma_store_postgres_connection_timeout"`
}

func (p *PostgresStoreConfig) Sanitize() {
	p.Password = config.SanitizedValue
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
		Host:              "127.0.0.1",
		Port:              15432,
		User:              "kuma",
		Password:          "kuma",
		DbName:            "kuma",
		ConnectionTimeout: 5,
	}
}
