package postgres

import (
	"github.com/Kong/kuma/pkg/config"
	"github.com/pkg/errors"
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
	// SSL settings
	SSL SSLPostgresStoreConfig `yaml:"ssl" envconfig:"kuma_store_postgres_ssl"`
}

// Modes available here https://godoc.org/github.com/lib/pq
type SSLMode string

const (
	Disable SSLMode = "disable"
	// Always SSL (skip verification)
	Require SSLMode = "require"
	// Always SSL (verify that the certificate presented by the server was signed by a trusted CA)
	VerifyCa SSLMode = "verify-ca"
	// Always SSL (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)
	VerifyFull SSLMode = "verify-full"
)

type SSLPostgresStoreConfig struct {
	// Mode of SSL connection. Available values (disable, require, verify-ca, verify-full)
	Mode SSLMode `yaml:"mode" envconfig:"kuma_store_postgres_ssl_mode"`
	// Path to SSL Certificate of the client. Used in require, verify-ca and verify-full modes
	CertPath string `yaml:"certPath" envconfig:"kuma_store_postgres_ssl_cert_path"`
	// Path to SSL Key of the client. Used in require, verify-ca and verify-full modes
	KeyPath string `yaml:"keyPath" envconfig:"kuma_store_postgres_ssl_key_path"`
	// Path to the root certificate. Used in verify-ca and verify-full modes.
	RootCertPath string `yaml:"rootCertPath" envconfig:"kuma_store_postgres_ssl_root_cert_path"`
}

func (s SSLPostgresStoreConfig) Sanitize() {
}

func (s SSLPostgresStoreConfig) Validate() error {
	switch s.Mode {
	case VerifyFull:
		fallthrough
	case VerifyCa:
		if s.RootCertPath == "" {
			return errors.New("RootCertPath cannot be empty")
		}
		fallthrough
	case Require:
		if s.CertPath == "" {
			return errors.New("CertPath cannot be empty")
		}
		if s.KeyPath == "" {
			return errors.New("KeyPath cannot be empty")
		}
	case Disable:
	default:
		return errors.Errorf("invalid mode: %s", s.Mode)
	}
	return nil
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
	if err := p.SSL.Validate(); err != nil {
		return errors.Wrap(err, "SSL validation failed")
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
		SSL:               DefaultSSLPostgresStoreConfig(),
	}
}

var _ config.Config = &SSLPostgresStoreConfig{}

func DefaultSSLPostgresStoreConfig() SSLPostgresStoreConfig {
	return SSLPostgresStoreConfig{
		Mode:         Disable,
		CertPath:     "",
		KeyPath:      "",
		RootCertPath: "",
	}
}
