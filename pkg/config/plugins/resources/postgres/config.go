package postgres

import (
	"github.com/pkg/errors"

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
	// Maximum number of open connections to the database
	// `0` value means number of open connections is unlimited
	MaxOpenConnections int `yaml:"maxOpenConnections" envconfig:"kuma_store_postgres_max_open_connections"`
	// TLS settings
	TLS TLSPostgresStoreConfig `yaml:"tls"`
}

// Modes available here https://godoc.org/github.com/lib/pq
type TLSMode string

const (
	Disable TLSMode = "disable"
	// Always TLS (skip verification)
	VerifyNone TLSMode = "verifyNone"
	// Always TLS (verify that the certificate presented by the server was signed by a trusted CA)
	VerifyCa TLSMode = "verifyCa"
	// Always TLS (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)
	VerifyFull TLSMode = "verifyFull"
)

type TLSPostgresStoreConfig struct {
	// Mode of TLS connection. Available values (disable, verifyNone, verifyCa, verifyFull)
	Mode TLSMode `yaml:"mode" envconfig:"kuma_store_postgres_tls_mode"`
	// Path to TLS Certificate of the client. Used in require, verifyCa and verifyFull modes
	CertPath string `yaml:"certPath" envconfig:"kuma_store_postgres_tls_cert_path"`
	// Path to TLS Key of the client. Used in verifyNone, verifyCa and verifyFull modes
	KeyPath string `yaml:"keyPath" envconfig:"kuma_store_postgres_tls_key_path"`
	// Path to the root certificate. Used in verifyCa and verifyFull modes.
	CAPath string `yaml:"caPath" envconfig:"kuma_store_postgres_tls_ca_path"`
}

func (s TLSPostgresStoreConfig) Sanitize() {
}

func (s TLSPostgresStoreConfig) Validate() error {
	switch s.Mode {
	case VerifyFull:
		fallthrough
	case VerifyCa:
		if s.CAPath == "" {
			return errors.New("CAPath cannot be empty")
		}
	case VerifyNone:
	case Disable:
	default:
		return errors.Errorf("invalid mode: %s", s.Mode)
	}
	if s.KeyPath == "" && s.CertPath != "" {
		return errors.New("KeyPath cannot be empty when CertPath is provided")
	}
	if s.CertPath == "" && s.KeyPath != "" {
		return errors.New("CertPath cannot be empty when KeyPath is provided")
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
	if err := p.TLS.Validate(); err != nil {
		return errors.Wrap(err, "TLS validation failed")
	}
	return nil
}

func DefaultPostgresStoreConfig() *PostgresStoreConfig {
	return &PostgresStoreConfig{
		Host:               "127.0.0.1",
		Port:               15432,
		User:               "kuma",
		Password:           "kuma",
		DbName:             "kuma",
		ConnectionTimeout:  5,
		MaxOpenConnections: 0, // number of open connections is unlimited
		TLS:                DefaultTLSPostgresStoreConfig(),
	}
}

var _ config.Config = &TLSPostgresStoreConfig{}

func DefaultTLSPostgresStoreConfig() TLSPostgresStoreConfig {
	return TLSPostgresStoreConfig{
		Mode:     Disable,
		CertPath: "",
		KeyPath:  "",
		CAPath:   "",
	}
}
