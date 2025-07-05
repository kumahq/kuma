package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

const (
	DriverNamePgx             = "pgx"
	DefaultConnectionTimeout  = 5
	DefaultMaxIdleConnections = 50
	DefaultMinOpenConnections = 0
)

var _ config.Config = &PostgresStoreConfig{}

var (
	DefaultMinReconnectInterval        = config_types.Duration{Duration: 10 * time.Second}
	DefaultMaxReconnectInterval        = config_types.Duration{Duration: 60 * time.Second}
	DefaultMaxConnectionLifetime       = config_types.Duration{Duration: time.Hour}
	DefaultMaxConnectionLifetimeJitter = config_types.Duration{Duration: 1 * time.Minute}
	DefaultHealthCheckInterval         = config_types.Duration{Duration: 30 * time.Second}
	DefaultMaxConnectionIdleTime       = config_types.Duration{Duration: 30 * time.Minute}
	// some of the above settings taken from pgx https://github.com/jackc/pgx/blob/ca022267dbbfe7a8ba7070557352a5cd08f6cb37/pgxpool/pool.go#L18-L22
)

// PostgresStoreConfig defines Postgres store configuration
type PostgresStoreConfig struct {
	config.BaseConfig

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
	// Driver to use, one of: pgx, postgres
	DriverName string `json:"driverName" envconfig:"kuma_store_postgres_driver_name"`
	// Connection Timeout to the DB in seconds
	ConnectionTimeout int `json:"connectionTimeout" envconfig:"kuma_store_postgres_connection_timeout"`
	// MaxConnectionIdleTime (applied only when driverName=pgx) is the duration after which an idle connection will be automatically closed by the health check.
	MaxConnectionIdleTime config_types.Duration `json:"maxConnectionIdleTime" envconfig:"kuma_store_postgres_max_connection_idle_time"`
	// MaxConnectionLifetime (applied only when driverName=pgx) is the duration since creation after which a connection will be automatically closed
	MaxConnectionLifetime config_types.Duration `json:"maxConnectionLifetime" envconfig:"kuma_store_postgres_max_connection_lifetime"`
	// MaxConnectionLifetimeJitter (applied only when driverName=pgx) is the duration after MaxConnectionLifetime to randomly decide to close a connection.
	// This helps prevent all connections from being closed at the exact same time, starving the pool.
	MaxConnectionLifetimeJitter config_types.Duration `json:"maxConnectionLifetimeJitter" envconfig:"kuma_store_postgres_max_connection_lifetime_jitter"`
	// HealthCheckInterval (applied only when driverName=pgx) is the duration between checks of the health of idle connections.
	HealthCheckInterval config_types.Duration `json:"healthCheckInterval" envconfig:"kuma_store_postgres_health_check_interval"`
	// MinOpenConnections (applied only when driverName=pgx) is the minimum number of open connections to the database
	MinOpenConnections int `json:"minOpenConnections" envconfig:"kuma_store_postgres_min_open_connections"`
	// MaxOpenConnections is the maximum number of open connections to the database
	// `0` value means number of open connections is unlimited
	MaxOpenConnections int `json:"maxOpenConnections" envconfig:"kuma_store_postgres_max_open_connections"`
	// MaxIdleConnections is the maximum number of connections in the idle connection pool
	// <0 value means no idle connections and 0 means default max idle connections
	MaxIdleConnections int `json:"maxIdleConnections" envconfig:"kuma_store_postgres_max_idle_connections"`
	// TLS settings
	TLS TLSPostgresStoreConfig `json:"tls"`
	// MaxListQueryElements defines maximum number of changed elements before requesting full list of elements from the store.
	MaxListQueryElements uint32 `json:"maxListQueryElements" envconfig:"kuma_store_postgres_max_list_query_elements"`
	// ReadReplica is a setting for a DB replica used only for read queries
	ReadReplica ReadReplica `json:"readReplica"`
}

type ReadReplica struct {
	// Host of the Postgres DB read replica. If not set, read replica is not used.
	Host string `json:"host" envconfig:"kuma_store_postgres_read_replica_host"`
	// Port of the Postgres DB read replica
	Port uint `json:"port" envconfig:"kuma_store_postgres_read_replica_port"`
	// Ratio in [0-100] range. How many SELECT queries (out of 100) will use read replica.
	Ratio uint `json:"ratio" envconfig:"kuma_store_postgres_read_replica_ratio"`
}

func (cfg ReadReplica) Validate() error {
	if cfg.Ratio > 100 {
		return errors.New(".Ratio out of [0-100] range")
	}
	return nil
}

