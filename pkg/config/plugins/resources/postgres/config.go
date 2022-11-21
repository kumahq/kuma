package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var _ config.Config = &PostgresStoreConfig{}

// Postgres store configuration
type PostgresStoreConfig struct {
	// Host of the Postgres DB
	Host string `json:"host" envconfig:"kuma_store_postgres_host"`
	// Port of the Postgres DB
	Port int `json:"port" envconfig:"kuma_store_postgres_port"`
	// User of the Postgres DB
	User string `json:"user" envconfig:"kuma_store_postgres_user"`
	// Password of the Postgres DB
	Password string `json:"password" envconfig:"kuma_store_postgres_password"`
	// Database name of the Postgres DB
	DbName string `json:"dbName" envconfig:"kuma_store_postgres_db_name"`
	// Connection Timeout to the DB in seconds
	ConnectionTimeout int `json:"connectionTimeout" envconfig:"kuma_store_postgres_connection_timeout"`
	// Maximum number of open connections to the database
	// `0` value means number of open connections is unlimited
	MaxOpenConnections int `json:"maxOpenConnections" envconfig:"kuma_store_postgres_max_open_connections"`
	// Maximum number of connections in the idle connection pool
	// <0 value means no idle connections and 0 means default max idle connections
	MaxIdleConnections int `json:"maxIdleConnections" envconfig:"kuma_store_postgres_max_idle_connections"`
	// TLS settings
	TLS TLSPostgresStoreConfig `json:"tls"`
	// MinReconnectInterval controls the duration to wait before trying to
	// re-establish the database connection after connection loss. After each
	// consecutive failure this interval is doubled, until MaxReconnectInterval
	// is reached. Successfully completing the connection establishment procedure
	// resets the interval back to MinReconnectInterval.
	MinReconnectInterval config_types.Duration `json:"minReconnectInterval" envconfig:"kuma_store_postgres_min_reconnect_interval"`
	// MaxReconnectInterval controls the maximum possible duration to wait before trying
	// to re-establish the database connection after connection loss.
	MaxReconnectInterval config_types.Duration `json:"maxReconnectInterval" envconfig:"kuma_store_postgres_max_reconnect_interval"`
}

func (cfg PostgresStoreConfig) ConnectionString() (string, error) {
	mode, err := cfg.TLS.Mode.postgresMode()
	if err != nil {
		return "", err
	}
	escape := func(value string) string { return strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `'`, `\'`) }
	boolOption := func(value bool) string {
		if value {
			return "1"
		} else {
			return "0"
		}
	}
	return fmt.Sprintf(
		`host='%s' port=%d user='%s' password='%s' dbname='%s' connect_timeout=%d sslmode=%s sslcert='%s' sslkey='%s' sslrootcert='%s' sslsni=%s`,
		escape(cfg.Host), cfg.Port, escape(cfg.User), escape(cfg.Password), escape(cfg.DbName), cfg.ConnectionTimeout, mode, escape(cfg.TLS.CertPath), escape(cfg.TLS.KeyPath), escape(cfg.TLS.CAPath), boolOption(!cfg.TLS.DisableSSLSNI),
	), nil
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

func (mode TLSMode) postgresMode() (string, error) {
	switch mode {
	case Disable:
		return "disable", nil
	case VerifyNone:
		return "require", nil
	case VerifyCa:
		return "verify-ca", nil
	case VerifyFull:
		return "verify-full", nil
	default:
		return "", errors.Errorf("could not translate mode %q to postgres mode", mode)
	}
}

type TLSPostgresStoreConfig struct {
	// Mode of TLS connection. Available values (disable, verifyNone, verifyCa, verifyFull)
	Mode TLSMode `json:"mode" envconfig:"kuma_store_postgres_tls_mode"`
	// Path to TLS Certificate of the client. Used in require, verifyCa and verifyFull modes
	CertPath string `json:"certPath" envconfig:"kuma_store_postgres_tls_cert_path"`
	// Path to TLS Key of the client. Used in verifyNone, verifyCa and verifyFull modes
	KeyPath string `json:"keyPath" envconfig:"kuma_store_postgres_tls_key_path"`
	// Path to the root certificate. Used in verifyCa and verifyFull modes.
	CAPath string `json:"caPath" envconfig:"kuma_store_postgres_tls_ca_path"`
	// Whether to disable SNI the postgres `sslsni` option.
	DisableSSLSNI bool `json:"disableSSLSNI" envconfig:"kuma_store_postgres_tls_disable_sslsni"`
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
	if p.MinReconnectInterval.Duration >= p.MaxReconnectInterval.Duration {
		return errors.New("MinReconnectInterval should be less than MaxReconnectInterval")
	}
	return nil
}

func DefaultPostgresStoreConfig() *PostgresStoreConfig {
	return &PostgresStoreConfig{
		Host:                 "127.0.0.1",
		Port:                 15432,
		User:                 "kuma",
		Password:             "kuma",
		DbName:               "kuma",
		ConnectionTimeout:    5,
		MaxOpenConnections:   50, // 0 for unlimited
		MaxIdleConnections:   50, // 0 for unlimited
		TLS:                  DefaultTLSPostgresStoreConfig(),
		MinReconnectInterval: config_types.Duration{Duration: 10 * time.Second},
		MaxReconnectInterval: config_types.Duration{Duration: 60 * time.Second},
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
