package postgres

import (
	"fmt"
	"strings"

	"github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
	"github.com/Kong/kuma/pkg/core"
)

func CreateRandomDb(cfg postgres.PostgresStoreConfig) (string, error) {
	db, err := ConnectToDb(cfg)
	if err != nil {
		return "", err
	}
	dbName := fmt.Sprintf("kuma_%s", strings.ReplaceAll(core.NewUUID(), "-", ""))
	statement := fmt.Sprintf("CREATE DATABASE %s", dbName)
	if _, err = db.Exec(statement); err != nil {
		return "", err
	}
	if err = db.Close(); err != nil {
		return "", err
	}
	return dbName, err
}
