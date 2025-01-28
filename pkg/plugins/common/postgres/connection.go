package postgres

import (
	"context"
	"database/sql"
	"math"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	pgx_config "github.com/kumahq/kuma/pkg/plugins/resources/postgres/config"
)

func ConnectToDb(cfg config.PostgresStoreConfig) (*sql.DB, error) {
	connStr, err := cfg.ConnectionString()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(cfg.DriverName, connStr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create connection to DB")
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)

	// check connection to DB, Open() does not check it.
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "cannot connect to DB")
	}

	return db, nil
}

// This attribute is necessary for tracing integrations like Datadog, to have
// full insights into sql queries connected with traces.
// ref. https://github.com/DataDog/dd-trace-go/blob/3d97fcec9f8b21fdd821af526d27d4335b26da66/contrib/database/sql/conn.go#L290
var spanTypeSQLAttribute = attribute.String("span.type", "sql")

func ConnectToDbPgx(postgresStoreConfig config.PostgresStoreConfig, customizers ...pgx_config.PgxConfigCustomization) (*pgxpool.Pool, error) {
	connectionString, err := postgresStoreConfig.ConnectionString()
	if err != nil {
		return nil, err
	}
	pgxConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}

	if postgresStoreConfig.MaxOpenConnections == 0 {
		// pgx MaxCons must be > 0, see https://github.com/jackc/puddle/blob/c5402ce53663d3c6481ea83c2912c339aeb94adc/pool.go#L160
		// so unlimited is just max int
		pgxConfig.MaxConns = math.MaxInt32
	} else {
		pgxConfig.MaxConns = int32(postgresStoreConfig.MaxOpenConnections)
	}
	pgxConfig.MaxConnIdleTime = postgresStoreConfig.MaxConnectionIdleTime.Duration
	pgxConfig.MinConns = int32(postgresStoreConfig.MinOpenConnections)
	pgxConfig.MaxConnLifetime = postgresStoreConfig.MaxConnectionLifetime.Duration
	pgxConfig.MaxConnLifetimeJitter = postgresStoreConfig.MaxConnectionLifetimeJitter.Duration
	pgxConfig.HealthCheckPeriod = postgresStoreConfig.HealthCheckInterval.Duration
	pgxConfig.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTracerAttributes(spanTypeSQLAttribute))
	for _, customizer := range customizers {
		customizer.Customize(pgxConfig)
	}

	return pgxpool.NewWithConfig(context.Background(), pgxConfig)
}
