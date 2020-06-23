package postgres

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	config "github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
)

func ConnectToDb(cfg config.PostgresStoreConfig) (*sql.DB, error) {
	mode, err := postgresMode(cfg.TLS.Mode)
	if err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s connect_timeout=%d sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName, cfg.ConnectionTimeout, mode, cfg.TLS.CertPath, cfg.TLS.KeyPath, cfg.TLS.CAPath)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create connection to DB")
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)

	// check connection to DB, Open() does not check it.
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "cannot connect to DB")
	}

	return db, nil
}

func postgresMode(mode config.TLSMode) (string, error) {
	switch mode {
	case config.Disable:
		return "disable", nil
	case config.VerifyNone:
		return "require", nil
	case config.VerifyCa:
		return "verify-ca", nil
	case config.VerifyFull:
		return "verify-full", nil
	default:
		return "", errors.Errorf("could not translate mode %q to postgres mode", mode)
	}
}
