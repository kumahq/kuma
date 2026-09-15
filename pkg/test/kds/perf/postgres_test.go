package perf

import (
	"context"
	"database/sql"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"

	pg_config "github.com/kumahq/kuma/v3/pkg/config/plugins/resources/postgres"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	postgres_store "github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres"
	pgx_config "github.com/kumahq/kuma/v3/pkg/plugins/resources/postgres/config"
)

// newPostgresStore points the harness at a real Postgres so that store latency
// and connection pool behavior are part of the measurement, which an in-memory
// store cannot model. KDS_POSTGRES takes a DSN, for example
// postgres://kuma:kuma@localhost:15432/kuma
func newPostgresStore(t *testing.T, dsn string) store.ResourceStore {
	t.Helper()

	cfg := postgresConfigFromDSN(t, dsn)

	migrated, err := postgres_store.IsDbMigrated(cfg)
	if err != nil {
		t.Fatalf("check postgres migration: %v", err)
	}
	if !migrated {
		if _, err := postgres_store.MigrateDb(cfg); err != nil {
			t.Fatalf("migrate postgres: %v", err)
		}
	}

	truncateResources(t, cfg)

	metrics, err := core_metrics.NewMetrics("perf-postgres")
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}

	pgStore, err := postgres_store.NewPgxStore(metrics, cfg, pgx_config.NoopPgxConfigCustomizationFn, resource_labels.ControlPlane{})
	if err != nil {
		t.Fatalf("postgres store: %v", err)
	}
	t.Cleanup(func() {
		if closer, ok := pgStore.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	})
	return pgStore
}

// truncateResources gives each run an empty store, so that repeated runs
// against the same database measure the same workload.
func truncateResources(t *testing.T, cfg pg_config.PostgresStoreConfig) {
	t.Helper()
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     cfg.DbName,
		RawQuery: "sslmode=disable",
	}
	db, err := sql.Open("pgx", dsn.String())
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(), "TRUNCATE TABLE resources"); err != nil {
		t.Fatalf("truncate resources: %v", err)
	}
}

func postgresConfigFromDSN(t *testing.T, dsn string) pg_config.PostgresStoreConfig {
	t.Helper()

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse KDS_POSTGRES: %v", err)
	}
	port := 5432
	if p := u.Port(); p != "" {
		if port, err = strconv.Atoi(p); err != nil {
			t.Fatalf("parse port: %v", err)
		}
	}
	password, _ := u.User.Password()

	cfgPtr := pg_config.DefaultPostgresStoreConfig()
	cfg := *cfgPtr
	cfg.Host = u.Hostname()
	cfg.Port = port
	cfg.User = u.User.Username()
	cfg.Password = password
	cfg.DbName = strings.TrimPrefix(u.Path, "/")
	cfg.TLS.Mode = pg_config.Disable
	cfg.DriverName = pg_config.DriverNamePgx
	if err := cfg.Validate(); err != nil {
		t.Fatalf("postgres config: %v", err)
	}
	return cfg
}
