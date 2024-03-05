package postgres

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
	postgres_cfg "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
	"github.com/kumahq/kuma/pkg/plugins/resources/postgres/migrations"
)

func MigrateDb(cfg postgres_cfg.PostgresStoreConfig) (core_plugins.DbVersion, error) {
	m, err := newMigrate(cfg)
	if err != nil {
		return 0, err
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			ver, _, err := m.Version()
			if err != nil {
				return 0, err
			}
			return ver, core_plugins.AlreadyMigrated
		}
		if strings.Contains(err.Error(), "file does not exist") {
			dbVer, _, err := m.Version()
			if err != nil {
				return 0, err
			}
			appVer, err := newestMigration()
			if err != nil {
				return 0, err
			}
			return 0, errors.Errorf("DB is migrated to newer version than Kuma. DB migration version %d. Kuma migration version %d. Run newer version of Kuma", dbVer, appVer)
		}
		return 0, errors.Wrap(err, "error while executing up migration")
	}
	ver, _, err := m.Version()
	if err != nil {
		return 0, err
	}
	return ver, nil
}

func newMigrate(cfg postgres_cfg.PostgresStoreConfig) (*migrate.Migrate, error) {
	db, err := common_postgres.ConnectToDb(cfg)
	if err != nil {
		return nil, err
	}
	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	sourceDriver, err := httpfs.New(http.FS(migrations.MigrationFS()), ".")
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithInstance("httpfs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func IsDbMigrated(cfg postgres_cfg.PostgresStoreConfig) (bool, error) {
	m, err := newMigrate(cfg)
	if err != nil {
		return false, err
	}
	dbVer, _, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return false, nil
		}
		return false, err
	}

	fileVer, err := newestMigration()
	if err != nil {
		return false, err
	}

	return dbVer == fileVer, nil
}

func newestMigration() (core_plugins.DbVersion, error) {
	files, err := data.ReadFiles(migrations.MigrationFS())
	if err != nil {
		return 0, err
	}
	latest := 0
	for _, file := range files {
		parts := strings.Split(file.Name, "_")
		ver, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		if ver > latest {
			latest = ver
		}
	}
	return core_plugins.DbVersion(latest), nil
}