func (cfg PostgresStoreConfig) ConnectionString() (string, error) {
	mode, err := cfg.TLS.Mode.postgresMode()
	if err != nil {
		return "", err
	}
	escape := func(value string) string { return strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `'`, `\'`) }
	intVariable := func(name string, value int) string {
		return fmt.Sprintf("%s=%d", name, value)
	}
	variable := func(name, value string) string {
		return fmt.Sprintf("%s=%s", name, value)
	}
	quotedVariable := func(name, value string) string {
		return fmt.Sprintf("%s='%s'", name, escape(value))
	}
	variables := []string{
		quotedVariable("host", cfg.Host),
		intVariable("port", cfg.Port),
		quotedVariable("user", cfg.User),
		quotedVariable("password", cfg.Password),
		quotedVariable("dbname", cfg.DbName),
		intVariable("connect_timeout", cfg.ConnectionTimeout),
		variable("sslmode", mode),
		quotedVariable("sslcert", cfg.TLS.CertPath),
		quotedVariable("sslkey", cfg.TLS.KeyPath),
		quotedVariable("sslrootcert", cfg.TLS.CAPath),
	}
	if cfg.TLS.DisableSSLSNI {
		variables = append(variables, "sslsni=0")
	}
	return strings.Join(variables, " "), nil
}

// TLSMode modes available here https://godoc.org/github.com/lib/pq
type TLSMode string

const (
	Disable TLSMode = "disable"
	// VerifyNone represents Always TLS (skip verification)
	VerifyNone TLSMode = "verifyNone"
	// VerifyCa represents Always TLS (verify that the certificate presented by the server was signed by a trusted CA)
	VerifyCa TLSMode = "verifyCa"
	// VerifyFull represents Always TLS (verify that the certification presented by the server was signed by a trusted CA and the server host name matches the one in the certificate)
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
	config.BaseConfig

	// Mode of TLS connection. Available values (disable, verifyNone, verifyCa, verifyFull)
	Mode TLSMode `json:"mode" envconfig:"kuma_store_postgres_tls_mode"`
	// Path to TLS Certificate of the client. Required when server has METHOD=cert
	CertPath string `json:"certPath" envconfig:"kuma_store_postgres_tls_cert_path"`
	// Path to TLS Key of the client. Required when server has METHOD=cert
	KeyPath string `json:"keyPath" envconfig:"kuma_store_postgres_tls_key_path"`
	// Path to the root certificate. Used in verifyCa and verifyFull modes.
	CAPath string `json:"caPath" envconfig:"kuma_store_postgres_tls_ca_path"`
	// Whether to disable SNI the postgres `sslsni` option.
	DisableSSLSNI bool `json:"disableSSLSNI" envconfig:"kuma_store_postgres_tls_disable_sslsni"`
}

func (s TLSPostgresStoreConfig) Validate() error {
	switch s.Mode {
	case VerifyFull, VerifyCa:
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
	if p.MinOpenConnections < 0 {
		return errors.New("MinOpenConnections should be greater than 0")
	}
	if p.MinOpenConnections > p.MaxOpenConnections {
		return errors.New("MinOpenConnections should be less than MaxOpenConnections")
	}
	if p.MaxConnectionLifetime.Duration < 0 {
		return errors.New("MaxConnectionLifetime should be greater than 0")
	}
	if p.MaxConnectionLifetimeJitter.Duration < 0 {
		return errors.New("MaxConnectionLifetimeJitter should be greater than 0")
	}
	if p.MaxConnectionLifetimeJitter.Duration > p.MaxConnectionLifetime.Duration {
		return errors.New("MaxConnectionLifetimeJitter should be less than MaxConnectionLifetime")
	}
	if p.HealthCheckInterval.Duration < 0 {
		return errors.New("HealthCheckInterval should be greater than 0")
	}
	if err := p.ReadReplica.Validate(); err != nil {
		return errors.Wrapf(err, "ReadReplica validation failed")
	}
	return nil
}

func DefaultPostgresStoreConfig() *PostgresStoreConfig {
	return &PostgresStoreConfig{
		Host:                        "127.0.0.1",
		Port:                        15432,
		User:                        "kuma",
		Password:                    "kuma",
		DbName:                      "kuma",
		DriverName:                  DriverNamePgx,
		ConnectionTimeout:           DefaultConnectionTimeout,
		MaxOpenConnections:          50,                        // 0 for unlimited
		MaxIdleConnections:          DefaultMaxIdleConnections, // 0 for unlimited
		TLS:                         DefaultTLSPostgresStoreConfig(),
		MinOpenConnections:          DefaultMinOpenConnections,
		MaxConnectionLifetime:       DefaultMaxConnectionLifetime,
		MaxConnectionLifetimeJitter: DefaultMaxConnectionLifetimeJitter,
		HealthCheckInterval:         DefaultHealthCheckInterval,
		MaxListQueryElements:        0,
		ReadReplica: ReadReplica{
			Port:  5432,
			Ratio: 100,
		},
		MaxConnectionIdleTime: DefaultMaxConnectionIdleTime,
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
