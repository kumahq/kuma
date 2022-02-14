package postgres

import (
	"database/sql"

	"github.com/pkg/errors"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
)

func ConnectToDb(cfg config.PostgresStoreConfig) (*sql.DB, error) {
	connStr, err := cfg.ConnectionString()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", connStr)
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
